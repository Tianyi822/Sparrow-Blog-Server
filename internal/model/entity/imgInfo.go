package entity

type ImgInfo struct {
	ImgId      int    `gorm:"column:img_id;primaryKey;autoIncrement"`                      // 图片ID
	ImgName    string `gorm:"column:img_name;unique"`                                      // 图片名称
	ImgOssPath string `gorm:"column:img_oss_path"`                                         // 图片的 OSS URL
	CreateTime string `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime string `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (ii *ImgInfo) TableName() string {
	return "H2_IMG_INFO"
}
