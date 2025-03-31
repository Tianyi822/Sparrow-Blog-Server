package blogRepo

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

// FindBlogTitleById 根据博客ID查询博客的标题信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 博客的唯一标识符，用于查询特定博客的信息。
//
// 返回值:
//   - *dto.BlogDto: 包含博客ID和标题的数据传输对象，如果查询失败则返回nil。
//   - error: 查询过程中发生的错误信息，如果没有错误则返回nil。
func FindBlogTitleById(ctx context.Context, id string) (string, error) {
	blog := &po.Blog{}

	if err := storage.Storage.Db.WithContext(ctx).Model(&po.Blog{}).
		Where("blog_id = ?", id).
		Select("blog_id", "blog_title").
		Find(&blog).Error; err != nil {
		msg := fmt.Sprintf("查询博客信息失败: %v", err)
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	return blog.BlogTitle, nil
}

// FindAllBlogs 查询所有博客信息，并根据需要返回简要信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - needBrief: 布尔值，指示是否需要包含博客的简要信息。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客数据的DTO列表，每个DTO表示一个博客的基本信息。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回nil。
func FindAllBlogs(ctx context.Context, needBrief bool) ([]*dto.BlogDto, error) {
	blogs := make([]*po.Blog, 0)

	// 使用GORM查询博客数据，按置顶优先级和创建时间排序。
	var result *gorm.DB

	// 根据需要返回简要信息，使用 Select 函数来选择特定的列。
	if needBrief {
		result = storage.Storage.Db.Model(&po.Blog{}).
			Select(
				"blog_id",
				"blog_title",
				"blog_brief",
				"category_id",
				"blog_state",
				"blog_words_num",
				"blog_is_top",
				"create_time",
				"update_time",
			).
			WithContext(ctx).
			Order("blog_is_top DESC").
			Order("create_time DESC").
			Find(&blogs)
	} else {
		result = storage.Storage.Db.Model(&po.Blog{}).
			Select(
				"blog_id",
				"blog_title",
				"category_id",
				"blog_state",
				"blog_words_num",
				"blog_is_top",
				"create_time",
				"update_time",
			).
			WithContext(ctx).
			Order("blog_is_top DESC").
			Order("create_time DESC").
			Find(&blogs)
	}

	// 如果查询过程中发生错误，记录错误日志并返回错误信息。
	if result.Error != nil {
		msg := fmt.Sprintf("查询博客信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 初始化博客DTO列表，用于存储转换后的博客数据。
	blogDtos := make([]*dto.BlogDto, 0)

	// 遍历查询到的博客数据，将其转换为DTO格式并添加到结果列表中。
	for _, blog := range blogs {
		blogDto := &dto.BlogDto{
			BlogId:       blog.BlogId,
			BlogTitle:    blog.BlogTitle,
			BlogIsTop:    blog.BlogIsTop,
			BlogState:    blog.BlogState,
			BlogWordsNum: blog.BlogWordsNum,
			CategoryId:   blog.CategoryId,
			CreateTime:   blog.CreateTime,
			UpdateTime:   blog.UpdateTime,
		}
		if needBrief {
			blogDto.BlogBrief = blog.BlogBrief
		}
		blogDtos = append(blogDtos, blogDto)
	}

	// 返回转换后的博客DTO列表和nil错误。
	return blogDtos, nil
}

// CalBlogsCountByCategoryId 根据分类 ID 查询博客数量。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - categoryId: 分类 ID，用于筛选特定分类下的博客。
//
// 返回值:
//   - int64: 符合条件的博客数量。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func CalBlogsCountByCategoryId(ctx context.Context, categoryId string) (int64, error) {
	var num int64

	// 使用 GORM 查询数据库，统计指定分类 ID 下的博客数量。
	result := storage.Storage.Db.WithContext(ctx).Model(&po.Blog{}).Where("category_id = ?", categoryId).Count(&num)
	if result.Error != nil {
		// 如果查询失败，记录错误日志并返回错误信息。
		msg := fmt.Sprintf("根据分类 ID 查询博客数量失败: %v", result.Error)
		logger.Warn(msg)
		return 0, errors.New(msg)
	}

	// 返回查询到的博客数量和 nil 错误。
	return num, nil
}

// GetCategoryIdByBlogId 根据博客 ID 查询对应的分类 ID。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - blogId: 博客的唯一标识符，用于查询对应的分类 ID。
//
// 返回值:
//   - string: 查询到的分类 ID。如果未找到或发生错误，则返回空字符串。
//   - error: 如果查询过程中发生错误，则返回包含错误信息的 error 对象；否则返回 nil。
func GetCategoryIdByBlogId(ctx context.Context, blogId string) (string, error) {
	var categoryId string

	// 使用数据库查询，根据博客 ID 获取对应的分类 ID。
	result := storage.Storage.Db.WithContext(ctx).Model(&po.Blog{}).Select("category_id").Where("blog_id = ?", blogId).Find(&categoryId)
	if result.Error != nil {
		msg := fmt.Sprintf("根据博客 ID 查询分类 ID 失败: %v", result.Error)
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	return categoryId, nil
}

// AddBlog 创建一篇新的博客并将其存储到数据库中。
// 参数:
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - blogDto: 包含博客相关信息的数据传输对象，包括标题、简介、分类ID等。
//
// 返回值:
//   - error: 如果操作成功则返回nil，否则返回包含错误信息的error对象。
func AddBlog(tx *gorm.DB, blogDto *dto.BlogDto) error {
	// 生成博客的唯一ID，如果生成失败则记录警告日志并返回错误。
	bid, err := utils.GenId(blogDto.BlogTitle)
	if err != nil {
		msg := fmt.Sprintf("生成博客ID失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	// 将 id 保存在 dto 中
	blogDto.BlogId = bid

	logger.Info("创建博客")
	// 将博客信息插入数据库，如果插入失败则记录警告日志并返回错误。
	if err := tx.Create(&po.Blog{
		BlogId:       bid,
		BlogTitle:    blogDto.BlogTitle,
		BlogBrief:    blogDto.BlogBrief,
		CategoryId:   blogDto.CategoryId,
		BlogState:    blogDto.BlogState,
		BlogWordsNum: blogDto.BlogWordsNum,
		BlogIsTop:    blogDto.BlogIsTop,
	}).Error; err != nil {
		msg := fmt.Sprintf("创建博客失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	logger.Info("创建博客成功")

	return nil
}

// UpdateBlog 更新博客信息。
// 该函数接收一个上下文和一个博客DTO对象，用于更新数据库中的博客信息。
// 参数:
//   - tx: 数据库事务对象
//   - blogDto: 包含要更新的博客信息的DTO对象。
//
// 返回值:
//   - 如果更新过程中发生错误，则返回错误。
func UpdateBlog(tx *gorm.DB, blogDto *dto.BlogDto) error {
	logger.Info("开始更新播客数据")
	// 更新博客信息。
	if err := tx.Model(&po.Blog{}).Where("b_id = ?", blogDto.BlogId).Updates(po.Blog{
		BlogBrief:    blogDto.BlogBrief,
		CategoryId:   blogDto.CategoryId,
		BlogTitle:    blogDto.BlogTitle,
		BlogIsTop:    blogDto.BlogIsTop,
		BlogState:    blogDto.BlogState,
		BlogWordsNum: blogDto.BlogWordsNum,
	}).Error; err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("更新博客数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("更新播客数据成功")

	return nil
}

// DeleteBlogById 根据博客ID删除对应的博客数据。
// 参数:
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlogById(tx *gorm.DB, id string) error {
	logger.Info("删除 ID 为 %v 的博客记录", id)
	// 根据博客ID删除对应的博客记录
	if err := tx.Where("blog_id = ?", id).Delete(&po.Blog{}).Error; err != nil {
		// 如果删除操作失败，回滚事务并记录错误日志。
		tx.Rollback()
		msg := fmt.Sprintf("删除博客数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("成功删除 ID 为 %v 的博客记录", id)

	// 如果删除成功，事务会自动提交（defer 中不会触发 Rollback）。
	return nil
}

// ChangeBlogStateById 根据博客 ID 设置其状态（显示或隐藏）。
// 参数:
//   - ctx context.Context: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id string: 博客的唯一标识符。
//
// 返回值:
//   - error: 如果操作失败，返回错误信息；如果成功，返回 nil。
func ChangeBlogStateById(tx *gorm.DB, id string) error {
	// 查询指定 ID 的博客状态。
	var blog po.Blog
	if err := tx.Select("blog_id", "blog_state").Where("blog_id = ?", id).Find(&blog).Error; err != nil {
		msg := fmt.Sprintf("获取博客状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 根据当前状态切换博客的显示/隐藏状态，并记录操作日志。
	if blog.BlogState {
		logger.Info("设置 ID 为 %v 的博客状态为隐藏", id)
		blog.BlogState = false
	} else {
		logger.Info("设置 ID 为 %v 的博客状态为显示", id)
		blog.BlogState = true
	}

	// 更新博客状态到数据库。
	if err := tx.Model(blog).Update("blog_state", blog.BlogState).Where("blog_id = ?", id).Error; err != nil {
		msg := fmt.Sprintf("设置博客状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("设置 ID 为 %v 的博客状态成功", id)

	return nil
}

// SetTopById 根据博客 ID 设置或取消博客的置顶状态。
// 参数：
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - id: 博客的唯一标识符，用于定位需要操作的博客。
//
// 返回值：
//   - error: 如果操作过程中发生错误，则返回错误信息；否则返回 nil。
func SetTopById(tx *gorm.DB, id string) error {
	// 查询指定 ID 的播客置顶状态
	var blog po.Blog
	if err := tx.Select("blog_id", "blog_is_top").Where("blog_id = ?", id).Find(&blog).Error; err != nil {
		msg := fmt.Sprintf("获取博客置顶状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 根据当前的置顶状态来更新博客的置顶状态
	if blog.BlogIsTop {
		logger.Info("取消 ID 为 %v 的博客置顶", id)
		blog.BlogIsTop = false
	} else {
		logger.Info("设置 ID 为 %v 的博客置顶", id)
		blog.BlogIsTop = true
	}

	// 更新数据库中的博客置顶状态
	if err := tx.Model(blog).Update("blog_is_top", blog.BlogIsTop).Where("blog_id = ?", id).Error; err != nil {
		msg := fmt.Sprintf("设置博客置顶失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("修改 ID 为 %v 的博客是否置顶成功", id)

	return nil
}
