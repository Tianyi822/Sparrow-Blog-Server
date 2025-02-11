package aof

import (
	"bytes"
	"context"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"testing"
)

func init() {
	config.LoadConfig()
	_ = logger.InitLogger(context.Background())
}

func TestFileOp_Write(t *testing.T) {
	// 初始化FileOp配置（1MB分割）
	fo, _ := CreateFileOp()

	// 写法 1
	for i := 0; i < 10; i++ {
		err := fo.Write(bytes.Repeat([]byte("a"), 1024*1024+512))
		if err != nil {
			t.Fatalf("写入日志发生错误: %v", err)
		}
	}

	// 强制关闭并添加延迟确保文件释放
	if err := fo.Close(); err != nil {
		t.Fatalf("关闭文件失败: %v", err)
	}
}
