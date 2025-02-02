package aof

import (
	"context"
	"crypto/rand"
	"fmt"
	"h2blog/cache/common"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"math/big"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func init() {
	config.LoadConfig()
	_ = logger.InitLogger(context.Background())
}

func TestAof_Store(t *testing.T) {
	aof := NewAof()
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "store SET command",
			cmd:     common.SET,
			args:    []string{"test-key", "test-value", "string", "0"},
			wantErr: false,
		},
		{
			name:    "store DELETE command",
			cmd:     common.DELETE,
			args:    []string{"test-key"},
			wantErr: false,
		},
		{
			name:    "store INCR command",
			cmd:     common.INCR,
			args:    []string{"test-key", "int"},
			wantErr: false,
		},
		{
			name:    "store CLEANUP command",
			cmd:     common.CLEANUP,
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "invalid SET command args",
			cmd:     common.SET,
			args:    []string{"test-key"},
			wantErr: true,
		},
		{
			name:    "invalid DELETE command args",
			cmd:     common.DELETE,
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid INCR command args",
			cmd:     common.INCR,
			args:    []string{"test-key"},
			wantErr: true,
		},
		{
			name:    "unsupported command",
			cmd:     "UNKNOWN",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := aof.Store(ctx, tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAof_LoadFile(t *testing.T) {
	aof := NewAof()
	ctx := context.Background()

	// 写入测试数据
	testCommands := []struct {
		cmd  string
		args []string
	}{
		{common.SET, []string{"key1", "value1", "string", "0"}},
		{common.SET, []string{"key2", "123", "int", "0"}},
		{common.DELETE, []string{"key3"}},
		{common.INCR, []string{"key4", "int"}},
		{common.CLEANUP, []string{}},
	}

	for _, cmd := range testCommands {
		err := aof.Store(ctx, cmd.cmd, cmd.args...)
		if err != nil {
			t.Fatalf("Failed to store test data: %v", err)
		}
	}

	// 测试加载文件
	commands, err := aof.LoadFile(ctx)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	// 验证加载的命令数量
	expectedCmdCount := len(testCommands)
	if len(commands) != expectedCmdCount {
		t.Errorf("LoadFile() got %d commands, want %d", len(commands), expectedCmdCount)
	}

	// 验证第一个 SET 命令的内容
	if len(commands) > 0 {
		firstCmd := commands[0]
		if firstCmd[0] != common.SET || firstCmd[1] != "key1" || firstCmd[2] != "value1" ||
			firstCmd[3] != "string" || firstCmd[4] != "0" {
			t.Errorf("First command incorrect: got %v", firstCmd)
		}
	}

	// 测试上下文取消
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = aof.LoadFile(cancelCtx)
	if err == nil {
		t.Error("LoadFile() with cancelled context should return error")
	}
}

func TestAOFWriteAndLoad(t *testing.T) {
	// Initialize AOF with test configuration
	config.Cache.Aof.MaxSize = 10 // Set to 10MB for testing
	aof := NewAof()
	ctx := context.Background()

	const dataCount = 66666
	testData := generateTestDataWithCount(dataCount)

	t.Logf("Starting to write %d entries...", dataCount)
	startTime := time.Now()

	// Write test data
	for i, key := range getSortedKeys(testData) {
		if i > 0 && i%5000 == 0 {
			t.Logf("Written %d entries...", i)
		}
		err := aof.Store(ctx, common.SET, key, testData[key], "string", "0")
		if err != nil {
			t.Fatalf("Failed to store data at index %d: %v", i, err)
		}
	}

	t.Logf("Write completed in %v", time.Since(startTime))

	// Force close and flush
	if err := aof.file.Close(); err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	// Load and verify
	t.Log("Starting data load...")
	loadStartTime := time.Now()
	commands, err := aof.LoadFile(ctx)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}
	t.Logf("Load completed in %v", time.Since(loadStartTime))

	// Verify loaded data
	t.Log("Verifying loaded data...")
	loadedData := make(map[string]string)
	for _, cmd := range commands {
		if cmd[0] == common.SET {
			loadedData[cmd[1]] = cmd[2]
		}
	}

	// Compare data count
	if len(loadedData) != dataCount {
		t.Errorf("Data count mismatch. Expected %d, got %d", dataCount, len(loadedData))
	}

	// Compare data content
	for key, value := range testData {
		if loadedValue, exists := loadedData[key]; !exists {
			t.Errorf("Key not found in loaded data: %s", key)
		} else if loadedValue != value {
			t.Errorf("Value mismatch for key %s. Expected %s, got %s", key, value, loadedValue)
		}
	}

	// Log statistics
	t.Logf("Successfully processed %d commands", len(commands))
	t.Logf("Original data count: %d, Loaded data count: %d", len(testData), len(loadedData))

	// Verify file rotation
	verifyFileRotation(t, aof.file.path)
}

func getSortedKeys(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// generateTestDataWithCount generates test data with a fixed number of entries
func generateTestDataWithCount(count int) map[string]string {
	data := make(map[string]string)
	valueSize := 100 // 100 bytes per value

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := generateRandomString(valueSize)
		data[key] = value
	}

	return data
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err) // 在测试中，遇到错误直接 panic 是可以接受的
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func verifyFileRotation(t *testing.T, aofPath string) {
	dir := filepath.Dir(aofPath)
	files, err := filepath.Glob(filepath.Join(dir, "*.aof*"))
	if err != nil {
		t.Fatalf("Failed to list AOF files: %v", err)
	}

	// Should only have the current AOF file after loading
	if len(files) != 1 {
		t.Errorf("Expected 1 AOF file after loading, found %d files", len(files))
		for _, f := range files {
			t.Logf("Found file: %s", f)
		}
	}
}
