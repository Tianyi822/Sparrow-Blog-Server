package categoryRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/utils"
	"h2blog_server/storage"
)

// AddCategory 添加一个新的播客分类到数据库中。
// 参数：
// - ctx context.Context: 上下文对象，用于控制请求的生命周期和传递元数据。
// - cateDto *dto.CategoryDto: 包含分类信息的数据传输对象，其中 CategoryName 是分类名称。
//
// 返回值：
// - error: 如果操作成功，返回 nil；如果发生错误，返回包含错误信息的 error 对象。
func AddCategory(ctx context.Context, cateDto *dto.CategoryDto) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 捕获可能的 panic，记录错误日志并回滚事务，避免事务未正确关闭。
		if r := recover(); r != nil {
			logger.Error("添加分类失败: %v", r)
			tx.Rollback()
		}
	}()

	// 根据分类名称生成唯一的分类 ID。如果生成失败，记录警告日志并返回错误。
	cId, err := utils.GenId(cateDto.CategoryName)
	if err != nil {
		msg := fmt.Sprintf("根据分类名称生成分类 ID 失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	cateDto.CategoryId = cId

	logger.Info("创建播客分类")
	// 将生成的分类信息保存到数据库中。如果保存失败，记录警告日志并返回错误。
	if err = tx.WithContext(ctx).Create(&po.Category{
		CategoryId:   cId,
		CategoryName: cateDto.CategoryName,
	}).Error; err != nil {
		msg := fmt.Sprintf("创建分类失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("创建分类成功")
	tx.Commit()

	// 如果所有操作成功，返回 nil 表示没有错误。
	return nil
}

// FindCategoryById 根据分类ID查找分类信息。
// 参数:
//
//	ctx context.Context: 上下文对象，用于取消请求和传递请求级值。
//	id string: 分类的唯一标识符。
//
// 返回值:
//
//	*po.Category: 找到的分类指针，如果未找到则返回nil。
//	error: 如果发生错误，则返回错误信息。
func FindCategoryById(ctx context.Context, id string) (*po.Category, error) {
	// 初始化一个Category类型变量以存储查询结果。
	var category po.Category

	// 使用给定的ID查询数据库中的分类信息。
	err := storage.Storage.Db.WithContext(ctx).Where("category_id = ?", id).First(&category).Error
	if err != nil {
		// 如果查询失败，记录错误日志并返回自定义错误信息。
		msg := fmt.Sprintf("根据分类 ID 查询分类数据失败: %v", err)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 返回找到的分类信息。
	return &category, nil
}
