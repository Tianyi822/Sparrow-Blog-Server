package categoryrepo

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
)

// FindAllCategories 获取所有分类数据
// 参数:
//   - ctx context.Context: 上下文，用于控制请求的生命周期和取消
//
// 返回值:
//   - []*dto.CategoryDto: 分类数据列表
//   - error: 错误信息，若查询失败则返回具体错误
func FindAllCategories(ctx context.Context) ([]*dto.CategoryDto, error) {
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

// FindCategoryById 根据分类ID查找分类信息。
// 参数:
//   - ctx context.Context: 上下文对象，用于取消请求和传递请求级值。
//   - id string: 分类的唯一标识符。
//
// 返回值:
//   - *po.Category: 找到的分类指针，如果未找到则返回nil。
//   - error: 如果发生错误，则返回错误信息。
func FindCategoryById(ctx context.Context, id string) (*dto.CategoryDto, error) {
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
	return &dto.CategoryDto{
		CategoryId:   category.CategoryId,
		CategoryName: category.CategoryName,
	}, nil
}

// AddCategory 添加一个新的播客分类到数据库中。
// 参数：
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - cateDto *dto.CategoryDto: 包含分类信息的数据传输对象，其中 CategoryName 是分类名称。
//
// 返回值：
//   - error: 如果操作成功，返回 nil；如果发生错误，返回包含错误信息的 error 对象。
func AddCategory(tx *gorm.DB, cateDto *dto.CategoryDto) error {
	if len(cateDto.CategoryName) == 0 {
		msg := "分类名称不能为空"
		logger.Warn(msg)
		return errors.New(msg)
	}

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
	if err = tx.Create(&po.Category{
		CategoryId:   cId,
		CategoryName: cateDto.CategoryName,
	}).Error; err != nil {
		msg := fmt.Sprintf("创建分类失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("创建分类成功")

	// 如果所有操作成功，返回 nil 表示没有错误。
	return nil
}

// CleanCategoriesWithoutBlog 删除没有博客关联的分类
// 参数:
//   - tx *gorm.DB: gorm数据库连接对象，用于执行数据库操作
//
// 返回值:
//   - error: 如果执行数据库操作时发生错误，则返回错误对象；否则返回nil
func CleanCategoriesWithoutBlog(tx *gorm.DB) error {
	// 删除没有博客关联的分类
	result := tx.Where("category_id NOT IN (SELECT category_id FROM H2_Blog)").Delete(&po.Category{})
	if result.Error != nil {
		// 如果删除操作失败，记录错误日志并返回新的错误对象
		msg := fmt.Sprintf("删除没有博客关联的分类失败: %v", result.Error)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("删除没有博客关联的分类成功, 共删除 %v 条数据", result.RowsAffected)
	// 如果删除操作成功，返回nil
	return nil
}
