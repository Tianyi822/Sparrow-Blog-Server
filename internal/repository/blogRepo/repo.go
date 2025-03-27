package blogRepo

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

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

// UpdateBlog 更新博客信息。
// 该函数接收一个上下文和一个博客DTO对象，用于更新数据库中的博客信息。
// 参数:
//   - ctx: 上下文，用于数据库操作的上下文管理。
//   - blogDto: 包含要更新的博客信息的DTO对象。
//
// 返回值:
//   - 如果更新过程中发生错误，则返回错误。
func UpdateBlog(ctx context.Context, blogDto *dto.BlogDto) error {
	// 开始一个数据库事务。
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	// 使用defer和recover来处理可能的panic，确保在出现panic时回滚事务。
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新博客数据失败: %v", r)
			tx.Rollback()
		}
	}()

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
	// 提交事务。
	tx.Commit()

	return nil
}

// DeleteBlogById 根据博客ID删除对应的博客数据。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlogById(ctx context.Context, id string) error {
	// 开启数据库事务，并绑定上下文以确保事务与请求生命周期一致
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 捕获可能的 panic，记录错误日志并回滚事务，避免事务未正确关闭。
		if r := recover(); r != nil {
			logger.Error("删除博客数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("删除 ID 为 %v 的博客记录", id)
	// 根据博客ID删除对应的博客记录
	if err := tx.Where("b_id = ?", id).Delete(&po.Blog{}).Error; err != nil {
		// 如果删除操作失败，回滚事务并记录错误日志。
		tx.Rollback()
		msg := fmt.Sprintf("删除博客数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 提交事务，确保事务被正确关闭
	tx.Commit()
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
func ChangeBlogStateById(ctx context.Context, id string) error {
	// 开启数据库事务，并在函数结束时根据情况提交或回滚。
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			// 捕获 panic 并记录错误日志，同时回滚事务。
			logger.Error("设置博客状态失败: %v", r)
			tx.Rollback()
		}
	}()

	// 查询指定 ID 的博客状态。
	var blog po.Blog
	if err := tx.Select("b_id", "b_state").Where("b_id = ?", id).Find(&blog).Error; err != nil {
		// 如果查询失败，记录错误日志并回滚事务。
		tx.Rollback()
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
	if err := tx.Model(blog).Update("b_state", blog.BlogState).Where("b_id = ?", id).Error; err != nil {
		// 如果更新失败，记录错误日志并回滚事务。
		tx.Rollback()
		msg := fmt.Sprintf("设置博客状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 提交事务以保存更改。
	tx.Commit()

	return nil
}

// SetTopById 通过博客的ID来设置博客的置顶状态
// 参数:
//
//	ctx - 上下文，用于数据库操作的上下文管理
//	id - 博客的唯一标识符
//
// 返回值:
//
//	如果操作成功，则返回nil；如果操作失败，则返回错误
func SetTopById(ctx context.Context, id string) error {
	// 开始数据库事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	// 使用defer和recover来处理可能的panic，确保事务的完整性
	defer func() {
		if r := recover(); r != nil {
			logger.Error("设置博客置顶失败: %v", r)
			tx.Rollback()
		}
	}()

	// 查询指定 ID 的播客置顶状态
	var blog po.Blog
	if err := tx.Select("b_id", "is_top").Where("b_id = ?", id).Find(&blog).Error; err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("获取博客置顶状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 根据当前的置顶状态来更新博客的置顶状态
	if blog.BlogIsTop {
		logger.Info("取消 ID 为 %v 的播客置顶", id)
		blog.BlogIsTop = false
	} else {
		logger.Info("设置 ID 为 %v 的播客置顶", id)
		blog.BlogIsTop = true
	}

	// 更新数据库中的博客置顶状态
	if err := tx.Model(blog).Update("is_top", blog.BlogIsTop).Where("b_id = ?", id).Error; err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("设置博客置顶失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	// 提交事务
	tx.Commit()

	return nil
}
