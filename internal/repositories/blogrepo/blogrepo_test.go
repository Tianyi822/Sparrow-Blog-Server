package blogrepo

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
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

func TestAddBlog(t *testing.T) {
	tx := storage.Storage.Db.WithContext(context.Background()).Begin()

	blogDto := dto.BlogDto{
		BlogTitle:    "测试博客",
		BlogBrief:    "测试博客",
		CategoryId:   "category00001",
		BlogWordsNum: 100,
		BlogState:    true,
		BlogIsTop:    false,
	}
	err := AddBlog(tx, &blogDto)

	if err != nil {
		tx.Rollback()
		t.Error(err)
	}
	tx.Commit()

	t.Log(blogDto.BlogId)
}

func TestFindBlogsInPage(t *testing.T) {
	page, err := FindAllBlogs(context.Background(), false)
	if err != nil {
		t.Error(err)
	}

	for _, v := range page {
		t.Log(v)
	}
}

func TestDeleteBlogById(t *testing.T) {
	tx := storage.Storage.Db.WithContext(context.Background()).Begin()
	err := DeleteBlogById(tx, "9810076e28f4e75d")

	if err != nil {
		tx.Rollback()
		t.Error(err)
	}
	tx.Commit()
}

func TestChangeBlogStateById(t *testing.T) {
	tx := storage.Storage.Db.WithContext(context.Background()).Begin()
	err := ChangeBlogStateById(tx, "9810076e28f4e75d")

	if err != nil {
		tx.Rollback()
		t.Error(err)
	}
	tx.Commit()
}

func TestSetTopById(t *testing.T) {
	tx := storage.Storage.Db.WithContext(context.Background()).Begin()
	err := SetTopById(tx, "9810076e28f4e75d")

	if err != nil {
		tx.Rollback()
		t.Error(err)
	}
	tx.Commit()
}
