package readcountrepo

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/po"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"time"
)

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
