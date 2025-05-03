package sysservices

import (
	"context"
	"errors"
	"h2blog_server/cache"
	"h2blog_server/internal/repositories/imgrepo"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"time"
)

// GetImgPresignUrlById 根据图片ID获取预签名URL。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - imgId: 图片的唯一标识符，用于定位具体的图片资源。
//
// 返回值:
//   - string: 预签名URL，用于访问图片资源。
//   - error: 如果在获取缓存、查询图片信息或生成预签名URL过程中发生错误，则返回相应的错误信息。
func GetImgPresignUrlById(ctx context.Context, imgId string) (string, error) {
	// 尝试从缓存中获取图片的预签名URL。
	url, err := storage.Storage.Cache.GetString(ctx, storage.BuildImgCacheKey(imgId))
	if errors.Is(err, cache.ErrNotFound) {
		// 如果缓存中未找到对应的URL，则从数据库中查询图片信息。
		imgDto, err := imgrepo.FindImgById(ctx, imgId)
		if err != nil {
			return "", err
		}

		// 根据图片名称和类型生成OSS存储路径。
		ossPath := ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)

		// 为生成的OSS路径创建预签名URL，有效期为35分钟。
		presign, err := storage.Storage.GenPreSignUrl(ctx, ossPath, imgDto.ImgType, ossstore.Get, 35*time.Minute)
		if err != nil {
			return "", err
		}

		// 缓存预签名URL到缓存中。
		err = storage.Storage.Cache.SetWithExpired(ctx, storage.BuildImgCacheKey(imgId), presign.URL, 30*time.Minute)
		if err != nil {
			return "", err
		}

		url = presign.URL
	}

	// 返回获取到的预签名URL。
	return url, nil
}
