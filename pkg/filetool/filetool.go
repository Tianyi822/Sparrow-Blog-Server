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

// IsExist checks if a path exists
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// MustOpenFile opens an existing file with secure settings
// Security features:
// 1. Restricts file permissions to owner read/write (0600)
// 2. Validates file type (must be regular file)
// 3. Uses secure open flags
func MustOpenFile(realPath string) (*os.File, error) {
	// Validate file is regular
	fileInfo, err := os.Stat(realPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", realPath, err)
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", realPath)
	}

	file, err := os.OpenFile(realPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", realPath, err)
	}
	return file, nil
}

// CreateFile creates and initializes a new file
// Features:
// 1. Uses secure file permissions (0600)
// 2. Creates parent directories if needed
// 3. Validates filename format
// 4. Handles concurrent creation safely
func CreateFile(path string) (*os.File, error) {
	// Validate filename has extension
	if !strings.Contains(filepath.Base(path), ".") {
		return nil, fmt.Errorf("filename missing extension: %s", path)
	}

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if !IsExist(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create or open file with secure settings
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return file, nil
}

// CompressFileToTarGz compresses a file to tar.gz format
// Parameters:
//   - src: source file path
//   - dst: destination path for the compressed file (including .tar.gz extension)
//
// Returns error if compression fails
// If dst is empty, it will use default naming: <src_filename>.tar.gz in the same directory
func CompressFileToTarGz(src, dst string) error {
	if dst == "" {
		// Use default destination if not specified
		dir := filepath.Dir(src)
		filePrefixName := strings.Split(filepath.Base(src), ".")[0]
		dst = filepath.Join(dir, filePrefixName+".tar.gz")
	}

	// Validate file extensions
	if !strings.HasSuffix(dst, ".tar.gz") {
		return fmt.Errorf("destination file must have .tar.gz extension: %s", dst)
	}

	// Cleanup on panic
	defer func() {
		if r := recover(); r != nil {
			_ = os.RemoveAll(dst)
		}
	}()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer closeFile(destFile, &err)

	// Setup compression pipeline
	gzw := gzip.NewWriter(destFile)
	defer closeGzip(gzw, &err)

	tw := tar.NewWriter(gzw)
	defer closeTar(tw, &err)

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer closeFile(srcFile, &err)

	// Get source file info
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create and write tar header
	header, err := tar.FileInfoHeader(srcFileInfo, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}
	header.Name = filepath.Base(src)

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Copy file content
	if _, err = io.Copy(tw, srcFile); err != nil {
		return fmt.Errorf("failed to write compressed content: %w", err)
	}

	return nil
}

// DecompressTarGz extracts a file from a tar.gz archive
// Parameters:
//   - src: source archive path
//   - dst: destination filename (without path)
//
// Returns error if decompression fails
func DecompressTarGz(src, dst string) error {
	// Remove existing target file
	if IsExist(dst) {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing target file: %w", err)
		}
	}

	// Open source archive
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer closeFile(srcFile, &err)

	// Setup decompression pipeline
	gzr, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer closeGzipReader(gzr, &err)

	tr := tar.NewReader(gzr)

	// Read first file from archive
	header, err := tr.Next()
	if err == io.EOF {
		return fmt.Errorf("empty archive")
	}
	if err != nil {
		return fmt.Errorf("failed to read tar header: %w", err)
	}

	// Create target file
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}
	defer closeFile(file, &err)

	// Copy content
	if _, err := io.Copy(file, tr); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dst, err)
	}

	return nil
}

// Helper functions for proper resource cleanup

func closeFile(f *os.File, err *error) {
	if cerr := f.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

func closeGzip(gzw *gzip.Writer, err *error) {
	if cerr := gzw.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

func closeTar(tw *tar.Writer, err *error) {
	if cerr := tw.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}

func closeGzipReader(gzr *gzip.Reader, err *error) {
	if cerr := gzr.Close(); cerr != nil && *err == nil {
		*err = cerr
	}
}
