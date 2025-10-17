package readcountrepo

import (
	"context"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"
)

func init() {
	// 加载配置文件
	config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

func TestGetRecentSevenDaysReadCount(t *testing.T) {
	res, err := FindRecentSevenDaysReadCount(context.Background())
	if err != nil {
		t.Errorf("FindRecentSevenDaysReadCount error: %v", err)
	}
	for _, v := range res {
		t.Logf("res: %v - %v", v.ReadCount, v.ReadDate)
	}
}
