package categoryrepo

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
	config.LoadConfig()
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

func TestAddCategory(t *testing.T) {
	ctx := context.Background()
	tx := storage.Storage.Db.WithContext(ctx).Begin()

	err := AddCategory(tx, &dto.CategoryDto{
		CategoryId:   "cat006",
		CategoryName: "测试分类",
	})

	if err != nil {
		tx.Rollback()
		t.Error(err)
	} else {
		tx.Commit()
		t.Log("添加分类成功")
	}
}

func TestGetAllCategories(t *testing.T) {
	categories, err := FindAllCategories(context.Background())
	if err != nil {
		t.Error(err)
	}

	for _, cate := range categories {
		t.Log(cate)
	}
}
