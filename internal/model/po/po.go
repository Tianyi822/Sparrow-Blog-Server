package po

import "time"

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

type ImgInfo struct {
	ImgId      string    `gorm:"column:img_id;primaryKey"`                                    // 图片ID
	ImgName    string    `gorm:"column:img_name;unique"`                                      // 图片名称
	ImgType    string    `gorm:"column:img_type"`                                             // 图片格式
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (ii *ImgInfo) TableName() string {
	return "H2_IMG_INFO"
}

type Comment struct {
	CommentId    string    `gorm:"column:comment_id;primaryKey"`                                // 评论ID
	UserName     string    `gorm:"column:user_name"`                                            // 用户名
	UserEmail    string    `gorm:"column:user_email"`                                           // 用户邮箱
	UserUrl      string    `gorm:"column:user_url"`                                             // 用户网址
	BlogId       string    `gorm:"column:blog_id"`                                              // 博客ID
	OriginPostId string    `gorm:"column:original_poster_id"`                                   // 楼主评论ID
	Content      string    `gorm:"column:content"`                                              // 评论内容
	CreateTime   time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime   time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (c *Comment) TableName() string {
	return "H2_COMMENT"
}
