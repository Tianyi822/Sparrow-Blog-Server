package libs

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FileOp 文件操作核心结构体
// 负责管理文件的生命周期、缓冲写入、自动分割和压缩等操作
type FileOp struct {
	file           *os.File      // 底层文件句柄（nil表示未打开）
	writer         *bufio.Writer // 缓冲写入器（默认缓冲区大小4096字节）
	isOpen         bool          // 文件可写状态标识（true时允许写入操作）
	needCompress   bool          // 分割后压缩开关（true启用tar.gz压缩）
	maxSize        int           // 文件分割阈值（单位MB，0表示不分割）
	path           string        // 当前活跃文件绝对路径（包含文件名）
	filePrefixName string        // 文件名主部（不含扩展名，用于构建分割文件名）
	fileSuffixName string        // 文件扩展名（不含点，用于构建分割文件名）
}

// FWConfig 文件写入器配置结构
// 用于定义文件滚动切割和压缩的配置参数
type FWConfig struct {
	NeedCompress bool   // 是否启用压缩 (true: 文件分割后进行tar.gz压缩)
	MaxSize      int    // 单个文件最大尺寸 (单位MB，建议值1-1024)
	Path         string // 文件完整路径 (示例: "/var/log/app/error.log")
}

// CreateFileWriter 初始化文件写入器实例
// 参数:
//
//	config *FWConfig - 文件配置参数，包含路径、大小限制和压缩设置
//
// 返回值:
//
//	*FileOp - 返回初始化完成但尚未打开的文件操作对象
//
// 注意:
//  1. 实际文件操作会在第一次Write调用时延迟打开
//  2. 文件路径需包含文件名和扩展名(如：app.log)
//  3. 文件目录不存在时会自动创建
func CreateFileWriter(config *FWConfig) *FileOp {
	baseName := filepath.Base(config.Path)
	ext := filepath.Ext(baseName)
	prefix := strings.TrimSuffix(baseName, ext)

	return &FileOp{
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
func (fw *FileOp) ready() (err error) {
	if fw.file == nil {
		if isExist(fw.path) {
			fw.file, err = mustOpenFile(fw.path)
			if err != nil {
				return err
			}
		} else {
			fw.file, err = createFile(fw.path)
			if err != nil {
				return err
			}
		}
		fw.writer = bufio.NewWriter(fw.file)
	}
	fw.isOpen = true
	return nil
}

// Close 关闭文件
func (fw *FileOp) Close() error {
	fw.isOpen = false

	// 将缓存中的数据落盘
	err := fw.writer.Flush()
	if err != nil {
		return err
	}

	err = fw.file.Close()
	if err != nil {
		return err
	}

	fw.file = nil
	fw.writer = nil
	return err
}

// needSplit 检查文件是否需要分割
// 返回值：(需要分割, 错误)
// 改进点：
// 1. 增加文件状态获取错误的详细上下文
// 2. 明确单位转换逻辑
func (fw *FileOp) needSplit() (bool, error) {
	// 判断是否需要进行分片
	if fw.maxSize <= 0 {
		return false, nil
	}

	// 判断文件大小是否超过最大值
	fileInfo, err := fw.file.Stat()
	if err != nil {
		return false, fmt.Errorf("获取文件状态失败: %w", err)
	}

	// 计算最大字节数（MB转字节）
	maxBytes := int64(fw.maxSize) * 1024 * 1024
	return fileInfo.Size() > maxBytes, nil
}

// Write 写入数据到文件
// 实现自动文件分割和压缩逻辑
// context: 要写入的字节数据
func (fw *FileOp) Write(context []byte) (err error) {
	// 错误处理包装
	defer func() {
		if err != nil {
			err = fmt.Errorf("文件写入失败: %w", err)
		}
	}()
	if !fw.isOpen {
		err := fw.ready()
		if err != nil {
			return err
		}
	}

	// 检查是否需要分割
	needSplit, err := fw.needSplit()
	if err != nil {
		return err
	}
	if needSplit {
		// 关闭文件
		err := fw.Close()
		if err != nil {
			return err
		}

		// 修改文件名
		newFileName := fmt.Sprintf("%v_%v.%v", fw.filePrefixName, strconv.FormatInt(time.Now().Unix(), 10), fw.fileSuffixName)
		destPath := filepath.Join(filepath.Dir(fw.path), newFileName)
		err = os.Rename(fw.path, destPath)
		if err != nil {
			return err
		}

		// 压缩文件
		if fw.needCompress {
			_, err = compressFileToTarGz(destPath)
			if err != nil {
				return err
			}
			// 删除原文件
			err = remove(destPath)
			if err != nil {
				return err
			}
		}

		// 重新打开文件
		err = fw.ready()
		if err != nil {
			return err
		}
	}

	// 写入数据
	buf := append(context, '\n')
	_, err = fw.writer.Write(buf)
	if err != nil {
		return err
	}
	// 数据落盘
	err = fw.writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新缓冲区失败: %w", err)
	}
	return err
}

// isExist 判断路径是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// mustOpenFile 打开已存在的文件
// 安全改进：
// 1. 限制文件权限为所有者读写 (0600)
// 2. 增加错误上下文信息
// 3. 添加文件状态验证
func mustOpenFile(realPath string) (*os.File, error) {
	// 验证文件是否为普通文件
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("文件状态获取失败 %s: %w", realPath, err)
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("路径不是普通文件 %s", realPath)
	}

	file, err := os.OpenFile(realPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败 %s: %w", realPath, err)
	}
	return file, nil
}

// createFile 创建新文件并初始化
// 改进点：
// 1. 使用更安全的文件权限 (0600)
// 2. 优化目录创建错误处理
// 3. 增加创建过程的错误上下文
// 4. 添加文件模式验证
func createFile(path string) (*os.File, error) {
	// 验证文件名格式
	if !strings.Contains(filepath.Base(path), ".") {
		return nil, fmt.Errorf("文件名缺少扩展名 %s", path)
	}
	dir := filepath.Dir(path)
	exist := isExist(dir)
	if !exist {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// 文件路径
	exist = isExist(path)
	if !exist {
		_, err := os.Create(path)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600) // 更安全的权限设置
	if err != nil {
		return nil, err
	}

	return file, nil
}

// remove 删除文件
func remove(path string) error {
	return os.RemoveAll(path)
}

// compressFileToTarGz 将文件压缩为tar.gz格式
// src: 源文件路径
// 返回：压缩成功后的文件路径，错误信息
func compressFileToTarGz(src string) (string, error) {
	dir := filepath.Dir(src)
	filePrefixName := strings.Split(filepath.Base(src), ".")[0]
	dst := filepath.Join(dir, filePrefixName+".tar.gz")

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("创建压缩文件失败: %w", err)
	}
	defer destFile.Close()

	// 创建Gzip压缩写入器
	gzw := gzip.NewWriter(destFile)
	defer gzw.Close()

	// 创建Tar写入器
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("创建文件头失败: %w", err)
	}
	defer srcFile.Close()

	// 获取源文件的信息
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return "", fmt.Errorf("创建文件头失败: %w", err)
	}

	// 构建文件头信息
	header, err := tar.FileInfoHeader(srcFileInfo, "")
	if err != nil {
		return "", fmt.Errorf("构建文件头失败: %w", err)
	}

	// 更新文件头中的路径信息
	header.Name = filepath.Base(src)

	// 写入文件头
	if err := tw.WriteHeader(header); err != nil {
		return "", fmt.Errorf("写入文件头失败: %w", err)
	}

	// 将源文件内容复制到Tar包中
	if _, err = io.Copy(tw, srcFile); err != nil {
		return "", fmt.Errorf("写入压缩内容失败: %w", err)
	}

	return dst, nil
}
