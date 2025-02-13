package imgInfoRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

// GetBackgroundImg 获取背景图片信息
// - ctx: 上下文对象
// 返回值: 图片信息实体指针和错误信息
func GetBackgroundImg(ctx context.Context) (*po.ImgInfo, error) {
	var img po.ImgInfo
	result := storage.Storage.Db.Model(img).WithContext(ctx).Where("img_name = ?", config.User.BackgroundImage).First(&img)
	if result.Error != nil {
		msg := fmt.Sprintf("查询背景图片数据失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	return &img, nil
}

// FindImgById 根据图片ID查询单条图片信息
// - ctx: 上下文对象
// - imgId: 图片ID
// 返回值: 图片信息实体指针和错误信息
func FindImgById(ctx context.Context, imgId string) (*po.ImgInfo, error) {
	var img po.ImgInfo
	// 使用GORM查询数据库，根据img_id查找单条记录
	result := storage.Storage.Db.Model(img).WithContext(ctx).Where("img_id = ?", imgId).First(&img)
	if result.Error != nil {
		msg := fmt.Sprintf("查询图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	return &img, nil
}

// FindImgsByNameLike 根据图片名称模糊查询图片信息
// - ctx: 上下文对象
// - nameLike: 图片名称模糊匹配字符串
// 返回值: 图片信息实体切片和错误信息
func FindImgsByNameLike(ctx context.Context, nameLike string) ([]po.ImgInfo, error) {
	var images []po.ImgInfo
	// 使用GORM进行模糊查询，查找img_name包含指定字符串的记录
	result := storage.Storage.Db.Model(&po.ImgInfo{}).WithContext(ctx).Where("img_name LIKE ?", "%"+nameLike+"%").Find(&images)
	if result.Error != nil {
		msg := fmt.Sprintf("模糊查询图片信息失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	return images, nil
}

// AddImgInfo 创建新的图片信息记录
// - ctx: 上下文对象
// - img: 图片信息实体指针
// 返回值: 受影响的行数和错误信息
func AddImgInfo(ctx context.Context, img *po.ImgInfo) (int64, error) {
	tx := storage.Storage.Db.Model(img).WithContext(ctx).Begin()
	// 使用defer确保在panic时回滚事务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("创建图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("添加图片信息数据")
	// 执行创建操作
	result := tx.Create(img)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("创建图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("创建图片信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// AddImgInfoBatch 批量添加图片信息记录
// - ctx: 上下文对象
// - imgs: 图片信息实体切片
//
// 返回值: 受影响的行数和错误信息
// - int64: 受影响的行数
// - error: 错误信息
func AddImgInfoBatch(ctx context.Context, imgs []po.ImgInfo) (int64, error) {
	if len(imgs) == 0 {
		return 0, nil
	}

	tx := storage.Storage.Db.Model(&po.ImgInfo{}).WithContext(ctx).Begin()
	// 使用defer确保在panic时回滚事务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("批量添加图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("批量添加图片信息数据")
	// 执行创建操作
	result := tx.Create(imgs)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("批量添加图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("批量添加图片信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// DeleteImgInfoBatch 批量删除图片信息记录
// - ctx: 上下文对象
// - ids: 图片信息ID切片
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func DeleteImgInfoBatch(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	tx := storage.Storage.Db.Model(&po.ImgInfo{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("批量删除图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("批量删除图片信息数据")

	// 将ID转换为ImgInfo对象
	var imgs []po.ImgInfo
	for _, id := range ids {
		imgs = append(imgs, po.ImgInfo{ImgId: id})
	}

	// 执行删除操作
	result := tx.Delete(imgs)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("批量删除图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 提交事务
	tx.Commit()
	logger.Info("批量删除图片信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// UpdateImgNameById 更新图片信息记录的名称
// - ctx: 上下文对象
// - id: 图片信息ID
// - name: 新的名称
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func UpdateImgNameById(ctx context.Context, id string, name string) (int64, error) {
	tx := storage.Storage.Db.Model(&po.ImgInfo{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("更新图片信息数据")
	result := tx.Model(&po.ImgInfo{}).Where("img_id = ?", id).Update("img_name", name)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("更新图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 提交事务
	tx.Commit()
	logger.Info("更新图片信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}
