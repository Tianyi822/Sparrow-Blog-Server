package readcountrepo

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/utils"
	"sparrow_blog_server/storage"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UpsertBlogReadCount 添加或更新博客阅读数
func UpsertBlogReadCount(tx *gorm.DB, brcdto *dto.BlogReadCountDto) error {
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

// FindRecentSevenDaysReadCount 查询最近七天博客阅读数
func FindRecentSevenDaysReadCount(ctx context.Context) ([]*dto.BlogReadCountDto, error) {
	logger.Info("查询最近七天博客阅读数")
	var readCounts []*po.BlogReadCount
	err := storage.Storage.Db.WithContext(ctx).Model(&po.BlogReadCount{}).
		Select("sum(read_count) as read_count, read_date").
		Where("read_date >= ?", time.Now().AddDate(0, 0, -7).Format("20060102")).
		Group("read_date").
		Order("read_date").
		Scan(&readCounts).Error
	if err != nil {
		msg := fmt.Sprintf("数据库查询出错: %v", err.Error())
		logger.Warn(msg)
		return nil, errors.New(msg)
	}
	logger.Info("查询最近七天博客阅读数完成")

	logger.Info("开始转换阅读数据为 DTO")
	var readCountDtos []*dto.BlogReadCountDto
	for _, readCount := range readCounts {
		readCountDtos = append(readCountDtos, &dto.BlogReadCountDto{
			ReadCount: readCount.ReadCount,
			ReadDate:  readCount.ReadDate,
		})
	}
	logger.Info("转换完成")

	return readCountDtos, nil
}
