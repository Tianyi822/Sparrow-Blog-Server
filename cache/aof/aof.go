package aof

import (
	"bufio"
	"context"
	"fmt"
	"h2blog/cache/core"
	"h2blog/pkg/config"
	"h2blog/pkg/fileTool"
	"h2blog/pkg/logger"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Aof represents an Append-Only File implementation for data persistence
// It provides thread-safe operations for storing and loading cache commands
type Aof struct {
	file *FileOp      // File operations handler
	mu   sync.RWMutex // Mutex for thread-safe operations
}

// NewAof creates and returns a new AOF instance.
// It initializes the AOF file with configuration from cache-config.yaml.
// Panics if file creation fails as this is critical for data persistence.
func NewAof() *Aof {
	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}

	fileOp, err := CreateFileOp(foConfig)
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}

	return &Aof{
		file: fileOp,
	}
}

// LoadFile reads and processes all AOF files (including compressed ones).
// It handles both regular .aof files and compressed .aof.tar.gz files.
// Files are processed in chronological order based on their timestamps.
// After successful loading, all processed files are cleaned up.
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

		// Find all AOF files
		files, err := filepath.Glob(filepath.Join(dir, prefix+"_*.aof.*"))
		if err != nil {
			return nil, fmt.Errorf("failed to list AOF files: %w", err)
		}

		// Sort files by timestamp for ordered processing
		sort.Slice(files, func(i, j int) bool {
			tsI := extractTimestamp(files[i])
			tsJ := extractTimestamp(files[j])
			return tsI < tsJ
		})

		var allCommands [][]string
		tempDir := filepath.Join(dir, "temp_aof")
		defer os.RemoveAll(tempDir) // Ensure temp directory cleanup

		// Process each file in order
		for _, f := range files {
			commands, err := processFile(f, tempDir)
			if err != nil {
				return nil, fmt.Errorf("failed to process file %s: %w", f, err)
			}
			allCommands = append(allCommands, commands...)
		}

		// Clean up processed files
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				logger.Warn("failed to remove processed AOF file %s: %v", f, err)
			}
		}

		logger.Info("AOF files loading completed: processed %d files, %d commands",
			len(files), len(allCommands))
		return allCommands, nil
	}
}

// Store writes a command to the AOF file.
// Commands are written in a specific format: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
// The operation is thread-safe and handles file rotation automatically.
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

// storeCommand handles the actual command storage logic
func (aof *Aof) storeCommand(cmd string, args ...string) error {
	switch cmd {
	case core.SET:
		if len(args) != 4 {
			return fmt.Errorf("SET command requires 4 args (key=%s, value=%s, type=%s, expired=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), safeGet(args, 2), safeGet(args, 3), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s;;%s;;%s", cmd, args[0], args[1], args[2], args[3])))

	case core.DELETE:
		if len(args) != 1 {
			return fmt.Errorf("DELETE command requires 1 arg (key=%s), got %d",
				safeGet(args, 0), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s", cmd, args[0])))

	case core.INCR:
		if len(args) != 2 {
			return fmt.Errorf("INCR command requires 2 args (key=%s, type=%s), got %d",
				safeGet(args, 0), safeGet(args, 1), len(args))
		}
		return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s", cmd, args[0], args[1])))

	case core.CLEANUP:
		if len(args) != 0 {
			return fmt.Errorf("CLEANUP command requires no args, got %d", len(args))
		}
		return aof.file.Write([]byte(cmd))

	default:
		return fmt.Errorf("unsupported command type: %s", cmd)
	}
}

// processFile handles a single AOF file, supporting both regular and compressed files
func processFile(path string, tempDir string) ([][]string, error) {
	var file *os.File
	var err error

	if strings.HasSuffix(path, ".tar.gz") {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}

		decompressedName := strings.TrimSuffix(filepath.Base(path), ".tar.gz")
		if err := fileTool.DecompressTarGz(path, decompressedName); err != nil {
			return nil, fmt.Errorf("failed to decompress %s: %w", path, err)
		}

		decompressedPath := filepath.Join(tempDir, decompressedName)
		file, err = os.Open(decompressedPath)
	} else {
		file, err = os.Open(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	return processAOFFile(scanner)
}

// extractTimestamp extracts the timestamp from an AOF filename
func extractTimestamp(filename string) int64 {
	parts := strings.Split(filepath.Base(filename), "_")
	if len(parts) < 2 {
		return 0
	}
	ts := strings.Split(parts[1], ".")[0]
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return timestamp
}

// processAOFFile processes the contents of an AOF file and returns the commands
func processAOFFile(scanner *bufio.Scanner) ([][]string, error) {
	var commands [][]string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		command := strings.Split(scanner.Text(), ";;")

		switch strings.TrimSpace(command[0]) {
		case core.SET:
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
		case core.DELETE:
			if len(command) != 2 {
				logger.Warn("line %d: invalid DELETE command format, skipping: %v", lineNum, command)
				continue
			}
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
			})
		case core.INCR:
			if len(command) != 3 {
				logger.Warn("line %d: invalid INCR command format, skipping: %v", lineNum, command)
				continue
			}
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]), // key
				strings.TrimSpace(command[2]), // type
			})
		case core.CLEANUP:
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

// safeGet safely retrieves an element from a slice
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return "<nil>"
	}
	return slice[index]
}
