package tagRepo

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/utils"
	"h2blog_server/storage"
)

// GetAllTags 查询数据库中的所有标签，并将其转换为 DTO（数据传输对象）格式返回。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.TagDto: 包含所有标签的 DTO 列表，每个 DTO 包含标签的 ID 和名称。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func GetAllTags(ctx context.Context) ([]*dto.TagDto, error) {
	// 创建一个空的标签列表，用于存储从数据库中查询到的标签数据。
	var tags []*po.Tag

	// 使用 GORM 查询数据库，获取所有标签数据。
	result := storage.Storage.Db.WithContext(ctx).Model(&po.Tag{}).Find(&tags)
	if result.Error != nil {
		msg := fmt.Sprintf("查询标签数据失败: %v", result.Error)
		logger.Warn(msg)
	}

	// 将查询到的标签数据转换为 DTO 格式，便于后续处理或返回给调用方。
	var tagDtos []*dto.TagDto
	for _, tag := range tags {
		tagDtos = append(tagDtos, &dto.TagDto{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		})
	}

	// 返回转换后的 DTO 列表和 nil 错误。
	return tagDtos, nil
}

// FindTagsByBlogId 根据博客 ID 查找所有关联的标签。
// 该函数首先查询中间表 BlogTag 以获取标签 ID，然后根据这些 ID 查询标签表以获取标签详细信息。
// 参数:
//
//	ctx - 上下文，用于处理请求和响应的生命周期管理。
//	blogId - 博客的唯一标识符。
//
// 返回值:
//
//	[]dto.TagDto - 与博客关联的标签列表。
//	error - 如果查询过程中发生错误，则返回该错误。
func FindTagsByBlogId(ctx context.Context, blogId string) ([]dto.TagDto, error) {
	// 查询中间表 BlogTag，获取与博客 ID 关联的标签 ID。
	var bt []po.BlogTag

	result := storage.Storage.Db.WithContext(ctx).Model(&po.BlogTag{}).Where("blog_id = ?", blogId).Find(&bt)

	// 检查查询过程中是否有错误发生。
	if result.Error != nil {
		msg := fmt.Sprintf("根据博客 ID 查询标签数据失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 提取查询结果中的标签 ID。
	var tagIds []string
	for _, tag := range bt {
		tagIds = append(tagIds, tag.TagId)
	}

	// 根据提取的标签 ID 查询标签表，获取标签详细信息。
	var tags []*po.Tag
	result = storage.Storage.Db.WithContext(ctx).Model(&po.Tag{}).Where("tag_id IN ?", tagIds).Find(&tags)

	// 再次检查查询过程中是否有错误发生。
	if result.Error != nil {
		msg := fmt.Sprintf("根据标签 ID 查询标签数据失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	tagsDto := make([]dto.TagDto, len(tags))
	for i, tag := range tags {
		tagsDto[i] = dto.TagDto{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		}
	}

	// 返回查询到的标签列表。
	return tagsDto, nil
}

// CalBlogsCountByTagId 根据标签 ID 查询关联的博客数量。
// 参数:
//   - ctx context.Context: 上下文对象，用于控制请求生命周期和传递上下文信息。
//   - tagId string: 标签的唯一标识符，用于查询与该标签关联的博客数量。
//
// 返回值:
//   - int64: 与指定标签 ID 关联的博客数量。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func CalBlogsCountByTagId(ctx context.Context, tagId string) (int64, error) {
	var count int64

	// 使用 GORM 查询数据库，统计与指定标签 ID 关联的博客数量。
	result := storage.Storage.Db.WithContext(ctx).Model(&po.BlogTag{}).Where("tag_id = ?", tagId).Count(&count)

	// 如果查询过程中发生错误，记录警告日志并返回错误信息。
	if result.Error != nil {
		msg := fmt.Sprintf("根据标签 ID 查询博客数量失败: %v", result.Error)
		logger.Warn(msg)
		return 0, errors.New(msg)
	}

	// 返回查询到的博客数量和 nil 错误。
	return count, nil
}

// AddTags 批量添加标签到数据库。
// 参数:
// - tx: 数据库事务对象，用于执行数据库操作。
// - tags: 包含标签信息的 DTO 列表，每个 DTO 包含标签名称 (TagName)。
//
// 返回值:
// - error: 如果操作成功则返回 nil，否则返回包含错误信息的 error 对象。
func AddTags(tx *gorm.DB, tags []dto.TagDto) ([]dto.TagDto, error) {
	var tagPos []po.Tag
	for index, tag := range tags {
		if len(tag.TagName) == 0 {
			msg := fmt.Sprintf("标签名称不能为空")
			logger.Warn(msg)
			return nil, errors.New(msg)
		}

		// 为每个标签生成唯一 ID
		tId, err := utils.GenId(tag.TagName)
		if err != nil {
			msg := fmt.Sprintf("生成标签 ID 失败: %v", err)
			logger.Warn(msg)
			return nil, errors.New(msg)
		}
		tags[index].TagId = tId

		// 将生成的标签信息转换为持久化对象并存储到切片中
		tagPos = append(tagPos, po.Tag{
			TagId:   tId,
			TagName: tag.TagName,
		})
	}

	// 记录日志并尝试批量保存标签数据到数据库
	logger.Info("批量保存标签数据")
	if err := tx.Create(&tagPos).Error; err != nil {
		msg := fmt.Sprintf("创建标签数据失败: %v", err)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}
	logger.Info("批量保存标签数据成功")

	return tags, nil
}

// DeleteTags 批量删除标签
// 该函数接收一个上下文和一个标签DTO列表，尝试从数据库中删除这些标签
// 参数:
//   - tx: 事物对象，用于控制数据事务的生命周期
//   - tags []dto.TagDto: 待删除的标签DTO列表
//
// 返回值:
//   - error: 如果删除操作失败，则返回错误信息；否则返回nil
func DeleteTags(tx *gorm.DB, tags []dto.TagDto) error {
	// 提取标签ID
	var tagIds []string
	for _, tag := range tags {
		tagIds = append(tagIds, tag.TagId)
	}

	logger.Info("批量删除标签数据")
	// 批量删除标签数据
	result := tx.Where("tag_id IN ?", tagIds).Delete(&po.Tag{})
	if result.Error != nil {
		msg := fmt.Sprintf("删除标签数据失败: %v", result.Error)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("批量删除标签数据成功")

	// 返回无错误
	return nil
}

// DeleteBlogTagAssociationByBlogId 根据博客ID删除博客与标签的关联数据。
// 参数:
//   - tx: *gorm.DB，数据库事务对象，用于执行删除操作。
//   - blogId: string，博客的唯一标识符，用于定位需要删除的关联数据。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回包含错误信息的error对象；否则返回nil。
func DeleteBlogTagAssociationByBlogId(tx *gorm.DB, blogId string) error {
	logger.Info("删除博客标签关联数据")

	// 使用GORM的Where方法根据blogId条件删除关联数据。
	result := tx.Where("blog_id = ?", blogId).Delete(&po.BlogTag{})
	if result.Error != nil {
		// 如果删除操作失败，记录警告日志并返回错误信息。
		msg := fmt.Sprintf("删除博客标签关联数据失败: %v", result.Error)
		logger.Warn(msg)
		return errors.New(msg)
	}

	logger.Info("删除博客标签关联数据成功")

	return nil
}

// AddBlogTagAssociation 用于将博客与标签的关联数据批量保存到数据库中。
// 参数说明：
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - blogId: 博客的唯一标识符，用于关联博客和标签。
//   - tags: 标签的 DTO（数据传输对象）切片，包含需要关联的标签信息。
//
// 返回值：
//   - error: 如果操作成功，则返回 nil；如果发生错误，则返回具体的错误信息。
func AddBlogTagAssociation(tx *gorm.DB, blogId string, tags []dto.TagDto) error {
	var blogTags []po.BlogTag
	for _, tag := range tags {
		// 将标签 DTO 转换为 BlogTag 模型对象，并构建批量插入的数据。
		blogTags = append(blogTags, po.BlogTag{
			BlogId: blogId,
			TagId:  tag.TagId,
		})
	}

	logger.Info("批量保存博客标签关联数据")
	// 批量插入博客标签关联数据到数据库。
	if err := tx.Create(&blogTags).Error; err != nil {
		msg := fmt.Sprintf("创建博客标签关联数据失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("批量保存博客标签关联数据成功")

	return nil
}

// UpdateBlogTagAssociation 更新博客与标签的关联关系。
// 参数：
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - blogId string: 博客的唯一标识符，用于定位需要更新标签关联的博客。
//   - newTags []dto.TagDto: 新的标签列表，表示需要与博客建立关联的标签数据。
//
// 返回值：
//   - error: 如果操作失败，返回错误信息；如果成功，返回 nil。
func UpdateBlogTagAssociation(tx *gorm.DB, blogId string, newTags []dto.TagDto) error {
	// 删除指定博客的所有旧标签关联数据。
	if err := tx.Delete(&po.BlogTag{}).Where("blog_id = ?", blogId).Error; err != nil {
		msg := fmt.Sprintf("删除博客标签关联数据失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}

	// 添加新的博客标签关联数据。
	if err := AddBlogTagAssociation(tx, blogId, newTags); err != nil {
		msg := fmt.Sprintf("更新博客标签关联数据失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}

	// 如果所有操作成功，返回 nil 表示成功。
	return nil
}
