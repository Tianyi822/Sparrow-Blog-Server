package aof

import (
	"context"
	"errors"
	"fmt"
	"h2blog/cache/core"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"strings"
	"sync"
)

type Aof struct {
	file *FileOp
	mu   sync.RWMutex
}

// NewAof creates and returns a new AOF instance with the given configuration.
// If initialization fails, it will panic with the error message.
func NewAof() *Aof {
	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}

	fileOp, err := CreateFileOp(foConfig)
	if err != nil {
		// Since this is initialization code, we panic if we can't create the AOF file
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}

	return &Aof{
		file: fileOp,
	}
}

// LoadFile reads and parses the AOF file, returning a slice of command strings.
// Each command is split into components (operation type, key, value, etc.).
// Returns error if file reading fails or context is cancelled.
func (aof *Aof) LoadFile(ctx context.Context) ([][]string, error) {
	select {
	case <-ctx.Done():
		logger.Warn("AOF file loading cancelled")
		return nil, ctx.Err()
	default:
		// Start scanning AOF file
		// Each line format: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
		// Examples:
		// - SET;;key:int;;1;;int;;0
		// - SET;;key:string;;hello world;;string;;0
		// - DELETE;;key
		// - INCR;;key;;int
		// - INCR;;key;;uint
		// - CLEANUP
		//
		// Notes:
		// - CLEANALL is not recorded as it clears all cache and AOF files directly
		// - CLEANUP is recorded to ensure expired data is properly handled during recovery
		scanner, err := aof.file.GetScanner()
		if err != nil {
			logger.Error("failed to get AOF file scanner: %v", err)
			return nil, err
		}

		aof.mu.Lock()
		defer aof.mu.Unlock()

		logger.Info("starting to load AOF file")
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
				// Save SET command
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
			logger.Error("error scanning AOF file: %v", err)
			return nil, fmt.Errorf("error scanning file: %w", err)
		}

		logger.Info("AOF file loading completed: %d lines read, %d valid commands", lineNum, len(commands))
		return commands, nil
	}
}

// Store saves a command to the AOF file.
// File format per line: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
// Examples:
// - SET;;key:int;;1;;int;;0
// - SET;;key:string;;hello world;;string;;0
// - DELETE;;key
// - INCR;;key;;int
// - INCR;;key;;uint
// - CLEANUP
func (aof *Aof) Store(ctx context.Context, cmd string, args ...string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	select {
	case <-ctx.Done():
		logger.Warn("command storage cancelled")
		return ctx.Err()
	default:
		switch cmd {
		case core.SET:
			if len(args) != 4 {
				msg := fmt.Sprintf("SET command requires 4 args (key=%s, value=%s, type=%s, expired=%s), got %d",
					safeGet(args, 0), safeGet(args, 1), safeGet(args, 2), safeGet(args, 3), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("storing SET command: key=%s, type=%s", args[0], args[2])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s;;%s;;%s", cmd, args[0], args[1], args[2], args[3])))
		case core.DELETE:
			if len(args) != 1 {
				msg := fmt.Sprintf("DELETE command requires 1 arg (key=%s), got %d", safeGet(args, 0), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("storing DELETE command: key=%s", args[0])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s", cmd, args[0])))
		case core.INCR:
			if len(args) != 2 {
				msg := fmt.Sprintf("INCR command requires 2 args (key=%s, type=%s), got %d",
					safeGet(args, 0), safeGet(args, 1), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("storing INCR command: key=%s, type=%s", args[0], args[1])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s", cmd, args[0], args[1])))
		case core.CLEANUP:
			if len(args) != 0 {
				msg := fmt.Sprintf("CLEANUP command requires no args, got %d", len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("storing CLEANUP command")
			return aof.file.Write([]byte(cmd))
		default:
			msg := fmt.Sprintf("unsupported command type: %s", cmd)
			logger.Error(msg)
			return errors.New(msg)
		}
	}
}

// safeGet safely retrieves an element from a slice, avoiding index out of bounds.
// Returns "<nil>" if index is out of range.
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return "<nil>"
	}
	return slice[index]
}
