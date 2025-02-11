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

// Package aof implements an Append-Only File (AOF) persistence mechanism for caching systems.
// It provides thread-safe operations for storing and loading cache commands, with support
// for file rotation, compression, and automatic cleanup.

// Aof represents an Append-Only File implementation for data persistence.
// It provides thread-safe operations for storing and loading cache commands.
// Features:
// - Thread-safe operations using mutex
// - Automatic file rotation based on size
// - Support for compressed file storage
// - Chronological command processing
// - Automatic cleanup of processed files
type Aof struct {
	file *FileOp      // Handles low-level file operations including rotation and compression
	mu   sync.RWMutex // Ensures thread-safe access to file operations
}

// NewAof creates and returns a new AOF instance.
// It initializes the AOF file with configuration from cache-config.yaml.
// Configuration includes:
// - File path for AOF storage
// - Maximum file size before rotation
// - Compression settings for rotated files
//
// Panics if file creation fails as this is critical for data persistence.
// This is a startup-critical operation - if it fails, the application cannot
// guarantee data persistence and should not continue.
func NewAof() *Aof {
	fileOp, err := CreateFileOp()
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}

	return &Aof{
		file: fileOp,
	}
}

// LoadFile reads and processes all AOF files (including compressed ones).
// Process:
// 1. Lists all AOF files in the directory (both .aof and .aof.tar.gz)
// 2. Sorts files by timestamp to maintain operation order
// 3. Creates temporary directory for decompression if needed
// 4. Processes each file in chronological order
// 5. Cleans up processed files after successful loading
// 6. Removes temporary directory
//
// Thread-safety:
// - Uses mutex to prevent concurrent access during loading
// - Supports context cancellation for graceful shutdown
//
// Error handling:
// - Returns error if file listing fails
// - Returns error if any file processing fails
// - Logs warnings for non-critical failures (e.g., file cleanup)
//
// Returns:
// - [][]string: Slice of command arguments, each inner slice represents one command
// - error: Any error encountered during the process
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

		// Create temp directory first
		tempDir := filepath.Join(dir, "temp_aof")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		var files []string
		var err error

		// First, get all compressed files
		if aof.file.needCompress {
			files, err = filepath.Glob(filepath.Join(dir, prefix+"_*.aof.tar.gz"))
			if err != nil {
				return nil, fmt.Errorf("failed to list compressed AOF files: %w", err)
			}
			// Sort compressed files by timestamp
			sort.Slice(files, func(i, j int) bool {
				tsI := extractTimestamp(files[i])
				tsJ := extractTimestamp(files[j])
				return tsI < tsJ
			})
		}

		// Then check for current AOF file
		currentFile := filepath.Join(dir, prefix+".aof")
		if fileTool.IsExist(currentFile) {
			files = append(files, currentFile)
		}

		var allCommands [][]string
		logger.Info("Processing %d files", len(files))

		// Process each file
		for i, f := range files {
			logger.Info("Processing file %d/%d: %s", i+1, len(files), f)
			commands, err := processFile(f, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to process file %s: %w", f, err)
			}
			logger.Info("File %s: loaded %d commands", f, len(commands))
			allCommands = append(allCommands, commands...)
		}

		// Clean up processed files
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				logger.Warn("failed to remove processed AOF file %s: %v", f, err)
			}
		}

		// Create new AOF file
		if err := aof.file.ready(); err != nil {
			return nil, fmt.Errorf("failed to create new AOF file: %w", err)
		}

		logger.Info("AOF files loading completed: processed %d files, %d commands",
			len(files), len(allCommands))
		return allCommands, nil
	}
}

// Store writes a command to the AOF file.
// Format: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
//
// Supported commands:
// - SET: requires 4 args (key, value, type, expiry)
// - DELETE: requires 1 arg (key)
// - INCR: requires 2 args (key, type)
// - CLEANUP: requires no args
//
// Thread-safety:
// - Uses mutex to prevent concurrent writes
// - Supports context cancellation
//
// Parameters:
// - ctx: Context for cancellation
// - cmd: Command type (SET, DELETE, INCR, CLEANUP)
// - args: Command arguments (varies by command type)
//
// Returns error if:
// - Context is cancelled
// - Invalid number of arguments
// - Unsupported command type
// - Write operation fails
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

// Internal helper functions

// storeCommand handles the actual command storage logic.
// It validates command arguments and formats the command string.
//
// Command formats:
// - SET: SET;;key;;value;;type;;expiry
// - DELETE: DELETE;;key
// - INCR: INCR;;key;;type
// - CLEANUP: CLEANUP
//
// Validation:
// - Checks argument count for each command type
// - Ensures all required arguments are present
//
// Returns error if:
// - Invalid number of arguments
// - Write operation fails
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

// processFile handles AOF files based on compression setting
func processFile(path string, tempDir string) ([][]string, error) {
	var file *os.File
	var err error

	if strings.HasSuffix(path, ".tar.gz") {
		// Handle compressed file
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		decompressedPath := filepath.Join(tempDir, strings.TrimSuffix(filepath.Base(path), ".tar.gz"))
		if err := fileTool.DecompressTarGz(path, decompressedPath); err != nil {
			return nil, fmt.Errorf("failed to decompress %s: %w", path, err)
		}

		file, err = os.Open(decompressedPath)
	} else {
		// Handle regular file
		file, err = os.Open(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	return processAOFFile(scanner)
}

// extractTimestamp extracts the timestamp from an AOF filename.
// Expected filename format: prefix_timestamp.aof.* or prefix_timestamp.aof.tar.gz
//
// Parameters:
// - filename: The AOF filename to parse
//
// Returns:
// - int64: Unix timestamp extracted from filename
// - Returns 0 if filename format is invalid
func extractTimestamp(filename string) int64 {
	parts := strings.Split(filepath.Base(filename), "_")
	if len(parts) < 2 {
		return 0
	}
	ts := strings.Split(parts[1], ".")[0]
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return timestamp
}

// processAOFFile processes the contents of an AOF file and returns the commands.
// It reads the file line by line and parses each command according to its format.
//
// Command validation:
// - Checks command format and argument count
// - Trims whitespace from all fields
// - Logs warnings for invalid commands but continues processing
//
// Error handling:
// - Skips invalid commands with warning
// - Returns error if scanner encounters read error
//
// Returns:
// - [][]string: Slice of valid commands
// - error: Any error encountered during processing
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

// safeGet safely retrieves an element from a slice.
// Prevents panic from index out of bounds by returning "<nil>"
// for invalid indices.
//
// Parameters:
// - slice: The string slice to access
// - index: The index to retrieve
//
// Returns:
// - string: The element at the index or "<nil>" if index is invalid
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return "<nil>"
	}
	return slice[index]
}
