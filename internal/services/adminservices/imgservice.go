package adminservices

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/cache"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repositories/imgrepo"
	"h2blog_server/pkg/logger"
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

// GetAllImgs 获取所有图片的基本信息，并为每张图片生成预签名的访问链接。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//
// 返回值:
//   - []dto.ImgDto: 包含所有图片信息的切片，每张图片的URL字段已更新为预签名链接。
//   - error: 如果在获取图片信息、生成预签名链接或缓存操作中发生错误，则返回相应的错误信息。
func GetAllImgs(ctx context.Context) ([]dto.ImgDto, error) {
	// 从存储库中获取所有图片的基本信息。
	imgs, err := imgrepo.FindAllImgs(ctx)
	if err != nil {
		return nil, err
	}

	// 遍历每张图片，为其生成预签名的访问链接并更新图片的URL字段。
	for _, img := range imgs {
		// 检查缓存中是否已存在该图片的预签名链接，如果不存在则将其写入缓存。
		cacheKey := storage.BuildImgCacheKey(img.ImgId)
		_, err = storage.Storage.Cache.GetString(ctx, cacheKey)
		if errors.Is(err, cache.ErrNotFound) {
			// 根据图片名称和类型生成OSS存储路径。
			path := ossstore.GenOssSavePath(img.ImgName, img.ImgType)

			// 为图片生成预签名的访问链接，有效期为30分钟。
			presign, err := storage.Storage.GenPreSignUrl(
				ctx,
				path,
				img.ImgType,
				ossstore.Get,
				35*time.Minute,
			)
			if err != nil {
				// 如果生成预签名链接失败，记录错误日志并返回错误。
				msg := fmt.Sprintf("获取图片链接失败: %v", err)
				logger.Error(msg)
				return nil, err
			}

			err = storage.Storage.Cache.SetWithExpired(ctx, cacheKey, presign.URL, 30*time.Minute)
			if err != nil {
				// 如果缓存写入失败，记录错误日志并返回错误。
				msg := fmt.Sprintf("缓存图片链接失败: %v", err)
				logger.Error(msg)
				return nil, err
			}
		}
	}

	// 返回包含预签名链接的图片信息切片。
	return imgs, nil
}

// AddImgs 批量添加图片信息到数据库。
// 参数：
//   - ctx context.Context: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - imgs []dto.ImgDto: 包含图片信息的 DTO（数据传输对象）切片，每个元素代表一张图片的信息。
//
// 返回值：
//   - error: 如果操作成功，则返回 nil；如果发生错误，则返回具体的错误信息。
func AddImgs(ctx context.Context, imgs []dto.ImgDto) error {
	// 开启事务，确保批量操作的原子性。
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 捕获 panic，确保在发生异常时回滚事务，避免数据库处于不一致状态。
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := imgrepo.AddImgBatch(tx, imgs); err != nil {
		return err
	}
	tx.Commit()

	return nil
}

// IsExistImgByName 检查指定的图片是否存在于数据库和OSS存储中。
// 如果图片在数据库中不存在，或者在OSS中不存在且数据库中存在，则返回true。
// 如果在数据库中找到图片但OSS中不存在，则会尝试删除数据库中的记录，并返回true。
// 如果发生错误，则返回true和错误信息。
// 参数:
//   - ctx: 上下文，用于传递请求范围的信息。
//   - imgName: 图片的名称。
//
// 返回值:
//   - bool: 图片是否存在。
//   - error: 错误信息，如果有的话。
func IsExistImgByName(ctx context.Context, imgName string) (bool, error) {
	// 通过图片名称从数据库中查找图片信息。
	imgDto, err := imgrepo.FindImgByName(ctx, imgName)
	if err != nil {
		// 如果查找过程中出现错误，返回false，表示图片不存在
		return false, nil
	}

	// 尝试从OSS中获取图片内容，检查图片是否存在于OSS存储中
	flag, err := storage.Storage.IsExist(ctx, ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType))
	// 如果OSS中图片不存在，但数据库中存在记录，则需要删除数据库中的记录
	if err != nil {
		// 开启数据库事务
		tx := storage.Storage.Db.WithContext(ctx).Begin()

		// 删除数据库中的图片记录
		if delErr := imgrepo.DeleteImgById(tx, imgDto.ImgId); delErr != nil {
			// 如果删除失败，回滚事务
			tx.Rollback()
			return false, delErr
		}

		// 提交事务
		tx.Commit()
		return false, err
	}

	// 返回图片是否存在的标志
	return flag, nil
}

// IsExistImgById 根据图片ID检查图片是否存在于数据库和OSS存储中
// 参数:
//   - ctx: 上下文，用于传递请求范围的信息
//   - imgId: 图片的唯一标识符
//
// 返回值:
//   - bool: 图片是否存在
//   - error: 错误信息，如果有的话
func IsExistImgById(ctx context.Context, imgId string) (bool, error) {
	// 通过图片 ID 从数据库中查找图片信息
	imgDto, err := imgrepo.FindImgById(ctx, imgId)
	if err != nil {
		// 如果查找过程中出现错误，返回false，表示图片不存在
		return false, nil
	}

	// 尝试从OSS中获取图片内容，检查图片是否存在于OSS存储中
	flag, err := storage.Storage.IsExist(ctx, ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType))
	// 如果OSS中图片不存在，但数据库中存在记录，则需要删除数据库中的记录
	if err != nil {
		// 开启数据库事务
		tx := storage.Storage.Db.WithContext(ctx).Begin()

		// 删除数据库中的图片记录
		if delErr := imgrepo.DeleteImgById(tx, imgDto.ImgId); delErr != nil {
			// 如果删除失败，回滚事务
			tx.Rollback()
			return false, delErr
		}

		// 提交事务
		tx.Commit()
		return false, err
	}

	// 返回图片是否存在的标志
	return flag, nil
}

// RenameImgById 根据图片 ID 修改图片名称，包括 OSS 中的文件名和数据库中的记录。
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
//   - imgId: 图片的唯一标识符，用于查询和更新图片信息。
//   - newName: 新的图片名称，用于替换旧的图片名称。
//
// 返回值:
//   - error: 如果操作失败，返回错误信息；如果成功，返回 nil。
func RenameImgById(ctx context.Context, imgId string, newName string) error {
	// 根据图片 ID 查询图片信息，确保图片存在并获取其详细信息
	imgDto, err := imgrepo.FindImgById(ctx, imgId)
	if err != nil {
		return err
	}

	logger.Info("重命名 OSS 中的图片名称")
	// 生成 OSS 中的旧路径和新路径，并调用存储服务重命名 OSS 中的文件
	oldPath := ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)
	newPath := ossstore.GenOssSavePath(newName, imgDto.ImgType)
	if renameErr := storage.Storage.RenameObject(ctx, oldPath, newPath); renameErr != nil {
		return renameErr
	}
	logger.Info("重命名 OSS 中的图片名称成功")

	logger.Info("更新数据库中的图片名称")
	// 开启数据库事务，更新数据库中图片名称，并根据更新结果提交或回滚事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("重命名图片失败: %v", r)
			tx.Rollback()
		}
	}()

	if err = imgrepo.UpdateImgNameById(tx, imgId, newName); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	logger.Info("更新数据库中的图片名称成功")

	logger.Info("删除缓存中保存的预签名 URL")
	if delErr := storage.Storage.Cache.Delete(ctx, storage.BuildImgCacheKey(imgId)); delErr != nil {
		return delErr
	}
	logger.Info("删除缓存中保存的预签名 URL 成功")

	// 生成新的预签名 URL
	presign, err := storage.Storage.GenPreSignUrl(
		ctx,
		ossstore.GenOssSavePath(newName, imgDto.ImgType),
		imgDto.ImgType,
		ossstore.Get,
		35*time.Minute,
	)
	if err != nil {
		return err
	}
	// 缓存新的预签名 URL
	err = storage.Storage.Cache.SetWithExpired(ctx, storage.BuildImgCacheKey(imgId), presign.URL, 30*time.Minute)
	if err != nil {
		logger.Warn("缓存新的预签名 URL 失败")
		return err
	}
	logger.Info("缓存新的预签名 URL 成功")
	logger.Info("完成图片的更新操作")

	return nil
}

// DeleteImg 删除指定 ID 的图片信息及其相关资源。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 图片的唯一标识符，用于定位需要删除的图片。
//
// 返回值:
//   - error: 如果在查找图片信息、删除 OSS 中的图片数据或删除数据库记录时发生错误，则返回相应的错误信息；否则返回 nil。
func DeleteImg(ctx context.Context, id string) error {
	// 查找图片信息，确保图片存在并获取其详细信息
	imgDto, err := imgrepo.FindImgById(ctx, id)
	if err != nil {
		return err
	}

	logger.Info("删除 OSS 中存储的图片文件")
	// 使用图片名称和类型生成存储路径
	if err := storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)); err != nil {
		return err
	}
	logger.Info("删除 OSS 中存储的图片文件成功")

	logger.Info("删除数据库中与图片相关的记录")
	// 开启数据库事务，删除数据库中与图片相关的记录
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除图片失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := imgrepo.DeleteImgById(tx, id); err != nil {
		return err
	}
	tx.Commit()
	logger.Info("删除数据库中与图片相关的记录成功")

	logger.Info("删除缓存中保存的预签名 URL")
	// 删除缓存中保存的预签名 URL
	if err := storage.Storage.Cache.Delete(ctx, storage.BuildImgCacheKey(id)); err != nil {
		return err
	}
	logger.Info("删除缓存中保存的预签名 URL 成功")

	return nil
}
