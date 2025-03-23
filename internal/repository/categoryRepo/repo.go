package categoryRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

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
	err := storage.Storage.Db.WithContext(ctx).Where("c_id = ?", id).First(&category).Error
	if err != nil {
		// 如果查询失败，记录错误日志并返回自定义错误信息。
		msg := fmt.Sprintf("根据分类 ID 查询分类数据失败: %v", err)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 返回找到的分类信息。
	return &category, nil
}
