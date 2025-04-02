package imgInfoRepo

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

// FindImgById 根据图片ID查询单条图片信息
// 参数：
//   - ctx: 上下文对象
//   - imgId: 图片ID
//
// 返回值: 图片信息实体指针和错误信息
func FindImgById(ctx context.Context, imgId string) (*dto.ImgDto, error) {
	var img po.H2Img
	// 使用GORM查询数据库，根据img_id查找单条记录
	result := storage.Storage.Db.Model(img).WithContext(ctx).Where("img_id = ?", imgId).First(&img)
	if result.Error != nil {
		msg := fmt.Sprintf("查询图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	return &dto.ImgDto{
		ImgId:   img.ImgId,
		ImgName: img.ImgName,
		ImgType: img.ImgType,
	}, nil
}

// FindAllImgs 查询数据库中所有的图片信息，并将其转换为 DTO 对象列表返回。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//
// 返回值:
//   - []dto.ImgDto: 包含所有图片信息的 DTO 列表，每个 DTO 包含图片的 ID、名称和类型。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func FindAllImgs(ctx context.Context) ([]dto.ImgDto, error) {
	var imgs []po.H2Img

	logger.Info("查询所有图片信息数据")
	// 使用 GORM 查询数据库中的所有图片信息。
	result := storage.Storage.Db.Model(&po.H2Img{}).WithContext(ctx).Find(&imgs)
	if result.Error != nil {
		// 如果查询失败，记录错误日志并返回错误信息。
		msg := fmt.Sprintf("查询所有图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	logger.Info("查询所有图片信息数据成功")

	// 将查询到的图片信息转换为 DTO 列表。
	imgDtos := make([]dto.ImgDto, 0)
	for _, img := range imgs {
		imgDtos = append(imgDtos, dto.ImgDto{
			ImgId:   img.ImgId,
			ImgName: img.ImgName,
			ImgType: img.ImgType,
		})
	}

	// 返回转换后的 DTO 列表和 nil 错误。
	return imgDtos, nil
}

// AddImgInfoBatch 批量添加图片信息记录
// - ctx: 上下文对象
// - imgs: 图片信息实体切片
//
// 返回值: 受影响的行数和错误信息
// - int64: 受影响的行数
// - error: 错误信息
func AddImgInfoBatch(ctx context.Context, imgs []po.H2Img) (int64, error) {
	if len(imgs) == 0 {
		return 0, nil
	}

	tx := storage.Storage.Db.Model(&po.H2Img{}).WithContext(ctx).Begin()
	// 使用defer确保在panic时回滚事务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("批量添加图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("批量添加图片信息数据")
	// 执行创建操作
	result := tx.Create(imgs)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("批量添加图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}

	// 提交事务
	tx.Commit()
	logger.Info("批量添加图片信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}

// DeleteImgById 根据图片 ID 删除对应的图片数据。
// 参数:
//   - tx: *gorm.DB，数据库事务对象，用于执行删除操作。
//   - id: string，图片的唯一标识符，用于定位需要删除的图片数据。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回包含错误信息的 error 对象；否则返回 nil。
func DeleteImgById(tx *gorm.DB, id string) error {
	logger.Info("删除 ID 为 %v 的图片数据", id)

	if err := tx.Where("img_id = ?", id).Delete(&po.H2Img{}).Error; err != nil {
		msg := fmt.Sprintf("删除图片数据失败: %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}

	logger.Info("删除 ID 为 %v 的图片数据成功", id)

	return nil
}

// UpdateImgNameById 更新图片信息记录的名称
// - ctx: 上下文对象
// - id: 图片信息ID
// - name: 新的名称
//
// 返回值:
// - int64: 受影响的行数
// - error: 错误信息
func UpdateImgNameById(ctx context.Context, id string, name string) (int64, error) {
	tx := storage.Storage.Db.Model(&po.H2Img{}).WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新图片信息数据失败: %v", r)
			tx.Rollback()
		}
	}()

	logger.Info("更新图片信息数据")
	result := tx.Model(&po.H2Img{}).Where("img_id = ?", id).Update("img_name", name)
	if result.Error != nil {
		tx.Rollback()
		msg := fmt.Sprintf("更新图片信息数据失败: %v", result.Error)
		logger.Error(msg)
		return 0, errors.New(msg)
	}
	// 提交事务
	tx.Commit()
	logger.Info("更新图片信息数据成功: %v", result.RowsAffected)

	return result.RowsAffected, nil
}
