package readcountrepo

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"

	"github.com/stretchr/testify/assert"
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

// TestAddBlogReadCount 测试添加博客阅读数功能
func TestAddBlogReadCount(t *testing.T) {
	ctx := context.Background()

	// 开始事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// 准备测试数据
	testBlogId := "test_blog_1234"
	testReadCount := uint(666)

	brcd := &dto.BlogReadCountDto{
		BlogId:    testBlogId,
		ReadCount: testReadCount,
		ReadDate:  "20230101",
	}

	// 执行测试函数
	err := UpsertBlogReadCount(tx, brcd)

	assert.Nil(t, err)

	tx.Commit()
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
