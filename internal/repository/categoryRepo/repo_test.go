package categoryRepo

import (
	"context"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
	"testing"
)

func init() {
	// 加载配置文件
	_ = config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

func TestFindCategoryById(t *testing.T) {
	cate, err := FindCategoryById(context.Background(), "cat005")
	if err != nil {
		t.Error(err)
	} else {
		t.Log(cate)
	}
}
