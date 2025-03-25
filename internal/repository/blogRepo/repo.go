package blogRepo

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage"
)

// FindBlogsInPage 查询指定页码的博客列表，并返回分页后的博客数据。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和取消操作。
//   - page: 当前页码，从1开始计数。
//   - pageSize: 每页显示的博客数量。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客信息的DTO（数据传输对象）列表。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回nil。
func FindBlogsInPage(ctx context.Context, page, pageSize int) ([]*dto.BlogDto, error) {
	blogs := make([]*po.Blog, 0)

	// 查询博客信息数据，按置顶优先和创建时间倒序排序，并进行分页处理。
	result := storage.Storage.Db.Model(&po.Blog{}).
		WithContext(ctx).
		Order("is_top DESC").
		Order("create_time DESC").
		Find(&blogs).
		Offset((page - 1) * pageSize).
		Limit(pageSize)

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
			BId:        blog.BId,
			Brief:      blog.Brief,
			Title:      blog.Title,
			IsTop:      blog.IsTop,
			State:      blog.State,
			WordsNum:   blog.WordsNum,
			CategoryId: blog.CategoryId,
			CreateTime: blog.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateTime: blog.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		blogDtos = append(blogDtos, blogDto)
	}

	// 返回转换后的博客DTO列表和nil错误。
	return blogDtos, nil
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
	if blog.State {
		logger.Info("设置 ID 为 %v 的博客状态为隐藏", id)
		blog.State = false
	} else {
		logger.Info("设置 ID 为 %v 的博客状态为显示", id)
		blog.State = true
	}

	// 更新博客状态到数据库。
	if err := tx.Model(blog).Update("b_state", blog.State).Where("b_id = ?", id).Error; err != nil {
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
	if blog.IsTop {
		logger.Info("取消 ID 为 %v 的播客置顶", id)
		blog.IsTop = false
	} else {
		logger.Info("设置 ID 为 %v 的播客置顶", id)
		blog.IsTop = true
	}

	// 更新数据库中的博客置顶状态
	if err := tx.Model(blog).Update("is_top", blog.IsTop).Where("b_id = ?", id).Error; err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("设置博客置顶失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	// 提交事务
	tx.Commit()

	return nil
}
