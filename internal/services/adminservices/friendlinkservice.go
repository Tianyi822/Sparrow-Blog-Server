package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/repositories/friendlinkrepo"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
)

// GetAllFriendLinks 获取所有友链信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.FriendLinkDto: 包含友链信息的 DTO 列表。
//   - error: 如果在查询友链时发生错误，则返回该错误。
func GetAllFriendLinks(ctx context.Context) ([]*dto.FriendLinkDto, error) {
	return friendlinkrepo.FindAllFriendLinks(ctx)
}

// CreateFriendLink 创建新的友链
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - friendLinkDto: 包含友链信息的数据传输对象。
//
// 返回值:
//   - error: 如果创建过程中发生错误，则返回错误信息；否则返回 nil。
func CreateFriendLink(ctx context.Context, friendLinkDto *dto.FriendLinkDto) error {
	// 开启创建友链事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("创建友链失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法创建友链
	err := friendlinkrepo.CreateFriendLink(tx, friendLinkDto)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("创建友链成功")

	return nil
}

// UpdateFriendLink 更新友链信息
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - friendLinkDto: 包含友链更新信息的数据传输对象。
//
// 返回值:
//   - error: 如果更新过程中发生错误，则返回错误信息；否则返回 nil。
func UpdateFriendLink(ctx context.Context, friendLinkDto *dto.FriendLinkDto) error {
	// 开启更新友链事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新友链失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法更新友链
	err := friendlinkrepo.UpdateFriendLinkByID(tx, friendLinkDto)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("更新友链成功")

	return nil
}

// DeleteFriendLinkById 删除指定ID的友链
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要删除的友链的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteFriendLinkById(ctx context.Context, id string) error {
	// 开启删除友链事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除友链失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法根据ID删除友链
	err := friendlinkrepo.DeleteFriendLinkById(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除友链成功")

	return nil
}

// UpdateFriendLinkDisplay 切换友链的显示状态
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - friendLinkId: 要更新的友链ID。
//
// 返回值:
//   - bool: 切换后的显示状态。
//   - error: 如果更新过程中发生错误，则返回错误信息；否则返回 nil。
func UpdateFriendLinkDisplay(ctx context.Context, friendLinkId string) (bool, error) {
	// 首先查询当前友链的显示状态
	friendLinkDto, err := friendlinkrepo.FindFriendLinkById(ctx, friendLinkId)
	if err != nil {
		return false, err
	}

	// 切换显示状态
	newDisplay := !friendLinkDto.Display

	// 开启更新友链显示状态事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("切换友链显示状态失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法更新友链显示状态
	err = friendlinkrepo.UpdateFriendLinkDisplayById(tx, friendLinkId, newDisplay)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 提交事务
	tx.Commit()
	logger.Info("切换友链显示状态成功")

	return newDisplay, nil
}
