package friendLinkRepo

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

// UpdateFriendLinkByID 根据友链ID更新友链信息。
// 参数:
// - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
// - friendLinkDto: 包含友链更新信息的数据传输对象，包含友链ID、友链名称和友链URL。
// 返回值:
// - int64: 受影响的行数，表示更新操作影响的数据库记录数量。
// - error: 错误信息，如果更新过程中发生错误，则返回具体的错误信息；否则返回nil。
func UpdateFriendLinkByID(ctx context.Context, friendLinkDto dto.FriendLinkDto) (int64, error) {
	// 开始数据库事务，确保操作的原子性。
	tx := storage.Storage.Db.Model(&po.FriendLink{}).WithContext(ctx).Begin()
	defer func() {
		// 捕获异常并回滚事务，防止因异常导致事务未正确关闭。
		if r := recover(); r != nil {
			logger.Error("更新友链数据失败: %v", r)
			tx.Rollback()
		}
	}()

	// 记录更新操作开始的日志，便于后续排查问题。
	logger.Info("更新友链数据")

	// 执行更新操作，根据友链ID更新友链名称和友链URL。
	result := tx.Model(&po.FriendLink{}).Where("friend_link_id = ?", friendLinkDto.FriendLinkId).
		Update("friend_link_name", friendLinkDto.FriendLinkName).
		Update("friend_link_url", friendLinkDto.FriendLinkUrl)
	if result.Error != nil {
		// 如果更新失败，回滚事务并记录错误日志。
		tx.Rollback()
		msg := fmt.Sprintf("更新友链数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 提交事务，确保更新操作生效。
	tx.Commit()

	// 记录更新成功的日志，包括受影响的行数。
	logger.Info("更新友链数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// DeleteFriendLinkById 根据ID删除友链数据，并使用数据库事务处理。
// 参数:
//
//	ctx: context.Context - 数据库操作的上下文。
//	friendLinkId: string - 需要删除的友链ID。
//
// 返回值:
//
//	int64: 删除操作影响的行数。
//	error: 操作失败时返回的错误。
func DeleteFriendLinkById(ctx context.Context, friendLinkId string) (int64, error) {
	// 开始数据库事务
	tx := storage.Storage.Db.Model(&po.FriendLink{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除友链数据失败: %v", r)
			tx.Rollback()
		}
	}()

	// 记录删除操作开始日志
	logger.Info("删除友链数据")

	// 根据ID执行删除操作
	result := tx.Delete(&po.FriendLink{FriendLinkId: friendLinkId})
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("删除友链数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()

	// 记录删除成功日志并返回结果
	logger.Info("删除友链数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}

// GetFriendLinkByNameLike 根据友链名称模糊查询友链信息
// 参数：
//
//	ctx context.Context: 请求上下文，用于控制超时和取消
//	name string: 友链名称的模糊查询条件（支持%通配符）
//
// 返回值：
//
//	*dto.FriendLinkDto: 查询到的友链数据
//	error: 查询失败时返回的错误信息
func GetFriendLinkByNameLike(ctx context.Context, name string) (*dto.FriendLinkDto, error) {
	friendLink := &po.FriendLink{}

	// 执行数据库查询以获取指定名称的友链记录
	result := storage.Storage.Db.Model(friendLink).
		WithContext(ctx).
		Where("friend_link_name LIKE ?", "%"+name+"%").
		First(friendLink)

	if result.Error != nil {
		// 处理查询错误并返回
		msg := fmt.Sprintf("查询友链失败: %v", result.Error)
		logger.Warn(msg)
		return nil, errors.New(msg)
	}

	// 将查询结果转换为DTO并返回
	return &dto.FriendLinkDto{
		FriendLinkId:   friendLink.FriendLinkId,
		FriendLinkName: friendLink.FriendLinkName,
		FriendLinkUrl:  friendLink.FriendLinkUrl,
	}, nil
}

// CreateFriendLink 创建友链记录
//
// 参数:
//
//	ctx: 上下文，用于控制请求范围和传递截止时间、取消信号等
//	friendLinkDto: 包含友链信息的数据传输对象
//
// 返回值:
//
//	int64: 影响的数据库行数（成功时返回1，失败时返回0）
//	error: 操作过程中产生的错误（成功时返回nil）
func CreateFriendLink(ctx context.Context, friendLinkDto dto.FriendLinkDto) (int64, error) {
	friendLinkId, err := utils.GenId(friendLinkDto.FriendLinkName)
	if err != nil {
		return 0, err
	}

	// 根据传入的FriendLinkDto对象初始化一个FriendLink对象
	friendLink := &po.FriendLink{
		FriendLinkId:   friendLinkId,
		FriendLinkName: friendLinkDto.FriendLinkName,
		FriendLinkUrl:  friendLinkDto.FriendLinkUrl,
	}

	// 开启事务
	tx := storage.Storage.Db.Model(friendLink).WithContext(ctx).Begin()

	// 使用defer语句确保在函数返回时事务被提交或回滚
	defer func() {
		if r := recover(); r != nil {
			logger.Error("添加友链失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("添加友链数据")
	result := tx.Create(friendLink)

	// 检查创建过程中是否有错误发生
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("添加友链失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()

	logger.Info("添加友链数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}
