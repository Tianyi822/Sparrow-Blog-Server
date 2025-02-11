package logger

import (
	"context"
	"h2blog_server/pkg/config"
	"testing"
)

func init() {
	// 加载配置文件
	config.LoadConfig()
	// 初始化 Logger 组件
	_ = InitLogger(context.Background())
}

func TestLogger(t *testing.T) {
	Info("info -- test logger component")
	Debug("debug -- test logger component")
	Warn("warn -- test logger component")
	Panic("panic -- test logger component")
	//Fatal("fatal -- test logger component")
}
