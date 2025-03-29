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

// GetAllCategories 获取所有分类数据
// 参数:
//   - ctx context.Context: 上下文，用于控制请求的生命周期和取消
//
// 返回值:
//   - []*dto.CategoryDto: 分类数据列表
//   - error: 错误信息，若查询失败则返回具体错误
func GetAllCategories(ctx context.Context) ([]*dto.CategoryDto, error) {
	// 执行数据库查询以获取所有分类数据，并处理可能的错误
	var categories []*po.Category
	if err := storage.Storage.Db.WithContext(ctx).Find(&categories).Error; err != nil {
		msg := fmt.Sprintf("查询分类数据失败: %v", err)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 将数据库实体转换为DTO结构
	var cateDtos []*dto.CategoryDto
	for _, c := range categories {
		cateDtos = append(cateDtos, &dto.CategoryDto{
			CategoryId:   c.CategoryId,
			CategoryName: c.CategoryName,
		})
	}

	return cateDtos, nil
}

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

// DeleteCategoryById 根据分类 ID 删除对应的分类记录。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 分类的唯一标识符，用于定位需要删除的分类记录。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteCategoryById(ctx context.Context, id string) error {
	// 开启数据库事务，并将上下文绑定到事务中。
	tx := storage.Storage.Db.WithContext(ctx).Begin()

	defer func() {
		// 捕获可能的 panic，记录错误日志并回滚事务，避免事务未正确关闭。
		if r := recover(); r != nil {
			logger.Error("删除分类失败: %v", r)
			tx.Rollback()
		}
	}()

	// 记录删除操作的日志，标明正在删除的分类 ID。
	logger.Info("删除 ID 为 %v 的分类记录", id)

	// 执行删除操作，根据分类 ID 删除对应的分类记录。
	if err := tx.Where("category_id = ?", id).Delete(&po.Category{}).Error; err != nil {
		// 如果删除操作失败，回滚事务并记录错误日志。
		tx.Rollback()
		msg := fmt.Sprintf("删除分类数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 提交事务，确保删除操作生效。
	tx.Commit()

	// 记录删除成功的日志，标明已删除的分类 ID。
	logger.Info("成功删除 ID 为 %v 的分类记录", id)

	return nil
}
