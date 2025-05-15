package aof

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileOp 核心文件操作结构，负责管理文件生命周期，
// 缓冲写入，自动分割和压缩操作。
// 它提供线程安全的文件操作，具有以下特性：
// - 使用缓冲写入提高性能
// - 在达到大小限制时自动轮转文件
// - 可选的轮转文件压缩
// - 并发访问保护
type FileOp struct {
	rwMu           sync.RWMutex  // 用于保护并发访问的读写锁
	file           *os.File      // 底层文件句柄（未打开时为 nil）
	writer         *bufio.Writer // 缓冲写入器（32KB 缓冲区）
	isOpen         bool          // 文件可写状态标志
	needCompress   bool          // 是否在轮转后压缩文件
	maxSize        uint16        // 轮转前的最大文件大小（单位：MB）
	path           string        // 当前活动文件的绝对路径
	filePrefixName string        // 不带扩展名的基本文件名
	fileSuffixName string        // 不带点的文件扩展名
}

// CreateFileOp 使用给定的配置初始化一个新的 FileOp 实例。
// 它验证配置并设置文件操作结构。
// 在第一次写入操作之前，实际文件不会被打开。
//
// 返回值：
//   - *FileOp: 初始化完成的 FileOp 实例，包含文件路径、前缀、后缀、压缩需求等信息。
//   - error: 如果文件路径为空或配置无效，则返回错误信息。
func CreateFileOp() (*FileOp, error) {
	// 验证必需的配置，确保文件路径不为空
	if config.Cache.Aof.Path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// 将文件路径拆分为文件名和扩展名，用于后续初始化
	baseName := filepath.Base(config.Cache.Aof.Path)
	ext := filepath.Ext(baseName)
	prefix := strings.TrimSuffix(baseName, ext)

	// 返回初始化的 FileOp 实例
	return &FileOp{
		filePrefixName: prefix,
		fileSuffixName: strings.TrimPrefix(ext, "."),
		path:           config.Cache.Aof.Path,
		needCompress:   config.Cache.Aof.Compress,
		maxSize:        config.Cache.Aof.MaxSize,
	}, nil
}

// ready 为写入操作准备文件。
// 如果需要，它会创建目录，打开或创建文件，
// 并初始化缓冲写入器。
func (fop *FileOp) ready() error {
	if fop.file != nil {
		return nil // 文件已经打开
	}

	// 确保目录存在
	dir := filepath.Dir(fop.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 打开现有文件或创建新文件
	var err error
	if filetool.IsExist(fop.path) {
		fop.file, err = filetool.MustOpenFile(fop.path)
	} else {
		fop.file, err = filetool.CreateFile(fop.path)
	}
	if err != nil {
		return fmt.Errorf("failed to open/create file: %w", err)
	}

	// 使用 32KB 缓冲区初始化缓冲写入器
	fop.writer = bufio.NewWriterSize(fop.file, 32*1024)
	fop.isOpen = true
	return nil
}

// Close 刷新所有缓冲数据并关闭文件。
// 多次调用 Close 是安全的。
func (fop *FileOp) Close() error {
	if !fop.isOpen {
		return nil
	}

	// 刷新缓冲数据
	if fop.writer != nil {
		if err := fop.writer.Flush(); err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
	}

	// 确保数据写入磁盘
	if err := fop.file.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// 关闭文件句柄
	if err := fop.file.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// 重置文件操作状态
	fop.isOpen = false
	fop.writer = nil
	fop.file = nil

	return nil
}

// needSplit 检查当前文件大小是否超过配置的最大大小。
// 如果应该轮转文件，则返回 true。
func (fop *FileOp) needSplit() bool {
	// 如果禁用了轮转则跳过
	if fop.maxSize <= 0 {
		return false
	}

	// 获取当前文件大小
	fileInfo, err := fop.file.Stat()
	if err != nil {
		return false
	}

	// 与最大大小比较（将 MB 转换为字节）
	return uint64(fileInfo.Size()) > uint64(fop.maxSize)*1024*1024
}

// Write 在文件中追加数据，并带有换行符。
// 如果达到大小限制，它会自动处理文件轮转。
// 该操作是线程安全的，并且使用缓冲提高性能。
func (fop *FileOp) Write(context []byte) error {
	if fop == nil {
		return fmt.Errorf("FileOp is nil")
	}
	if len(context) == 0 {
		return nil
	}

	fop.rwMu.Lock()
	defer fop.rwMu.Unlock()

	// 确保文件准备好写入
	if !fop.isOpen {
		if err := fop.ready(); err != nil {
			return err
		}
	}

	// 写入前检查是否需要轮转
	if fop.needSplit() {
		if err := fop.checkAndRotate(); err != nil {
			return fmt.Errorf("rotation failed: %w", err)
		}
	}

	// 准备带有换行符的数据
	buf := make([]byte, len(context)+1)
	copy(buf, context)
	buf[len(context)] = '\n'

	// 写入数据
	if _, err := fop.writer.Write(buf); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// 写入后始终刷新以确保数据被写入
	if err := fop.writer.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}

	return nil
}

// checkAndRotate 处理达到大小限制时的文件轮转逻辑。
func (fop *FileOp) checkAndRotate() error {
	if !fop.needSplit() {
		return nil
	}

	// 关闭前刷新缓冲数据
	if err := fop.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer before rotation: %w", err)
	}

	// 同步到磁盘
	if err := fop.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file before rotation: %w", err)
	}

	// 关闭当前文件
	if err := fop.Close(); err != nil {
		return err
	}

	// 生成轮转后的文件名
	rotatedName := fmt.Sprintf("%v.aof", fop.filePrefixName)
	destPath := filepath.Join(filepath.Dir(fop.path), rotatedName)

	// 重命名当前文件
	if err := os.Rename(fop.path, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// 如果启用，则处理压缩
	if fop.needCompress {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		compressedPath := fmt.Sprintf("%v_%v.aof.tar.gz", fop.filePrefixName, timestamp)
		compressedPath = filepath.Join(filepath.Dir(fop.path), compressedPath)

		// 压缩文件
		if err := filetool.CompressFileToTarGz(destPath, compressedPath); err != nil {
			// 如果压缩失败，尝试恢复原始文件
			_ = os.Rename(destPath, fop.path)
			return fmt.Errorf("compression failed: %w", err)
		}

		// 在删除原始文件前验证压缩文件是否存在
		if !filetool.IsExist(compressedPath) {
			_ = os.Rename(destPath, fop.path)
			return fmt.Errorf("compressed file not found after compression")
		}

		// 仅在成功压缩后删除原始文件
		if err := os.RemoveAll(destPath); err != nil {
			logger.Warn("failed to remove rotated file after compression: %v", err)
		}
	}

	// 创建新的写入文件
	return fop.ready()
}
