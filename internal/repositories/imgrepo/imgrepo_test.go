package imgrepo

import (
	"context"
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

func TestGetAllImgs(t *testing.T) {
	imgs, err := FindAllImgs(context.Background())
	if err != nil {
		t.Error(err)
	}

	for _, img := range imgs {
		t.Log(img)
	}
}

func TestUpdateImgNameById(t *testing.T) {
	tx := storage.Storage.Db.WithContext(context.Background()).Begin()

	if err := UpdateImgNameById(tx, "85fba0685fa281a5", "test"); err != nil {
		tx.Rollback()
		t.Error(err)
	}
	tx.Commit()

	t.Log("更新成功")
}
