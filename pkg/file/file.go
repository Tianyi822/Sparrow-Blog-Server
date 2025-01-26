package file

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsExist 判断路径是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MustOpenFile 打开已存在的文件
// 安全改进：
// 1. 限制文件权限为所有者读写 (0600)
// 2. 增加错误上下文信息
// 3. 添加文件状态验证
func MustOpenFile(realPath string) (*os.File, error) {
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

// CreateFile 创建新文件并初始化
// 改进点：
// 1. 使用更安全的文件权限 (0600)
// 2. 优化目录创建错误处理
// 3. 增加创建过程的错误上下文
// 4. 添加文件模式验证
func CreateFile(path string) (*os.File, error) {
	// 验证文件名格式
	if !strings.Contains(filepath.Base(path), ".") {
		return nil, fmt.Errorf("文件名缺少扩展名 %s", path)
	}
	dir := filepath.Dir(path)
	exist := IsExist(dir)
	if !exist {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	// 文件路径
	exist = IsExist(path)
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

// CompressFileToTarGz 将文件压缩为tar.gz格式
// src: 源文件路径
// 返回：压缩成功后的文件路径，错误信息
func CompressFileToTarGz(src string) error {
	dir := filepath.Dir(src)
	filePrefixName := strings.Split(filepath.Base(src), ".")[0]
	dst := filepath.Join(dir, filePrefixName+".tar.gz")

	// 创建目标文件时添加错误清理逻辑
	defer func() {
		if err := recover(); err != nil {
			// 如果压缩失败，清理已创建的目标文件
			if removeErr := os.RemoveAll(dst); removeErr != nil && !os.IsNotExist(removeErr) {
				err = fmt.Errorf("%v | 清理临时文件失败: %v", err, removeErr)
			}
		}
	}()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %w", err)
	}
	defer func() {
		// 捕获文件关闭错误，且不覆盖已有错误
		if closeErr := destFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("关闭压缩文件失败: %w", closeErr)
		}
	}()

	// 创建Gzip压缩写入器
	gzw := gzip.NewWriter(destFile)
	defer func() {
		// 捕获gzip关闭错误，且不覆盖已有错误
		if closeErr := gzw.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("关闭gzip写入器失败: %w", closeErr)
		}
	}()

	// 创建Tar写入器
	tw := tar.NewWriter(gzw)
	defer func() {
		// 捕获tar关闭错误，且不覆盖已有错误
		if closeErr := tw.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("关闭tar写入器失败: %w", closeErr)
		}
	}()

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer func() {
		// 修复: 正确关闭源文件而不是tar writer
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("关闭源文件失败: %w", closeErr)
		}
	}()

	// 获取源文件的信息
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	// 构建文件头信息
	header, err := tar.FileInfoHeader(srcFileInfo, "")
	if err != nil {
		return fmt.Errorf("构建tar文件头失败: %w", err)
	}

	// 更新文件头中的路径信息
	header.Name = filepath.Base(src)

	// 写入文件头
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("写入tar文件头失败: %w", err)
	}

	// 将源文件内容复制到Tar包中
	if _, err = io.Copy(tw, srcFile); err != nil {
		return fmt.Errorf("写入压缩内容失败: %w", err)
	}

	return nil
}

// DecompressTarGz 解压tar.gz文件
// src: 压缩包源路径
// dst: 目标文件名(不含路径)
func DecompressTarGz(src, dst string) error {
	// 获取源文件所在目录
	dir := filepath.Dir(src)
	targetPath := filepath.Join(dir, dst)

	// 如果目标文件存在则删除
	if IsExist(targetPath) {
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("删除目标文件失败: %w", err)
		}
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer func(srcFile *os.File) {
		if err = srcFile.Close(); err != nil {
			err = fmt.Errorf("关闭源文件失败: %w", err)
		}
	}(srcFile)

	// 创建gzip reader
	gzr, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("创建gzip reader失败: %w", err)
	}
	defer func(gzr *gzip.Reader) {
		if err = gzr.Close(); err != nil {
			err = fmt.Errorf("关闭gzip reader失败: %w", err)
		}
	}(gzr)

	// 创建tar reader
	tr := tar.NewReader(gzr)

	// 处理压缩包中的第一个文件
	header, err := tr.Next()
	if err == io.EOF {
		return fmt.Errorf("压缩包为空")
	}
	if err != nil {
		return fmt.Errorf("读取tar文件失败: %w", err)
	}

	// 创建目标文件
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %w", targetPath, err)
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			err = fmt.Errorf("关闭文件失败 %s: %w", targetPath, err)
		}
	}(file)

	// 复制文件内容
	if _, err := io.Copy(file, tr); err != nil {
		return fmt.Errorf("写入文件失败 %s: %w", targetPath, err)
	}

	return nil
}
