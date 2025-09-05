package blogreadrepo

import (
	"errors"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"strings"

	"gorm.io/gorm"
)

// UpInsertBlogReadCount 添加或更新博客阅读数
func UpInsertBlogReadCount(tx *gorm.DB, brcdto *dto.BlogReadCountDto) error {
	// 根据博客ID和日期生成唯一的阅读记录ID
	readId, err := utils.GenId(brcdto.BlogId + brcdto.ReadDate)
	if err != nil {
		msg := fmt.Sprintf("生成博客阅读数 ID 失败: %v", err)
		logger.Warn(msg)
		return errors.New(msg)
	}
	// 将生成的readId赋值给传入的dto对象
	brcdto.ReadId = readId

	// 先查询是否存在数据
	brc := &po.BlogReadCount{}
	err = tx.Model(&po.BlogReadCount{}).
		Where("read_id = ?", readId).
		Find(&brc).
		Error
	if err != nil {
		msg := fmt.Sprintf("数据库查询出错: %v", err.Error())
		logger.Warn(msg)
		return err
	}

	if len(strings.Trim(brc.BlogId, " ")) == 0 {
		logger.Info("开始添加阅读记录数")
		if err := tx.Create(&po.BlogReadCount{
			ReadId:    readId,
			BlogId:    brcdto.BlogId,
			ReadCount: brcdto.ReadCount,
			ReadDate:  brcdto.ReadDate,
		}).Error; err != nil {
			msg := fmt.Sprintf("添加博客阅读数失败: %v", err)
			logger.Warn(msg)
			return errors.New(msg)
		}
		logger.Info("添加阅读记录数完成")
		return nil
	} else {
		// 数据已存在，则更新数据
		logger.Info("开始更新阅读记录数")
		if err := tx.Model(&po.BlogReadCount{}).
			Where("read_id = ?", readId).
			Update("read_count", brcdto.ReadCount+brc.ReadCount).
			Error; err != nil {
			msg := fmt.Sprintf("更新博客阅读数失败: %v", err)
			logger.Warn(msg)
			return errors.New(msg)
		}
		logger.Info("更新阅读记录数据完成")
		return nil
	}
}
