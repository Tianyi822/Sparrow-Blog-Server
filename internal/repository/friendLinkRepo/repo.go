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

// GetFriendLinkByName 根据友链名称查询友链信息
//
// 参数:
//
//	ctx      context.Context - 上下文对象，用于控制查询请求的生命周期和截止时间
//	name     string          - 需要查询的友链名称
//
// 返回值:
//
//	*dto.FriendLinkDto - 包含友链ID、名称、URL的DTO对象，若未找到返回nil
//	error              - 查询过程中产生的错误，成功时返回nil
func GetFriendLinkByName(ctx context.Context, name string) (*dto.FriendLinkDto, error) {
	friendLink := &po.FriendLink{}

	// 执行数据库查询以获取指定名称的友链记录
	result := storage.Storage.Db.Model(friendLink).
		WithContext(ctx).
		Where("H2_FRIEND_LINK.friend_link_name = ?", name).
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
