package tagRepo

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

// AddTags 批量添加标签到数据库。
// 参数:
// - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
// - tags: 包含标签信息的 DTO 列表，每个 DTO 包含标签名称 (TName)。
//
// 返回值:
// - error: 如果操作成功则返回 nil，否则返回包含错误信息的 error 对象。
func AddTags(ctx context.Context, tags []dto.TagDto) ([]dto.TagDto, error) {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var tagPos []po.Tag
	var newTags []dto.TagDto
	for _, tag := range tags {
		// 为每个标签生成唯一 ID
		tId, err := utils.GenId(tag.TName)
		if err != nil {
			msg := fmt.Sprintf("生成标签 ID 失败: %v", err)
			logger.Warn(msg)
			return nil, errors.New(msg)
		}

		// 将生成的标签信息转换为持久化对象并存储到切片中
		tagPos = append(tagPos, po.Tag{
			TId:   tId,
			TName: tag.TName,
		})

		// 将生成的标签信息转换为 DTO 并添加到新标签列表中
		newTags = append(newTags, dto.TagDto{
			TId:   tId,
			TName: tag.TName,
		})
	}

	// 记录日志并尝试批量保存标签数据到数据库
	logger.Info("批量保存标签数据")
	if err := tx.Create(&tagPos).Error; err != nil {
		msg := fmt.Sprintf("创建标签数据失败: %v", err)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}
	// 提交事务以完成数据库写入
	tx.Commit()

	logger.Info("批量保存标签数据成功")

	return newTags, nil
}

// AddBlogTagAssociation 用于将博客与标签的关联数据批量保存到数据库中。
// 参数说明：
// - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
// - blogId: 博客的唯一标识符，用于关联博客和标签。
// - tags: 标签的 DTO（数据传输对象）切片，包含需要关联的标签信息。
//
// 返回值：
// - error: 如果操作成功，则返回 nil；如果发生错误，则返回具体的错误信息。
func AddBlogTagAssociation(ctx context.Context, blogId string, tags []dto.TagDto) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var blogTags []po.BlogTag
	for _, tag := range tags {
		// 将标签 DTO 转换为 BlogTag 模型对象，并构建批量插入的数据。
		blogTags = append(blogTags, po.BlogTag{
			BId: blogId,
			TId: tag.TId,
		})
	}

	logger.Info("批量保存博客标签关联数据")
	// 批量插入博客标签关联数据到数据库。
	if err := tx.Create(&blogTags).Error; err != nil {
		msg := fmt.Sprintf("创建博客标签关联数据失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	// 提交事物
	tx.Commit()
	logger.Info("批量保存博客标签关联数据成功")

	return nil
}
