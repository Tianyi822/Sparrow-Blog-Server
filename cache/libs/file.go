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

type FileOp struct {
	file           *os.File
	writer         *bufio.Writer
	isOpen         bool   // 用于判断是否可以进行操作
	needCompress   bool   // 是否需要压缩
	maxSize        int    // 以 MB 为单位
	path           string // 文件路径
	filePrefixName string // 文件前缀名
	fileSuffixName string // 文件后缀名
}

// FWConfig 日志文件配置项
type FWConfig struct {
	NeedCompress bool   // 是否需要压缩
	MaxSize      int    // 以 MB 为单位
	Path         string // 文件保存路径
}

// CreateFileWriter 只是创建一个文件操作对象，但不代表要立即操作这个文件，所以 isOpen 默认为 false
func CreateFileWriter(config *FWConfig) *FileOp {
	fileInfo := strings.Split(filepath.Base(config.Path), ".")

	return &FileOp{
		filePrefixName: fileInfo[0],
		fileSuffixName: fileInfo[1],
		path:           config.Path,
		needCompress:   config.NeedCompress,
		isOpen:         false,
		maxSize:        config.MaxSize,
	}
}

// ready 用于进行文件操作前的准备工作
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

// needSplit 判断是否需要进行分片
func (fw *FileOp) needSplit() bool {
	// 判断是否需要进行分片
	if fw.maxSize <= 0 {
		return false
	}

	// 判断文件大小是否超过最大值
	fileInfo, err := fw.file.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() > int64(fw.maxSize*1024*1024)
}

// Write 写入日志数据
// 该函数不做并发处理，传入的数据都是通过 channel 传递过来的，所以不需要考虑并发问题
// 并不会出现多个协程往同一个文件里面写数据，文件操作模块主要集中于对日志文件的分片管理，对历史日志打包
func (fw *FileOp) Write(context []byte) error {
	if !fw.isOpen {
		err := fw.ready()
		if err != nil {
			return err
		}
	}

	// 判断是否需要进行分片
	if fw.needSplit() {
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
			err = compressFileToTarGz(destPath)
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
	_, err := fw.writer.Write(buf)
	if err != nil {
		return err
	}
	// 数据落盘
	err = fw.writer.Flush()
	if err != nil {
		return err
	}
	return err
}

// isExist 判断路径是否存在
func isExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// mustOpenFile 直接打开文件，使用该方法的前提是确定文件一定存在
func mustOpenFile(realPath string) (*os.File, error) {
	file, err := os.OpenFile(realPath, os.O_APPEND|os.O_RDWR, 0666)
	return file, err
}

// createFile 创建文件，先检查文件是否存在，存在就报错，不存在就创建
func createFile(path string) (*os.File, error) {
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

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// remove 删除文件
func remove(path string) error {
	return os.RemoveAll(path)
}

// compressFileToTarGz 将文件打包压缩成 .tar.gz 文件
func compressFileToTarGz(src string) error {
	dir := filepath.Dir(src)
	filePrefixName := strings.Split(filepath.Base(src), ".")[0]
	dst := filepath.Join(dir, filePrefixName+".tar.gz")

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			panic(err)
		}
	}(destFile)

	// 创建Gzip压缩写入器
	gzw := gzip.NewWriter(destFile)
	defer func(gzw *gzip.Writer) {
		err := gzw.Close()
		if err != nil {
			panic(err)
		}
	}(gzw)

	// 创建Tar写入器
	tw := tar.NewWriter(gzw)
	defer func(tw *tar.Writer) {
		err := tw.Close()
		if err != nil {
			panic(err)
		}
	}(tw)

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			panic(err)
		}
	}(srcFile)

	// 获取源文件的信息
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 构建文件头信息
	header, err := tar.FileInfoHeader(srcFileInfo, "")
	if err != nil {
		return err
	}

	// 更新文件头中的路径信息
	header.Name = filepath.Base(src)

	// 写入文件头
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// 将源文件内容复制到Tar包中
	_, err = io.Copy(tw, srcFile)
	if err != nil {
		return err
	}

	return nil
}
