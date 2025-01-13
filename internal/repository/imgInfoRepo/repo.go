package imgInfoRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog/internal/model/po"
	"h2blog/pkg/logger"
	"h2blog/storage"
)

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

// FindImgByNameLike 根据图片名称模糊查询图片信息
// - ctx: 上下文对象
// - nameLike: 图片名称模糊匹配字符串
// 返回值: 图片信息实体切片和错误信息
func FindImgByNameLike(ctx context.Context, nameLike string) ([]po.ImgInfo, error) {
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

// CreateImgInfo 创建新的图片信息记录
// - ctx: 上下文对象
// - img: 图片信息实体指针
// 返回值: 受影响的行数和错误信息
func CreateImgInfo(ctx context.Context, img *po.ImgInfo) (int64, error) {
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
