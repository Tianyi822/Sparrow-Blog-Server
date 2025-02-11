package aof

import (
	"bufio"
	"fmt"
	"h2blog/pkg/config"
	"h2blog/pkg/fileTool"
	"h2blog/pkg/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileOp Core file operation structure responsible for managing file lifecycle,
// buffered writing, automatic splitting and compression operations.
// It provides thread-safe file operations with features like:
// - Buffered writing for better performance
// - Automatic file rotation when size limit is reached
// - Optional compression of rotated files
// - Concurrent access protection
type FileOp struct {
	rwMu           sync.RWMutex  // RWMutex to protect concurrent access
	file           *os.File      // Underlying file handle (nil if not opened)
	writer         *bufio.Writer // Buffered writer (32KB buffer)
	isOpen         bool          // File writable status flag
	needCompress   bool          // Whether to compress files after rotation
	maxSize        uint16        // Maximum file size before rotation (in MB)
	path           string        // Current active file absolute path
	filePrefixName string        // Base filename without extension
	fileSuffixName string        // File extension without dot
}

// CreateFileOp initializes a new FileOp instance with the given configuration.
// It validates the configuration and sets up the file operation structure.
// The actual file is not opened until the first write operation.
func CreateFileOp() (*FileOp, error) {
	// Validate required configuration
	if config.Cache.Aof.Path == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}
	if config.Cache.Aof.MaxSize < 0 {
		return nil, fmt.Errorf("max size cannot be negative")
	}

	// Split the file path into components
	baseName := filepath.Base(config.Cache.Aof.Path)
	ext := filepath.Ext(baseName)
	prefix := strings.TrimSuffix(baseName, ext)

	return &FileOp{
		filePrefixName: prefix,
		fileSuffixName: strings.TrimPrefix(ext, "."),
		path:           config.Cache.Aof.Path,
		needCompress:   config.Cache.Aof.Compress,
		maxSize:        config.Cache.Aof.MaxSize,
	}, nil
}

// GetScanner returns a bufio.Scanner for reading the file contents.
// If the file doesn't exist, returns a scanner with empty content.
// The scanner is configured with a 512KB buffer for handling long lines.
func (fop *FileOp) GetScanner() (*bufio.Scanner, error) {
	if fop == nil {
		return nil, fmt.Errorf("FileOp is nil")
	}

	fop.rwMu.RLock()
	defer fop.rwMu.RUnlock()

	f, err := os.OpenFile(fop.path, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return bufio.NewScanner(strings.NewReader("")), nil
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f) // Close file after scanner is created

	scanner := bufio.NewScanner(f)
	const maxCapacity = 512 * 1024 // 512KB buffer for long lines
	buf := make([]byte, 0, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	return scanner, nil
}

// ready prepares the file for write operations.
// It creates the directory if needed, opens or creates the file,
// and initializes the buffered writer.
func (fop *FileOp) ready() error {
	if fop.file != nil {
		return nil // File is already open
	}

	// Ensure directory exists
	dir := filepath.Dir(fop.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open existing file or create new one
	var err error
	if fileTool.IsExist(fop.path) {
		fop.file, err = fileTool.MustOpenFile(fop.path)
	} else {
		fop.file, err = fileTool.CreateFile(fop.path)
	}
	if err != nil {
		return fmt.Errorf("failed to open/create file: %w", err)
	}

	// Initialize buffered writer with 32KB buffer
	fop.writer = bufio.NewWriterSize(fop.file, 32*1024)
	fop.isOpen = true
	return nil
}

// Close flushes any buffered data and closes the file.
// It's safe to call Close multiple times.
func (fop *FileOp) Close() error {
	if !fop.isOpen {
		return nil
	}

	// Flush buffered data
	if fop.writer != nil {
		if err := fop.writer.Flush(); err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
	}

	// Ensure data is written to disk
	if err := fop.file.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Close file handle
	if err := fop.file.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// Reset file operation state
	fop.isOpen = false
	fop.writer = nil
	fop.file = nil

	return nil
}

// needSplit checks if the current file size exceeds the configured maximum size.
// Returns true if the file should be rotated.
func (fop *FileOp) needSplit() bool {
	// Skip if rotation is disabled
	if fop.maxSize <= 0 {
		return false
	}

	// Get current file size
	fileInfo, err := fop.file.Stat()
	if err != nil {
		return false
	}

	// Compare with max size (converting MB to bytes)
	return fileInfo.Size() > int64(fop.maxSize*1024*1024)
}

// Write appends data to the file with a newline character.
// It handles file rotation automatically if the size limit is reached.
// The operation is thread-safe and buffered for better performance.
func (fop *FileOp) Write(context []byte) error {
	if fop == nil {
		return fmt.Errorf("FileOp is nil")
	}
	if len(context) == 0 {
		return nil
	}

	fop.rwMu.Lock()
	defer fop.rwMu.Unlock()

	// Ensure file is ready for writing
	if !fop.isOpen {
		if err := fop.ready(); err != nil {
			return err
		}
	}

	// Check if rotation is needed before writing
	if fop.needSplit() {
		if err := fop.checkAndRotate(); err != nil {
			return fmt.Errorf("rotation failed: %w", err)
		}
	}

	// Prepare data with newline
	buf := make([]byte, len(context)+1)
	copy(buf, context)
	buf[len(context)] = '\n'

	// Write data
	if _, err := fop.writer.Write(buf); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	// Always flush after write to ensure data is written
	if err := fop.writer.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}

	return nil
}

// checkAndRotate handles the file rotation logic when size limit is reached.
func (fop *FileOp) checkAndRotate() error {
	if !fop.needSplit() {
		return nil
	}

	// Flush buffered data before closing
	if err := fop.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer before rotation: %w", err)
	}

	// Sync to disk
	if err := fop.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file before rotation: %w", err)
	}

	// Close current file
	if err := fop.Close(); err != nil {
		return err
	}

	// Generate rotated filename
	rotatedName := fmt.Sprintf("%v.aof", fop.filePrefixName)
	destPath := filepath.Join(filepath.Dir(fop.path), rotatedName)

	// Rename current file
	if err := os.Rename(fop.path, destPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Handle compression if enabled
	if fop.needCompress {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		compressedPath := fmt.Sprintf("%v_%v.aof.tar.gz", fop.filePrefixName, timestamp)
		compressedPath = filepath.Join(filepath.Dir(fop.path), compressedPath)

		// Compress file
		if err := fileTool.CompressFileToTarGz(destPath, compressedPath); err != nil {
			// If compression fails, try to restore the original file
			_ = os.Rename(destPath, fop.path)
			return fmt.Errorf("compression failed: %w", err)
		}

		// Verify compressed file exists before removing original
		if !fileTool.IsExist(compressedPath) {
			_ = os.Rename(destPath, fop.path)
			return fmt.Errorf("compressed file not found after compression")
		}

		// Remove original file only after successful compression
		if err := os.RemoveAll(destPath); err != nil {
			logger.Warn("failed to remove rotated file after compression: %v", err)
		}
	}

	// Create new file for writing
	return fop.ready()
}
