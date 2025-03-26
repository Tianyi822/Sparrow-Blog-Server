package aof

import (
	"bufio"
	"context"
	"fmt"
	"h2blog_server/cache/common"
	"h2blog_server/pkg/fileTool"
	"h2blog_server/pkg/logger"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// aof 包实现了一个追加式文件（AOF）持久化机制，用于缓存系统。
// 它提供了线程安全的操作，用于存储和加载缓存命令，并支持
// 文件轮转、压缩和自动清理功能。

// Aof 表示用于数据持久化的追加式文件实现。
// 它提供了线程安全的操作，用于存储和加载缓存命令。
// 特性:
// - 使用互斥锁的线程安全操作
// - 基于大小的自动文件轮转
// - 支持压缩文件存储
// - 按时间顺序处理命令
// - 自动清理已处理的文件
type Aof struct {
	file *FileOp      // 处理文件底层操作，包括轮转和压缩
	mu   sync.RWMutex // 确保文件操作的线程安全访问
}

// NewAof 创建并返回一个新的 AOF 实例。
// 它使用 cache-config.yaml 中的配置初始化 AOF 文件。
// 配置包括:
// - AOF 存储的文件路径
// - 轮转前的最大文件大小
// - 被轮转文件的压缩设置
//
// 如果文件创建失败则会触发 panic，因为这对数据持久化至关重要。
// 这是启动过程中的关键操作 - 如果失败，应用程序将无法
// 保证数据持久性，不应继续运行。
func NewAof() *Aof {
	fileOp, err := CreateFileOp()
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}

	return &Aof{
		file: fileOp,
	}
}

// LoadFile 读取并处理所有 AOF 文件（包括压缩文件）。
// 处理过程:
// 1. 列出目录中的所有 AOF 文件（包括 .aof 和 .aof.tar.gz）
// 2. 按时间戳排序文件以维持操作顺序
// 3. 根据需要创建解压缩的临时目录
// 4. 按时间顺序处理每个文件
// 5. 成功加载后清理已处理的文件
// 6. 删除临时目录
//
// 线程安全:
// - 在加载过程中使用互斥锁防止并发访问
// - 支持上下文取消以实现优雅关闭
//
// 错误处理:
// - 如果文件列举失败则返回错误
// - 如果任何文件处理失败则返回错误
// - 对于非关键性失败（如文件清理）记录警告
//
// 返回:
// - [][]string: 命令参数的切片，每个内部切片表示一个命令
// - error: 处理过程中遇到的任何错误
func (aof *Aof) LoadFile(ctx context.Context) ([][]string, error) {
	select {
	case <-ctx.Done():
		logger.Warn("AOF file loading cancelled")
		return nil, ctx.Err()
	default:
		aof.mu.Lock()
		defer aof.mu.Unlock()

		dir := filepath.Dir(aof.file.path)
		prefix := aof.file.filePrefixName

		// 首先创建临时目录
		tempDir := filepath.Join(dir, "temp_aof")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				logger.Warn("failed to remove temp directory %s: %v", tempDir, err)
			}
		}()

		var files []string
		var err error

		// 首先，获取所有压缩文件
		if aof.file.needCompress {
			files, err = filepath.Glob(filepath.Join(dir, prefix+"_*.aof.tar.gz"))
			if err != nil {
				return nil, fmt.Errorf("failed to list compressed AOF files: %w", err)
			}
			// 按时间戳排序压缩文件
			sort.Slice(files, func(i, j int) bool {
				tsI := extractTimestamp(files[i])
				tsJ := extractTimestamp(files[j])
				return tsI < tsJ
			})
		}

		// 然后检查当前的 AOF 文件
		currentFile := filepath.Join(dir, prefix+".aof")
		if fileTool.IsExist(currentFile) {
			files = append(files, currentFile)
		}

		var allCommands [][]string
		logger.Info("Processing %d files", len(files))

		// 处理每个文件
		for i, f := range files {
			logger.Info("Processing file %d/%d: %s", i+1, len(files), f)
			commands, err := processFile(f, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to process file %s: %w", f, err)
			}
			logger.Info("File %s: loaded %d commands", f, len(commands))
			allCommands = append(allCommands, commands...)
		}

		// 清理已处理的文件
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				logger.Warn("failed to remove processed AOF file %s: %v", f, err)
			}
		}

		// 创建新的 AOF 文件
		if err := aof.file.ready(); err != nil {
			return nil, fmt.Errorf("failed to create new AOF file: %w", err)
		}

		logger.Info("AOF files loading completed: processed %d files, %d commands",
			len(files), len(allCommands))
		return allCommands, nil
	}
}

// Store 将命令写入 AOF 文件。
// 格式: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
//
// 支持的命令:
// - SET: 需要 4 个参数（键、值、类型、过期时间）
// - DELETE: 需要 1 个参数（键）
// - INCR: 需要 2 个参数（键、类型）
// - CLEANUP: 不需要参数
//
// 线程安全:
// - 使用互斥锁防止并发写入
// - 支持上下文取消
//
// 参数:
// - ctx: 用于取消的上下文
// - cmd: 命令类型（SET, DELETE, INCR, CLEANUP）
// - args: 命令参数（根据命令类型不同而变化）
//
// 如果出现以下情况则返回错误:
// - 上下文被取消
// - 参数数量无效
// - 不支持的命令类型
// - 写入操作失败
func (aof *Aof) Store(ctx context.Context, cmd string, args ...string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return aof.storeCommand(cmd, args...)
	}
}

// Close 安全地关闭 AOF 文件并将缓冲数据刷新到磁盘。
// 它确保在关闭前所有数据都被正确持久化。
//
// 线程安全:
// - 使用互斥锁防止关闭过程中的并发操作
//
// 返回:
// - error: 关闭操作过程中遇到的任何错误
func (aof *Aof) Close() error {
	if aof == nil || aof.file == nil {
		return nil
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	// 关闭底层文件
	return aof.file.Close()
}

// 内部辅助函数

// storeCommand 处理实际的命令存储逻辑。
// 它验证命令参数并格式化命令字符串。
//
// 命令格式:
// - SET: SET;;key;;value;;type;;expiry
// - DELETE: DELETE;;key
// - INCR: INCR;;key;;type
// - CLEANUP: CLEANUP
//
// 验证:
// - 检查每种命令类型的参数数量
// - 确保所有必需的参数都存在
//
// 如果出现以下情况则返回错误:
// - 参数数量无效
// - 写入操作失败
func (aof *Aof) storeCommand(cmd string, args ...string) error {
	switch cmd {
	case common.SET:
		if len(args) != 4 {
			return fmt.Errorf("SET command requires 4 args (key=%s, value=%s, type=%s, expired=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), safeGet(args, 2), safeGet(args, 3), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s;;%s;;%s", cmd, args[0], args[1], args[2], args[3])))

	case common.DELETE:
		if len(args) != 1 {
			return fmt.Errorf("DELETE command requires 1 arg (key=%s), got %d",
				safeGet(args, 0), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s", cmd, args[0])))

	case common.INCR:
		if len(args) != 2 {
			return fmt.Errorf("INCR command requires 2 args (key=%s, type=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s", cmd, args[0], args[1])))

	case common.CLEANUP:
		if len(args) != 0 {
			return fmt.Errorf("CLEANUP command requires no args, got %d", len(args))
		}
		return aof.file.Write([]byte(cmd))

	default:
		return fmt.Errorf("unsupported command type: %s", cmd)
	}
}

// processFile 根据压缩设置处理 AOF 文件
func processFile(path string, tempDir string) ([][]string, error) {
	var file *os.File
	var err error

	if strings.HasSuffix(path, ".tar.gz") {
		// 处理压缩文件
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		decompressedPath := filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(path), ".tar.gz"))
		if err := fileTool.DecompressTarGz(path, decompressedPath); err != nil {
			return nil, fmt.Errorf("failed to decompress %s: %w", path, err)
		}

		file, err = os.Open(decompressedPath)
	} else {
		// 处理常规文件
		file, err = os.Open(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Warn("failed to close file %s: %v", path, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	return processAOFFile(scanner)
}

// extractTimestamp 从 AOF 文件名中提取时间戳。
// 预期的文件名格式: prefix_timestamp.aof.* 或 prefix_timestamp.aof.tar.gz
//
// 参数:
// - filename: 要解析的 AOF 文件名
//
// 返回:
// - int64: 从文件名中提取的 Unix 时间戳
// - 如果文件名格式无效则返回 0
func extractTimestamp(filename string) int64 {
	parts := strings.Split(filepath.Base(filename), "_")
	if len(parts) < 2 {
		return 0
	}
	ts := strings.Split(parts[1], ".")[0]
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return timestamp
}

// processAOFFile 处理 AOF 文件的内容并返回命令。
// 它逐行读取文件并根据格式解析每个命令。
//
// 命令验证:
// - 检查命令格式和参数数量
// - 删除所有字段的空格
// - 记录无效命令的警告但继续处理
//
// 错误处理:
// - 跳过无效命令并发出警告
// - 如果扫描器遇到读取错误则返回错误
//
// 返回:
// - [][]string: 有效命令的切片
// - error: 处理过程中遇到的任何错误
func processAOFFile(scanner *bufio.Scanner) ([][]string, error) {
	var commands [][]string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		command := strings.Split(scanner.Text(), ";;")

		switch strings.TrimSpace(command[0]) {
		case common.SET:
			if len(command) != 5 {
				logger.Warn("line %d: invalid SET command format, skipping: %v", lineNum, command)
				continue
			}
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
				strings.TrimSpace(command[2]), // value
				strings.TrimSpace(command[3]), // type
				strings.TrimSpace(command[4]), // expiry
			})
		case common.DELETE:
			if len(command) != 2 {
				logger.Warn("line %d: invalid DELETE command format, skipping: %v", lineNum, command)
				continue
			}
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
			})
		case common.INCR:
			if len(command) != 3 {
				logger.Warn("line %d: invalid INCR command format, skipping: %v", lineNum, command)
				continue
			}
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
				strings.TrimSpace(command[2]), // type
			})
		case common.CLEANUP:
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
			})
		default:
			logger.Warn("line %d: unknown command, skipping: %v", lineNum, command[0])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return commands, nil
}

// safeGet 安全地从切片中检索元素。
// 通过对无效索引返回 "<nil>" 来防止索引超出范围的 panic。
//
// 参数:
// - slice: 要访问的字符串切片
// - index: 要检索的索引
//
// 返回:
// - string: 索引处的元素，如果索引无效则返回 "<nil>"
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return "<nil>"
	}
	return slice[index]
}
