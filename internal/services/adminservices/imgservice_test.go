package adminservices

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

func TestGetPresignUrlById(t *testing.T) {
	ctx := context.Background()
	url, err := GetImgPresignUrlById(ctx, "0ab6f800e0ea3270")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("url = %v\n", url)
	storage.Storage.Close(ctx)
}

func TestGetAllImgs(t *testing.T) {
	ctx := context.Background()

	imgs, err := GetAllImgs(ctx)
	if err != nil {
		t.Error(err)
	}

	for _, img := range imgs {
		url, err := storage.Storage.Cache.GetString(ctx, storage.BuildImgCacheKey(img.ImgId))
		if err != nil {
			t.Error(err)
		}
		t.Log(url)
	}
}

func TestDeleteImg(t *testing.T) {
	err := DeleteImg(context.Background(), "cbbc9654d0219858")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("success")
}

func TestRenameImgById(t *testing.T) {
	ctx := context.Background()
	err := RenameImgById(ctx, "6c76a6ece25a36c2", "test2")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rename success")
	storage.Storage.Close(ctx)
}

func TestIsExistImg(t *testing.T) {
	ctx := context.Background()
	exist, err := IsExistImgByName(ctx, "test22")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("exist: %v", exist)
}
