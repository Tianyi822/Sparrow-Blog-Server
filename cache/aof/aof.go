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
// 这是启动过程中的关键操作 - 如果失败，应用程序将无法保证数据持久性，不应继续运行。
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
		// 如果上下文被取消，记录警告日志并返回上下文错误。
		logger.Warn("AOF 文件加载已取消")
		return nil, ctx.Err()
	default:
		// 加锁以确保线程安全，防止并发访问。
		aof.mu.Lock()
		defer aof.mu.Unlock()

		// 获取 AOF 文件所在的目录和文件前缀名。
		dir := filepath.Dir(aof.file.path)
		prefix := aof.file.filePrefixName

		// 创建临时目录用于解压和处理文件。
		tempDir := filepath.Join(dir, "temp_aof")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("无法创建临时目录: %w", err)
		}

		// 确保函数结束时删除临时目录。
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				logger.Warn("无法删除临时目录 %s: %v", tempDir, err)
			}
		}()

		var files []string
		var err error

		// 如果启用了压缩功能，获取所有压缩文件并按时间戳排序。
		if aof.file.needCompress {
			files, err = filepath.Glob(filepath.Join(dir, prefix+"_*.aof.tar.gz"))
			if err != nil {
				return nil, fmt.Errorf("无法列出压缩的 AOF 文件: %w", err)
			}
			// 按文件名中的时间戳对压缩文件进行排序。
			sort.Slice(files, func(i, j int) bool {
				tsI := extractTimestamp(files[i])
				tsJ := extractTimestamp(files[j])
				return tsI < tsJ
			})
		}

		// 检查当前的 AOF 文件是否存在，并将其加入文件列表。
		currentFile := filepath.Join(dir, prefix+".aof")
		if fileTool.IsExist(currentFile) {
			files = append(files, currentFile)
		}

		var allCommands [][]string
		logger.Info("正在处理 %d 个文件", len(files))

		// 按顺序处理每个文件。
		for i, f := range files {
			logger.Info("正在处理文件 %d/%d: %s", i+1, len(files), f)
			commands, err := processFile(f, tempDir)
			if err != nil {
				return nil, fmt.Errorf("无法处理文件 %s: %w", f, err)
			}
			logger.Info("文件 %s: 已加载 %d 条命令", f, len(commands))
			allCommands = append(allCommands, commands...)
		}

		// 清理已处理的文件。
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				logger.Warn("无法删除已处理的 AOF 文件 %s: %v", f, err)
			}
		}

		// 创建新的 AOF 文件以准备后续写入。
		if err := aof.file.ready(); err != nil {
			return nil, fmt.Errorf("无法创建新的 AOF 文件: %w", err)
		}

		// 记录加载完成的日志。
		logger.Info("AOF 文件加载完成: 处理了 %d 个文件，%d 条命令",
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
//
// 返回值：
// - error: 如果上下文被取消、命令参数无效或写入文件失败，则返回相应的错误；否则返回 nil。
func (aof *Aof) Store(ctx context.Context, cmd string, args ...string) error {
	// 加锁以确保线程安全，防止并发写入问题。
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// 检查上下文是否已取消。
	select {
	case <-ctx.Done():
		// 如果上下文被取消，返回上下文的错误信息。
		return ctx.Err()
	default:
		// 调用内部方法 storeCommand 执行实际的命令存储逻辑。
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
//
// 参数：
// - cmd: 表示要执行的命令类型，例如 SET、DELETE、INCR 或 CLEANUP。
// - args: 表示命令的参数列表，具体内容取决于命令类型。
//
// 返回值：
// - error: 如果命令参数不合法或写入文件失败，则返回错误；否则返回 nil。
func (aof *Aof) storeCommand(cmd string, args ...string) error {
	switch cmd {
	case common.SET:
		// 检查 SET 命令的参数数量是否正确，SET 命令需要 4 个参数：key、value、type 和 expired。
		if len(args) != 4 {
			return fmt.Errorf("SET command requires 4 args (key=%s, value=%s, type=%s, expired=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), safeGet(args, 2), safeGet(args, 3), len(args))
		}
		// 将 SET 命令及其参数格式化为字符串并写入 AOF 文件。
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s;;%s;;%s", cmd, args[0], args[1], args[2], args[3])))

	case common.DELETE:
		// 检查 DELETE 命令的参数数量是否正确，DELETE 命令需要 1 个参数：key。
		if len(args) != 1 {
			return fmt.Errorf("DELETE command requires 1 arg (key=%s), got %d",
				safeGet(args, 0), len(args))
		}
		// 将 DELETE 命令及其参数格式化为字符串并写入 AOF 文件。
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s", cmd, args[0])))

	case common.INCR:
		// 检查 INCR 命令的参数数量是否正确，INCR 命令需要 2 个参数：key 和 type。
		if len(args) != 2 {
			return fmt.Errorf("INCR command requires 2 args (key=%s, type=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), len(args))
		}
		// 将 INCR 命令及其参数格式化为字符串并写入 AOF 文件。
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s", cmd, args[0], args[1])))

	case common.CLEANUP:
		// 检查 CLEANUP 命令的参数数量是否正确，CLEANUP 命令不需要任何参数。
		if len(args) != 0 {
			return fmt.Errorf("CLEANUP command requires no args, got %d", len(args))
		}
		// 将 CLEANUP 命令直接写入 AOF 文件。
		return aof.file.Write([]byte(cmd))

	default:
		// 如果命令类型不被支持，则返回错误。
		return fmt.Errorf("unsupported command type: %s", cmd)
	}
}

// processFile 处理指定路径的文件，支持常规文件和 .tar.gz 压缩文件。
// 参数:
// - path: 文件路径，可以是常规文件或 .tar.gz 压缩文件。
// - tempDir: 临时目录路径，用于解压 .tar.gz 文件。
//
// 返回值:
// - [][]string: 处理后的文件内容，以二维字符串切片形式返回。
// - error: 如果处理过程中发生错误，则返回具体的错误信息。
func processFile(path string, tempDir string) ([][]string, error) {
	var file *os.File
	var err error

	// 如果文件是 .tar.gz 压缩文件，则先解压到临时目录
	if strings.HasSuffix(path, ".tar.gz") {
		// 创建临时目录用于存放解压后的文件
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		// 解压 .tar.gz 文件到临时目录
		decompressedPath := filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(path), ".tar.gz"))
		if err := fileTool.DecompressTarGz(path, decompressedPath); err != nil {
			return nil, fmt.Errorf("failed to decompress %s: %w", path, err)
		}

		// 打开解压后的文件
		file, err = os.Open(decompressedPath)
	} else {
		// 如果是常规文件，直接打开文件
		file, err = os.Open(path)
	}

	// 检查文件打开是否成功
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	// 确保函数结束时关闭文件，并记录可能的关闭错误
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Warn("failed to close file %s: %v", path, closeErr)
		}
	}()

	// 使用 bufio.Scanner 逐行读取文件内容
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
//
// 函数逻辑:
// 1. 使用 filepath.Base 获取文件名部分，并通过 "_" 分割文件名。
// 2. 如果分割后的部分少于两部分，则认为文件名格式不符合预期，返回 0。
// 3. 从分割结果的第二部分中提取时间戳字符串，并尝试将其转换为 int64 类型。
func extractTimestamp(filename string) int64 {
	// 将文件名按 "_" 分割，获取文件名的组成部分。
	parts := strings.Split(filepath.Base(filename), "_")
	if len(parts) < 2 {
		// 如果分割结果少于两部分，说明文件名格式不正确，返回 0。
		return 0
	}

	// 从分割结果的第二部分中提取时间戳字符串，并去掉可能存在的文件扩展名。
	ts := strings.Split(parts[1], ".")[0]

	// 将时间戳字符串转换为 int64 类型，忽略转换错误。
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return timestamp
}

// processAOFFile 处理 AOF 文件的内容，逐行读取文件并根据格式解析每个命令。
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
//
// 函数逻辑:
// 1. 逐行读取文件内容，并根据 ";;" 分隔符将每行拆分为命令和参数。
// 2. 根据命令类型（SET、DELETE、INCR、CLEANUP）验证参数数量和格式。
// 3. 对于无效命令或格式错误的命令，记录警告日志并跳过该行。
// 4. 如果扫描器在读取过程中遇到错误，则返回错误信息。
func processAOFFile(scanner *bufio.Scanner) ([][]string, error) {
	var commands [][]string
	lineNum := 0

	// 逐行扫描文件内容
	for scanner.Scan() {
		lineNum++
		command := strings.Split(scanner.Text(), ";;")

		// 根据命令类型处理不同的逻辑
		switch strings.TrimSpace(command[0]) {
		case common.SET:
			// 验证 SET 命令的参数数量是否为 5
			if len(command) != 5 {
				logger.Warn("第 %d 行: SET 命令格式无效，跳过: %v", lineNum, command)
				continue
			}
			// 将 SET 命令及其参数添加到结果中
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
				strings.TrimSpace(command[2]), // value
				strings.TrimSpace(command[3]), // type
				strings.TrimSpace(command[4]), // expiry
			})
		case common.DELETE:
			// 验证 DELETE 命令的参数数量是否为 2
			if len(command) != 2 {
				logger.Warn("第 %d 行: DELETE 命令格式无效，跳过: %v", lineNum, command)
				continue
			}
			// 将 DELETE 命令及其参数添加到结果中
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
			})
		case common.INCR:
			// 验证 INCR 命令的参数数量是否为 3
			if len(command) != 3 {
				logger.Warn("第 %d 行: INCR 命令格式无效，跳过: %v", lineNum, command)
				continue
			}
			// 将 INCR 命令及其参数添加到结果中
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
				strings.TrimSpace(command[2]), // type
			})
		case common.CLEANUP:
			// CLEANUP 命令不需要参数，直接添加到结果中
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
			})
		default:
			// 记录未知命令的警告日志并跳过
			logger.Warn("第 %d 行: 未知命令，跳过: %v", lineNum, command[0])
		}
	}

	// 检查扫描器是否在读取过程中遇到错误
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("扫描文件时出错: %w", err)
	}

	// 返回解析后的命令和 nil 错误
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
		return "<nil>" // 索引无效时返回默认值 "<nil>"
	}
	return slice[index] // 返回索引处的有效元素
}
