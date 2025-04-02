package adminService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webp"
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
	// 初始化转换器
	_ = webp.InitConverter(context.Background())
}

func TestGetBlogsInPage(t *testing.T) {
	blogDtos, err := GetBlogsToAdminPosts(context.Background())

	if err != nil {
		t.Error(err)
	}

	for _, blogDto := range blogDtos {
		t.Log(blogDto)
	}
}

func TestUpdateBlogData(t *testing.T) {
	blogDto := &dto.BlogDto{
		BlogId:     "blog00006",
		BlogTitle:  "test6",
		BlogBrief:  "test",
		Category:   dto.CategoryDto{CategoryName: "测试分类 7"},
		CategoryId: "",
		BlogIsTop:  false,
		BlogState:  true,
		Tags: []dto.TagDto{
			{TagId: "tag001", TagName: "Java"},
			{TagName: "tag00006"},
		},
	}

	presignUrl, err := UpdateOrAddBlog(context.Background(), blogDto)

	if err != nil {
		t.Error(err)
	}
	t.Log(presignUrl)
}

func TestAddBlog(t *testing.T) {
	blogDto := &dto.BlogDto{
		BlogTitle: "test6",
		BlogBrief: "test",
		Category: dto.CategoryDto{
			CategoryName: "测试分类 6",
		},
		CategoryId: "",
		BlogIsTop:  false,
		BlogState:  true,
		Tags: []dto.TagDto{
			{TagId: "tag001", TagName: "Java"},
			{TagName: "TEST_TAG"},
		},
	}

	presignUrl, err := UpdateOrAddBlog(context.Background(), blogDto)

	if err != nil {
		t.Error(err)
	}
	t.Log(blogDto.BlogId)
	t.Log(presignUrl)
}

func TestDeleteBlog(t *testing.T) {
	err := DeleteBlogById(context.Background(), "d5fbb6bf6c5b3a27")
	if err != nil {
		t.Error(err)
	}
	t.Log("success")
}

func TestGetBlogData(t *testing.T) {
	blogDto, url, err := GetBlogData(context.Background(), "0e317bad75975c0a")
	if err != nil {
		t.Error(err)
	}

	// 格式化输出
	t.Logf("%+v", blogDto)
	t.Log(url)
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
	}
	t.Log("success")
}
