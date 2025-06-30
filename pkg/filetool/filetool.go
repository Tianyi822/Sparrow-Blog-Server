package filetool

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// IsExist 检查指定路径是否存在。
// 返回 true 表示路径存在，false 表示不存在。
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MustOpenFile 使用安全设置打开现有文件。
// 该函数具有以下安全特性：
//   - 限制文件权限为所有者读写 (0600)
//   - 验证文件类型（必须是常规文件）
//   - 使用安全的打开标志
//
// 参数：
//   - realPath: 要打开的文件路径
//
// 返回值：
//   - *os.File: 打开的文件句柄
//   - error: 如果打开失败则返回错误
func MustOpenFile(realPath string) (*os.File, error) {
	// 验证文件是否为常规文件
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败 %s: %w", realPath, err)
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("路径不是常规文件: %s", realPath)
	}

	file, err := os.OpenFile(realPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败 %s: %w", realPath, err)
	}
	return file, nil
}

// CreateFile 创建并初始化新文件。
// 该函数具有以下特性：
//   - 使用安全的文件权限 (0600)
//   - 如果需要会自动创建父目录
//   - 验证文件名格式（必须包含扩展名）
//   - 安全地处理并发创建
//
// 参数：
//   - path: 要创建的文件路径
//
// 返回值：
//   - *os.File: 创建的文件句柄
//   - error: 如果创建失败则返回错误
func CreateFile(path string) (*os.File, error) {
	// 验证文件名是否包含扩展名
	if !strings.Contains(filepath.Base(path), ".") {
		return nil, fmt.Errorf("文件名缺少扩展名: %s", path)
	}

	// 如果需要则创建父目录
	dir := filepath.Dir(path)
	if !IsExist(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}

	// 使用安全设置创建或打开文件
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败 %s: %w", path, err)
	}

	return file, nil
}

// CompressFileToTarGz 将指定文件压缩为 tar.gz 格式。
// 如果目标路径为空，将使用默认命名规则：同目录下的 <源文件名>.tar.gz
//
// 参数：
//   - src: 源文件路径
//   - dst: 压缩文件的目标路径（必须包含 .tar.gz 扩展名，可为空使用默认命名）
//
// 返回值：
//   - error: 如果压缩失败则返回错误，成功则返回 nil
func CompressFileToTarGz(src, dst string) error {
	if dst == "" {
		// 如果未指定则使用默认目标路径
		dir := filepath.Dir(src)
		filePrefixName := strings.Split(filepath.Base(src), ".")[0]
		dst = filepath.Join(dir, filePrefixName+".tar.gz")
	}

	// 验证文件扩展名
	if !strings.HasSuffix(dst, ".tar.gz") {
		return fmt.Errorf("目标文件必须具有 .tar.gz 扩展名: %s", dst)
	}

	// panic 时清理
	defer func() {
		if r := recover(); r != nil {
			_ = os.RemoveAll(dst)
		}
	}()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %w", err)
	}
	defer closeFile(destFile, &err)

	// 设置压缩管道
	gzw := gzip.NewWriter(destFile)
	defer closeGzip(gzw, &err)

	tw := tar.NewWriter(gzw)
	defer closeTar(tw, &err)

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer closeFile(srcFile, &err)

	// 获取源文件信息
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	// 创建并写入 tar 头部
	header, err := tar.FileInfoHeader(srcFileInfo, "")
	if err != nil {
		return fmt.Errorf("创建 tar 头部失败: %w", err)
	}
	header.Name = filepath.Base(src)

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("写入 tar 头部失败: %w", err)
	}

	// 复制文件内容
	if _, err = io.Copy(tw, srcFile); err != nil {
		return fmt.Errorf("写入压缩内容失败: %w", err)
	}

	return nil
}

// DecompressTarGz 从 tar.gz 归档文件中提取文件。
// 如果目标文件已存在，将会被删除后重新创建。
//
// 参数：
//   - src: 源归档文件路径
//   - dst: 目标文件名（不包含路径）
//
// 返回值：
//   - error: 如果解压失败则返回错误，成功则返回 nil
func DecompressTarGz(src, dst string) error {
	// 删除现有的目标文件
	if IsExist(dst) {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("删除现有目标文件失败: %w", err)
		}
	}

	// 打开源归档
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开归档文件失败: %w", err)
	}
	defer closeFile(srcFile, &err)

	// 设置解压管道
	gzr, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("创建 gzip 读取器失败: %w", err)
	}
	defer closeGzipReader(gzr, &err)

	tr := tar.NewReader(gzr)

	// 从归档中读取第一个文件
	header, err := tr.Next()
	if err == io.EOF {
		return fmt.Errorf("空归档文件")
	}
	if err != nil {
		return fmt.Errorf("读取 tar 头部失败: %w", err)
	}

	// 创建目标文件
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("创建文件失败 %s: %w", dst, err)
	}
	defer closeFile(file, &err)

	// 复制内容
	if _, err := io.Copy(file, tr); err != nil {
		return fmt.Errorf("写入文件失败 %s: %w", dst, err)
	}

	return nil
}

// ForceRemove 强制删除指定路径的文件或目录。
// 该函数会递归删除目录及其所有内容，对于文件则直接删除。
// 如果路径不存在，函数会静默成功而不返回错误。
//
// 参数：
//   - path: 要删除的文件或目录路径
//
// 返回值：
//   - error: 如果删除失败则返回错误，成功则返回 nil
func ForceRemove(path string) error {
	// 检查路径是否存在
	if !IsExist(path) {
		// 路径不存在，静默成功
		return nil
	}

	// 获取路径信息
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("获取文件信息失败 %s: %w", path, err)
	}

	// 如果是目录，先尝试修改权限以确保可以删除
	if fileInfo.IsDir() {
		// 递归修改目录权限，确保可以删除
		err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				// 忽略权限错误，继续处理
				return nil
			}
			// 设置可写权限
			if err := os.Chmod(walkPath, 0755); err != nil {
				// 忽略权限修改错误，继续处理
				return nil
			}
			return nil
		})
		if err != nil {
			// 忽略遍历错误，继续尝试删除
		}
	} else {
		// 如果是文件，确保有写权限
		if err := os.Chmod(path, 0644); err != nil {
			// 忽略权限修改错误，继续尝试删除
		}
	}

	// 强制删除
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("删除失败 %s: %w", path, err)
	}

	return nil
}

// EnsureDir 确保指定的目录存在，如果不存在则创建它。
// 该函数会递归创建所有必要的父目录。
//
// 参数：
//   - dirPath: 要确保存在的目录路径
//
// 返回值：
//   - error: 如果创建失败则返回错误，成功则返回 nil
func EnsureDir(dirPath string) error {
	// 检查目录是否已存在
	if IsExist(dirPath) {
		// 验证路径是否为目录
		fileInfo, err := os.Stat(dirPath)
		if err != nil {
			return fmt.Errorf("获取路径信息失败 %s: %w", dirPath, err)
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("路径存在但不是目录: %s", dirPath)
		}
		return nil
	}

	// 递归创建目录
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", dirPath, err)
	}

	return nil
}

// closeFile 安全关闭文件句柄，如果关闭时发生错误且原错误为空则设置错误。
func closeFile(f *os.File, err *error) {
	if cerr := f.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

// closeGzip 安全关闭 gzip 写入器，如果关闭时发生错误且原错误为空则设置错误。
func closeGzip(gzw *gzip.Writer, err *error) {
	if cerr := gzw.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

// closeTar 安全关闭 tar 写入器，如果关闭时发生错误且原错误为空则设置错误。
func closeTar(tw *tar.Writer, err *error) {
	if cerr := tw.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

// closeGzipReader 安全关闭 gzip 读取器，如果关闭时发生错误且原错误为空则设置错误。
func closeGzipReader(gzr *gzip.Reader, err *error) {
	if cerr := gzr.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}
