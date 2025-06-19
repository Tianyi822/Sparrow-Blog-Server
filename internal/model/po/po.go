package po

import "time"

type Blog struct {
	BlogId       string    `gorm:"column:blog_id;primaryKey"`                                   // 博客 ID
	BlogTitle    string    `gorm:"column:blog_title;unique"`                                    // 博客标题
	BlogImageId  string    `gorm:"column:blog_image_id"`                                        // 博客图片
	BlogBrief    string    `gorm:"column:blog_brief"`                                           // 博客简介
	CategoryId   string    `gorm:"column:category_id"`                                          // 逻辑外键字段（无约束）
	BlogState    bool      `gorm:"column:blog_state"`                                           // 博客状态
	BlogWordsNum uint16    `gorm:"column:blog_words_num"`                                       // 博客字数
	BlogIsTop    bool      `gorm:"column:blog_is_top"`                                          // 是否置顶
	CreateTime   time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime   time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (hb *Blog) TableName() string {
	return "BLOG"
}

type Category struct {
	CategoryId   string    `gorm:"column:category_id;primaryKey"`                               // 分类 ID
	CategoryName string    `gorm:"column:category_name;unique"`                                 // 分类名称
	CreateTime   time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime   time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (c *Category) TableName() string {
	return "CATEGORY"
}

type Tag struct {
	TagId      string    `gorm:"column:tag_id;primaryKey"`                                    // 标签 ID
	TagName    string    `gorm:"column:tag_name;unique"`                                      // 标签名称
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (t *Tag) TableName() string {
	return "TAG"
}

type BlogTag struct {
	BlogId string `gorm:"column:blog_id"`
	TagId  string `gorm:"column:tag_id"`
}

func (hb *BlogTag) TableName() string {
	return "BLOG_TAG"
}

type H2Img struct {
	ImgId      string    `gorm:"column:img_id;primaryKey"`                                    // 图片ID
	ImgName    string    `gorm:"column:img_name;unique"`                                      // 图片名称
	ImgType    string    `gorm:"column:img_type"`                                             // 图片格式
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (ii *H2Img) TableName() string {
	return "IMG"
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
	return "COMMENT"
}

type FriendLink struct {
	FriendLinkId    string    `gorm:"column:friend_link_id;primaryKey"`                            // 友链 ID
	FriendLinkName  string    `gorm:"column:friend_link_name"`                                     // 友链名称
	FriendLinkUrl   string    `gorm:"column:friend_link_url"`                                      // 友链地址
	FriendAvatarUrl string    `gorm:"column:friend_avatar_url"`                                    // 友链头像
	FriendDescribe  string    `gorm:"column:friend_describe"`                                      // 友链描述
	CreateTime      time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime      time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (fl *FriendLink) TableName() string {
	return "FRIEND_LINK"
}
