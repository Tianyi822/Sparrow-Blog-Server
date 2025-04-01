package imgService

import (
	"context"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webp"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"strings"
	"testing"
	"time"
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

func TestConvertAndAddImg(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	files, err := storage.Storage.ListOssDirFiles(ctx, config.Oss.ImageOssPath)
	if err != nil {
		t.Fatal(err)
	}

	imgDtos := make([]dto.ImgDto, 0, len(files))
	for _, file := range files {
		imgInfo := strings.Split(file, "/")[1]
		imgName := strings.Split(imgInfo, ".")[0]
		imgType := strings.Split(imgInfo, ".")[1]
		imgDto := dto.ImgDto{
			ImgName: imgName,
		}
		switch imgType {
		case "jpg":
			imgDto.ImgType = ossstore.JPG
		case "jpeg":
			imgDto.ImgType = ossstore.JPEG
		case "png":
			imgDto.ImgType = ossstore.PNG
		default:
			continue
		}
		imgDtos = append(imgDtos, imgDto)
	}

	imgsDto := &dto.ImgsDto{
		Imgs: imgDtos,
	}

	imgsVo, err := ConvertAndAddImg(ctx, imgsDto)
	if err != nil {
		t.Fatal(err)
	}
	for _, imgVo := range imgsVo.Success {
		t.Logf("imgVo Successed: %#v\n", imgVo)
	}
	for _, imgVo := range imgsVo.Failure {
		t.Logf("imgVo Failed: %#v\n", imgVo)
	}
	t.Logf("success len = %v\n", len(imgsVo.Success))
	t.Logf("fail len = %v\n", len(imgsVo.Failure))
}

func TestDeleteImgs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// First add some test images to have something to delete
	files, err := storage.Storage.ListOssDirFiles(ctx, config.Oss.ImageOssPath)
	if err != nil {
		t.Fatal(err)
	}

	imgDtos := make([]dto.ImgDto, 0, len(files))
	for _, file := range files {
		imgInfo := strings.Split(file, "/")[1]
		imgName := strings.Split(imgInfo, ".")[0]
		imgType := strings.Split(imgInfo, ".")[1]
		imgDto := dto.ImgDto{
			ImgName: imgName,
		}
		switch imgType {
		case "jpg":
			imgDto.ImgType = ossstore.JPG
		case "jpeg":
			imgDto.ImgType = ossstore.JPEG
		case "png":
			imgDto.ImgType = ossstore.PNG
		default:
			continue
		}
		imgDtos = append(imgDtos, imgDto)
	}

	imgsDto := &dto.ImgsDto{
		Imgs: imgDtos,
	}

	addedImgs, err := ConvertAndAddImg(ctx, imgsDto)
	if err != nil {
		t.Fatal(err)
	}

	// Get IDs of added images
	var imgIds []string
	for _, img := range addedImgs.Success {
		imgIds = append(imgIds, img.ImgId)
	}

	// Test deletion
	imgsVo, err := DeleteImgs(ctx, imgIds)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results
	t.Logf("Successfully deleted images: %d", len(imgsVo.Success))
	for _, img := range imgsVo.Success {
		t.Logf("Deleted image: %#v", img)
	}

	t.Logf("Failed to delete images: %d", len(imgsVo.Failure))
	for _, img := range imgsVo.Failure {
		t.Logf("Failed to delete image: %#v", img)
	}

	// Test with non-existent IDs
	nonExistentIds := []string{"non-existent-1", "non-existent-2"}
	imgsVo, err = DeleteImgs(ctx, nonExistentIds)
	if err != nil {
		t.Fatal(err)
	}

	if len(imgsVo.Success) != 0 {
		t.Errorf("Expected 0 successful deletions for non-existent IDs, got %d", len(imgsVo.Success))
	}
	if len(imgsVo.Failure) != len(nonExistentIds) {
		t.Errorf("Expected %d failed deletions for non-existent IDs, got %d", len(nonExistentIds), len(imgsVo.Failure))
	}
}

func TestRenameImgs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// First add a test image to have something to rename
	imgDto := dto.ImgDto{
		ImgName: "test-image",
		ImgType: ossstore.JPG,
	}
	imgsDto := &dto.ImgsDto{
		Imgs: []dto.ImgDto{imgDto},
	}

	addedImgs, err := ConvertAndAddImg(ctx, imgsDto)
	if err != nil {
		t.Fatal(err)
	}

	if len(addedImgs.Success) == 0 {
		t.Fatal("Failed to add test image")
	}

	imgId := addedImgs.Success[0].ImgId
	newName := "renamed-test-image"

	// Test renaming
	renamedImg, err := RenameImgs(ctx, imgId, newName)
	if err != nil {
		t.Fatal(err)
	}

	// Verify results
	if renamedImg.ImgId != imgId {
		t.Errorf("Expected ImgId %s, got %s", imgId, renamedImg.ImgId)
	}
	if renamedImg.ImgName != newName {
		t.Errorf("Expected ImgName %s, got %s", newName, renamedImg.ImgName)
	}

	// Test with non-existent ID
	_, err = RenameImgs(ctx, "non-existent-id", "new-name")
	if err == nil {
		t.Error("Expected error when renaming non-existent image, got nil")
	}

	// Cleanup
	_, err = DeleteImgs(ctx, []string{imgId})
	if err != nil {
		t.Logf("Failed to cleanup test image: %v", err)
	}
}

func TestFindNameLikeImgs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	imgs, err := FindImgsByNameLike(ctx, "萤")
	if err != nil {
		t.Fatal(err)
	}

	for _, img := range imgs.Success {
		t.Logf("Found image: %#v", img)
	}
}
