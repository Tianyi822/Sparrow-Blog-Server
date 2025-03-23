package tagRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

// FindTagsByBlogId 根据博客 ID 查找所有关联的标签。
// 该函数首先查询中间表 BlogTag 以获取标签 ID，然后根据这些 ID 查询标签表以获取标签详细信息。
// 参数:
//
//	ctx - 上下文，用于处理请求和响应的生命周期管理。
//	blogId - 博客的唯一标识符。
//
// 返回值:
//
//	[]*po.Tag - 与博客关联的标签列表。
//	error - 如果查询过程中发生错误，则返回该错误。
func FindTagsByBlogId(ctx context.Context, blogId string) ([]*po.Tag, error) {
	// 查询中间表 BlogTag，获取与博客 ID 关联的标签 ID。
	var bt []po.BlogTag

	result := storage.Storage.Db.WithContext(ctx).Model(&po.BlogTag{}).Where("b_id = ?", blogId).Find(&bt)

	// 检查查询过程中是否有错误发生。
	if result.Error != nil {
		msg := fmt.Sprintf("根据博客 ID 查询标签数据失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 提取查询结果中的标签 ID。
	var tagIds []string
	for _, tag := range bt {
		tagIds = append(tagIds, tag.TId)
	}

	// 根据提取的标签 ID 查询标签表，获取标签详细信息。
	var tags []*po.Tag
	result = storage.Storage.Db.WithContext(ctx).Model(&po.Tag{}).Where("t_id IN ?", tagIds).Find(&tags)

	// 再次检查查询过程中是否有错误发生。
	if result.Error != nil {
		msg := fmt.Sprintf("根据标签 ID 查询标签数据失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 返回查询到的标签列表。
	return tags, nil
}
