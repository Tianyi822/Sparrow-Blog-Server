package aof

import (
	"bufio"
	"fmt"
	"h2blog/pkg/file"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileOp 文件操作核心结构体，负责管理文件的生命周期、缓冲写入、自动分割和压缩等操作
type FileOp struct {
	rwMu           sync.RWMutex   // 读写锁保护并发访问
	file           *os.File       // 底层文件句柄（nil表示未打开）
	writer         *bufio.Writer  // 缓冲写入器（默认4KB缓冲区）
	scanner        *bufio.Scanner // 缓冲读取器
	isOpen         bool           // 文件可写状态标识
	needCompress   bool           // 分割后压缩开关
	maxSize        int            // 文件分割阈值（单位MB）
	path           string         // 当前活跃文件绝对路径
	filePrefixName string         // 文件名主部（不含扩展名）
	fileSuffixName string         // 文件扩展名
}

// FoConfig 定义文件滚动切割和压缩的配置参数
type FoConfig struct {
	NeedCompress bool   // 是否启用分割后压缩
	MaxSize      int    // 单个文件最大尺寸（单位MB）
	Path         string // 文件完整路径
}

// CreateFileOp 初始化文件写入器实例
// 参数:
//
//	config *FoConfig - 文件配置参数，包含路径、大小限制和压缩设置
//
// 返回值:
//
//	*FileOp - 返回初始化完成但尚未打开的文件操作对象
//
// 注意:
//  1. 实际文件操作会在第一次Write调用时延迟打开
//  2. 文件路径需包含文件名和扩展名(如：app.log)
//  3. 文件目录不存在时会自动创建
func CreateFileOp(config FoConfig) FileOp {
	baseName := filepath.Base(config.Path)
	ext := filepath.Ext(baseName)
	prefix := strings.TrimSuffix(baseName, ext)

	return FileOp{
		filePrefixName: prefix,
		fileSuffixName: strings.TrimPrefix(ext, "."), // 移除扩展名前的点
		path:           config.Path,
		needCompress:   config.NeedCompress,
		isOpen:         false,
		maxSize:        config.MaxSize,
	}
}

// ready 准备文件进行写入操作
// 1. 检查文件是否已打开，未打开时根据路径是否存在决定打开或创建文件
// 2. 初始化缓冲写入器
// 返回错误包含：
//   - 文件打开失败
//   - 文件创建失败
//   - 目录创建失败
func (fop *FileOp) ready() (err error) {
	if fop.file == nil {
		if file.IsExist(fop.path) {
			fop.file, err = file.MustOpenFile(fop.path)
			if err != nil {
				return err
			}
		} else {
			fop.file, err = file.CreateFile(fop.path)
			if err != nil {
				return err
			}
		}
		fop.writer = bufio.NewWriter(fop.file)
		fop.scanner = bufio.NewScanner(fop.file)
	}
	fop.isOpen = true
	return nil
}

// Close 关闭文件
func (fop *FileOp) Close() error {
	if !fop.isOpen {
		return nil
	}

	// 先处理 bufio.Writer
	if fop.writer != nil {
		if err := fop.writer.Flush(); err != nil {
			return fmt.Errorf("刷新失败: %w", err)
		}
	}

	// 同步文件到磁盘
	if err := fop.file.Sync(); err != nil {
		return fmt.Errorf("同步失败: %w", err)
	}

	// 关闭文件句柄
	if err := fop.file.Close(); err != nil {
		return fmt.Errorf("关闭失败: %w", err)
	}

	fop.isOpen = false
	fop.writer = nil
	fop.file = nil

	return nil
}

// needSplit 检查文件是否需要分割
// 返回值：是否需要分割
func (fop *FileOp) needSplit() bool {
	// 判断是否需要进行分片
	if fop.maxSize <= 0 {
		return false
	}

	// 判断文件大小是否超过最大值
	fileInfo, err := fop.file.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() > int64(fop.maxSize*1024*1024)
}

// Write 写入数据到文件
// 实现自动文件分割和压缩逻辑
// context: 要写入的字节数据
func (fop *FileOp) Write(context []byte) error {
	fop.rwMu.Lock()
	defer fop.rwMu.Unlock()

	if !fop.isOpen {
		if err := fop.ready(); err != nil {
			return err
		}
	}

	// 判断是否需要进行分片
	if fop.needSplit() {
		// 关闭文件
		if err := fop.Close(); err != nil {
			return err
		}

		// 修改文件名
		newFileName := fmt.Sprintf("%v_%v.%v", fop.filePrefixName, strconv.FormatInt(time.Now().Unix(), 10), fop.fileSuffixName)
		destPath := filepath.Join(filepath.Dir(fop.path), newFileName)
		if err := os.Rename(fop.path, destPath); err != nil {
			return err
		}

		// 压缩文件
		if fop.needCompress {
			if err := file.CompressFileToTarGz(destPath); err != nil {
				return err
			}

			// 删除原文件
			if err := os.RemoveAll(destPath); err != nil {
				return err
			}
		}

		// 重新打开文件
		if err := fop.ready(); err != nil {
			return err
		}
	}

	// 写入数据
	buf := append(context, '\n')
	_, err := fop.writer.Write(buf)
	if err != nil {
		return err
	}
	// 数据落盘
	err = fop.writer.Flush()
	if err != nil {
		return err
	}
	return err
}

// GetScanner 获取文件扫描器
func (fop *FileOp) GetScanner() (*bufio.Scanner, error) {
	if !fop.isOpen {
		if err := fop.ready(); err != nil {
			return nil, err
		}
	}

	return fop.scanner, nil
}
