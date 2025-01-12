package po

import (
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
