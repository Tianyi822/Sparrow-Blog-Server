package friendlinkrepo

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"

	"gorm.io/gorm"
)

// UpdateFriendLinkByID 根据友链ID更新友链信息。
// 参数:
// - tx: 数据库事务对象，用于执行数据库操作。
// - friendLinkDto: 包含友链更新信息的数据传输对象，包含友链ID、友链名称、友链URL、友链头像和友链描述。
// 返回值:
// - error: 错误信息，如果更新过程中发生错误，则返回具体的错误信息；否则返回nil。
func UpdateFriendLinkByID(tx *gorm.DB, friendLinkDto *dto.FriendLinkDto) error {
	// 记录更新操作开始的日志，便于后续排查问题。
	logger.Info("更新友链数据")

	// 执行更新操作，根据友链ID更新所有友链字段。
	if err := tx.Model(&po.FriendLink{}).Where("friend_link_id = ?", friendLinkDto.FriendLinkId).Updates(po.FriendLink{
		FriendLinkName:      friendLinkDto.FriendLinkName,
		FriendLinkUrl:       friendLinkDto.FriendLinkUrl,
		FriendLinkAvatarUrl: friendLinkDto.FriendAvatarUrl,
		FriendDescribe:      friendLinkDto.FriendDescribe,
		Display:             friendLinkDto.Display,
	}).Error; err != nil {
		// 如果更新失败，记录错误日志。
		msg := fmt.Sprintf("更新友链数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 记录更新成功的日志。
	logger.Info("更新友链数据成功")
	return nil
}

// DeleteFriendLinkById 根据ID删除友链数据。
// 参数:
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - friendLinkId: 需要删除的友链ID。
//
// 返回值:
//   - error: 操作失败时返回的错误。
func DeleteFriendLinkById(tx *gorm.DB, friendLinkId string) error {
	// 记录删除操作开始日志
	logger.Info("删除友链数据")

	// 根据ID执行删除操作
	if err := tx.Where("friend_link_id = ?", friendLinkId).Delete(&po.FriendLink{}).Error; err != nil {
		msg := fmt.Sprintf("删除友链数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 记录删除成功日志并返回结果
	logger.Info("删除友链数据成功")
	return nil
}

// CreateFriendLink 创建友链记录
//
// 参数:
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - friendLinkDto: 包含友链信息的数据传输对象
//
// 返回值:
//   - error: 操作过程中产生的错误（成功时返回nil）
func CreateFriendLink(tx *gorm.DB, friendLinkDto *dto.FriendLinkDto) error {
	friendLinkId, err := utils.GenId(friendLinkDto.FriendLinkUrl)
	if err != nil {
		return err
	}

	// 将生成的ID保存到DTO中
	friendLinkDto.FriendLinkId = friendLinkId

	// 根据传入的FriendLinkDto对象初始化一个FriendLink对象
	friendLink := &po.FriendLink{
		FriendLinkId:        friendLinkId,
		FriendLinkName:      friendLinkDto.FriendLinkName,
		FriendLinkUrl:       friendLinkDto.FriendLinkUrl,
		FriendLinkAvatarUrl: friendLinkDto.FriendAvatarUrl,
		FriendDescribe:      friendLinkDto.FriendDescribe,
		Display:             friendLinkDto.Display,
	}

	logger.Info("添加友链数据")
	if err := tx.Create(friendLink).Error; err != nil {
		msg := fmt.Sprintf("添加友链失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	logger.Info("添加友链数据成功")
	return nil
}

// FindAllFriendLinks 查询所有友链信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//
// 返回值:
//   - []*dto.FriendLinkDto: 包含友链数据的DTO列表
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回nil。
func FindAllFriendLinks(ctx context.Context) ([]*dto.FriendLinkDto, error) {
	friendLinks := make([]*po.FriendLink, 0)

	// 使用GORM查询友链数据，按创建时间排序
	if err := storage.Storage.Db.Model(&po.FriendLink{}).
		WithContext(ctx).
		Order("create_time DESC").
		Find(&friendLinks).Error; err != nil {
		msg := fmt.Sprintf("查询友链信息数据失败: %v", err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 初始化友链DTO列表，用于存储转换后的友链数据
	friendLinkDtos := make([]*dto.FriendLinkDto, 0)

	// 遍历查询到的友链数据，将其转换为DTO格式并添加到结果列表中
	for _, friendLink := range friendLinks {
		friendLinkDto := &dto.FriendLinkDto{
			FriendLinkId:    friendLink.FriendLinkId,
			FriendLinkName:  friendLink.FriendLinkName,
			FriendLinkUrl:   friendLink.FriendLinkUrl,
			FriendAvatarUrl: friendLink.FriendLinkAvatarUrl,
			FriendDescribe:  friendLink.FriendDescribe,
			Display:         friendLink.Display,
		}
		friendLinkDtos = append(friendLinkDtos, friendLinkDto)
	}

	// 返回转换后的友链DTO列表和nil错误
	return friendLinkDtos, nil
}

// FindFriendLinkById 根据友链ID查询单个友链信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - friendLinkId: 需要查询的友链ID。
//
// 返回值:
//   - *dto.FriendLinkDto: 包含友链数据的DTO对象
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回nil。
func FindFriendLinkById(ctx context.Context, friendLinkId string) (*dto.FriendLinkDto, error) {
	friendLink := &po.FriendLink{}

	// 使用GORM查询指定ID的友链数据
	if err := storage.Storage.Db.Model(&po.FriendLink{}).
		WithContext(ctx).
		Where("friend_link_id = ?", friendLinkId).
		First(&friendLink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("友链不存在")
		}
		msg := fmt.Sprintf("查询友链信息失败: %v", err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 将查询到的友链数据转换为DTO格式
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkId:    friendLink.FriendLinkId,
		FriendLinkName:  friendLink.FriendLinkName,
		FriendLinkUrl:   friendLink.FriendLinkUrl,
		FriendAvatarUrl: friendLink.FriendLinkAvatarUrl,
		FriendDescribe:  friendLink.FriendDescribe,
		Display:         friendLink.Display,
	}

	return friendLinkDto, nil
}

// UpdateFriendLinkDisplayById 根据友链ID更新友链的显示状态
// 参数:
//   - tx: 数据库事务对象，用于执行数据库操作。
//   - friendLinkId: 需要更新的友链ID。
//   - display: 新的显示状态。
//
// 返回值:
//   - error: 操作失败时返回的错误。
func UpdateFriendLinkDisplayById(tx *gorm.DB, friendLinkId string, display bool) error {
	// 记录更新操作开始日志
	logger.Info("更新友链显示状态")

	// 根据ID执行更新操作，只更新 display 字段
	if err := tx.Model(&po.FriendLink{}).Where("friend_link_id = ?", friendLinkId).Update("display", display).Error; err != nil {
		msg := fmt.Sprintf("更新友链显示状态失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	// 记录更新成功日志并返回结果
	logger.Info("更新友链显示状态成功")
	return nil
}
