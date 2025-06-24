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

// Comment 评论实体对象
//
// 回复场景说明：
// 1. 楼主评论（顶级评论）:
//   - OriginPostId = ""（空字符串或NULL）
//   - ReplyToCommentId = ""（空字符串或NULL）
//
// 2. 回复楼主评论（一级回复）:
//   - OriginPostId = 楼主评论的CommentId
//   - ReplyToCommentId = 楼主评论的CommentId
//
// 3. 回复子评论（多级回复）:
//   - OriginPostId = 楼主评论的CommentId（保持不变）
//   - ReplyToCommentId = 被回复评论的CommentId
//
// 示例：
// 楼主评论A: OriginPostId="", ReplyToCommentId=""
// ├─ 回复A的评论B: OriginPostId="A", ReplyToCommentId="A"
// ├─ 回复A的评论C: OriginPostId="A", ReplyToCommentId="A"
// │  └─ 回复C的评论D: OriginPostId="A", ReplyToCommentId="C"
// └─ 回复B的评论E: OriginPostId="A", ReplyToCommentId="B"
type Comment struct {
	CommentId        string    `gorm:"column:comment_id;primaryKey"`                                // 评论 ID
	CommenterEmail   string    `gorm:"column:commenter_email"`                                      // 评论者邮箱
	BlogId           string    `gorm:"column:blog_id"`                                              // 博客 ID
	OriginPostId     string    `gorm:"column:original_poster_id"`                                   // 楼主评论 ID（用于分组，空值表示自己就是楼主评论）
	ReplyToCommentId string    `gorm:"column:reply_to_comment_id"`                                  // 回复的评论 ID（具体回复哪条评论，空值表示不回复任何评论）
	Content          string    `gorm:"column:comment_content"`                                      // 评论内容
	CreateTime       time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime       time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (c *Comment) TableName() string {
	return "COMMENT"
}

type FriendLink struct {
	FriendLinkId        string    `gorm:"column:friend_link_id;primaryKey"`                            // 友链 ID
	FriendLinkName      string    `gorm:"column:friend_link_name"`                                     // 友链名称
	FriendLinkUrl       string    `gorm:"column:friend_link_url"`                                      // 友链地址
	FriendLinkAvatarUrl string    `gorm:"column:friend_link_avatar_url"`                               // 友链头像
	FriendDescribe      string    `gorm:"column:friend_describe"`                                      // 友链描述
	Display             bool      `gorm:"column:display"`                                              // 是否展示
	CreateTime          time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`                // 创建时间
	UpdateTime          time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP;autoUpdateTime"` // 更新时间
}

func (fl *FriendLink) TableName() string {
	return "FRIEND_LINK"
}
