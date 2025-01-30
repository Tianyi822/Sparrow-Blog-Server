package aof

import (
	"context"
	"h2blog/cache/core"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"testing"
)

func init() {
	config.LoadConfig("../../resources/config/test/cache-config.yaml")
	_ = logger.InitLogger()
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
			cmd:     core.SET,
			args:    []string{"test-key", "test-value", "string", "0"},
			wantErr: false,
		},
		{
			name:    "store DELETE command",
			cmd:     core.DELETE,
			args:    []string{"test-key"},
			wantErr: false,
		},
		{
			name:    "store INCR command",
			cmd:     core.INCR,
			args:    []string{"test-key", "int"},
			wantErr: false,
		},
		{
			name:    "store CLEANUP command",
			cmd:     core.CLEANUP,
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "invalid SET command args",
			cmd:     core.SET,
			args:    []string{"test-key"},
			wantErr: true,
		},
		{
			name:    "invalid DELETE command args",
			cmd:     core.DELETE,
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid INCR command args",
			cmd:     core.INCR,
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
		{core.SET, []string{"key1", "value1", "string", "0"}},
		{core.SET, []string{"key2", "123", "int", "0"}},
		{core.DELETE, []string{"key3"}},
		{core.INCR, []string{"key4", "int"}},
		{core.CLEANUP, []string{}},
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
		if firstCmd[0] != core.SET || firstCmd[1] != "key1" || firstCmd[2] != "value1" ||
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
