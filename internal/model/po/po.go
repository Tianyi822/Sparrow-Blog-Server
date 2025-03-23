package po

import "time"

type HBlog struct {
	BId        string    `gorm:"column:b_id;primaryKey"`                                      // 博客 ID
	Title      string    `gorm:"column:b_title;unique"`                                       // 博客标题
	Brief      string    `gorm:"column:b_brief"`                                              // 博客简介
	CategoryId string    `gorm:"column:category_id"`                                          // 逻辑外键字段（无约束）
	State      bool      `gorm:"column:b_state"`                                              // 博客状态
	WordsNum   uint16    `gorm:"column:b_words_num"`                                          // 博客字数
	IsTop      bool      `gorm:"column:is_top"`                                               // 是否置顶
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (hb *HBlog) TableName() string {
	return "H2_BLOG"
}

type Category struct {
	CId        string    `gorm:"column:c_id;primaryKey"`                                      // 分类 ID
	CName      string    `gorm:"column:c_name;unique"`                                        // 分类名称
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (c *Category) TableName() string {
	return "H2_CATEGORY"
}

type Tag struct {
	TId        string    `gorm:"column:t_id;primaryKey"`                                      // 标签 ID
	TName      string    `gorm:"column:t_name;unique"`                                        // 标签名称
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (t *Tag) TableName() string {
	return "H2_TAG"
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
	CommentId      string    `gorm:"column:comment_id;primaryKey"`                                // 评论ID
	CommenterName  string    `gorm:"column:commenter_name"`                                       // 评论者名
	CommenterEmail string    `gorm:"column:commenter_email"`                                      // 评论者邮箱
	CommenterUrl   string    `gorm:"column:commenter_url"`                                        // 评论者网址
	BlogId         string    `gorm:"column:blog_id"`                                              // 博客ID
	OriginPostId   string    `gorm:"column:original_poster_id"`                                   // 楼主评论ID
	Content        string    `gorm:"column:content"`                                              // 评论内容
	CreateTime     time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime     time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (c *Comment) TableName() string {
	return "H2_COMMENT"
}

type FriendLink struct {
	FriendLinkId   string    `gorm:"column:friend_link_id;primaryKey"`                            // 友链ID
	FriendLinkName string    `gorm:"column:friend_link_name"`                                     // 友链名称
	FriendLinkUrl  string    `gorm:"column:friend_link_url"`                                      // 友链地址
	CreateTime     time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime     time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (fl *FriendLink) TableName() string {
	return "H2_FRIEND_LINK"
}
