package entity

import (
	"context"
	"errors"
	"fmt"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"time"
)

type ImgInfo struct {
	ImgId      string    `gorm:"column:img_id;primaryKey"`                                    // 图片ID
	ImgName    string    `gorm:"column:img_name;unique"`                                      // 图片名称
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (ii *ImgInfo) TableName() string {
	return "H2_IMG_INFO"
}

// FindOneById 根据图片ID查找图片信息
// 参数:
//   - ctx: 上下文对象，用于控制请求的上下文，如超时、取消等
//
// 返回值:
//   - error: 如果查找过程中发生错误，返回错误信息；否则返回nil
func (ii *ImgInfo) FindOneById(ctx context.Context) (int64, error) {
	// 查询图片信息数据
	result := storage.Storage.Db.WithContext(ctx).Model(ii).Where("H2_IMG_INFO.img_id = ?", ii.ImgId).First(ii)

	// 检查查询结果是否有错误
	if result.Error != nil {
		// 如果查询过程中发生其他错误，则记录错误日志并返回错误
		msg := fmt.Sprintf("查询图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return result.RowsAffected, errors.New(msg)
	}
	// 如果查询成功，返回nil表示没有错误
	return result.RowsAffected, nil
}

// FindByNameLike 根据图片名称模糊查询
// 参数:
//   - ctx: 上下文对象
//
// 返回值:
//   - []ImgInfo: 匹配的图片列表
//   - error: 错误信息
func (ii *ImgInfo) FindByNameLike(ctx context.Context) ([]ImgInfo, error) {
	var images []ImgInfo

	// 查询图片信息数据
	result := storage.Storage.Db.WithContext(ctx).Model(ii).Where("H2_IMG_INFO.img_name LIKE ?", "%"+ii.ImgName+"%").Find(&images)

	// 检查查询结果是否有错误
	if result.Error != nil {
		msg := fmt.Sprintf("模糊查询图片信息失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 如果查询成功，返回nil表示没有错误
	return images, nil
}

// AddOne 添加一条记录
func (ii *ImgInfo) AddOne(ctx context.Context) (affectedNum int64, err error) {
	//开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		// 恢复机制，用于捕获和处理 panic
		if r := recover(); r != nil {
			// 如果发生 panic，记录错误日志并回滚事务
			logger.Error("创建图片信息数据失败: %v", r)
			tx.Rollback()
		} else if err != nil {
			// 如果有其他错误，回滚事务
			tx.Rollback()
		} else {
			// 如果没有错误，提交事务
			tx.Commit()
		}
	}()

	// 如果查询失败，则添加新数据
	logger.Info("添加图片信息数据")

	// 使用事务创建图片信息数据
	result := tx.Create(ii)
	if result.Error != nil {
		// 如果创建数据失败，记录错误日志并返回错误
		msg := fmt.Sprintf("创建图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 记录创建成功的日志并返回受影响的行数
	logger.Info("创建图片信息数据成功: %v", result.RowsAffected)
	return result.RowsAffected, nil
}
