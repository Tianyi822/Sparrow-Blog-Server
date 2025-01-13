package po

import (
	"time"
)

type BlogInfo struct {
	BlogId     string    `gorm:"column:blog_id;primaryKey;"`                                  // 博客ID
	Title      string    `gorm:"column:title;unique"`                                         // 博客标题
	Brief      string    `gorm:"column:brief"`                                                // 博客简介
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (hbi *BlogInfo) TableName() string {
	return "H2_BLOG_INFO"
}
