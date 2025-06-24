package dto

import "time"

// Dto 是一个接口，定义了所有数据传输对象（DTO）必须实现的方法
type Dto interface {
	// DtoFlag 方法返回一个字符串，用于标识 Dto 的某种特性或状态
	DtoFlag() string

	// Name 方法返回一个字符串，用于获取 Dto 的名称
	Name() string
}

type BlogDto struct {
	BlogId       string       `json:"blog_id,omitempty"`
	BlogTitle    string       `json:"blog_title,omitempty"`
	BlogImageId  string       `json:"blog_image_id,omitempty"`
	BlogBrief    string       `json:"blog_brief,omitempty"`
	CategoryId   string       `json:"category_id,omitempty"`
	Category     *CategoryDto `json:"category,omitempty"`
	Tags         []TagDto     `json:"tags,omitempty"`
	BlogState    bool         `json:"blog_state"`
	BlogWordsNum uint16       `json:"blog_words_num,omitempty"`
	BlogIsTop    bool         `json:"blog_is_top"`
	CreateTime   time.Time    `json:"create_time,omitempty"`
	UpdateTime   time.Time    `json:"update_time,omitempty"`
}

func (hb *BlogDto) DtoFlag() string {
	return "BlogDto"
}

func (hb *BlogDto) Name() string {
	return hb.BlogTitle
}

type TagDto struct {
	TagId   string `json:"tag_id,omitempty"`
	TagName string `json:"tag_name,omitempty"`
}

func (ht *TagDto) DtoFlag() string {
	return "TagDto"
}

func (ht *TagDto) Name() string {
	return ht.TagName
}

type CategoryDto struct {
	CategoryId   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

func (hc *CategoryDto) DtoFlag() string {
	return "CategoryDto"
}

func (hc *CategoryDto) Name() string {
	return hc.CategoryName
}

// ImgDto 图片数据
type ImgDto struct {
	ImgId      string    `json:"img_id,omitempty"`
	ImgName    string    `json:"img_name,omitempty"`
	ImgType    string    `json:"img_type,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

func (i *ImgDto) DtoFlag() string {
	return "ImgDto"
}

func (i *ImgDto) Name() string {
	return i.ImgName
}

type ImgsDto struct {
	Imgs []ImgDto `json:"imgs,omitempty"`
}

func (i *ImgsDto) DtoFlag() string {
	return "ImgsDto"
}

func (i *ImgsDto) Name() string {
	return "ImgsDto"
}

type FriendLinkDto struct {
	FriendLinkId    string `json:"friend_link_id,omitempty"`
	FriendLinkName  string `json:"friend_link_name,omitempty"`
	FriendLinkUrl   string `json:"friend_link_url,omitempty"`
	FriendAvatarUrl string `json:"friend_avatar_url,omitempty"`
	FriendDescribe  string `json:"friend_describe,omitempty"`
	Display         bool   `json:"display"`
}

func (fl *FriendLinkDto) DtoFlag() string {
	return "FriendLinkDto"
}

func (fl *FriendLinkDto) Name() string {
	return fl.FriendLinkName
}

type CommentDto struct {
	CommentId        string    `json:"comment_id,omitempty"`
	CommenterEmail   string    `json:"commenter_email,omitempty"`
	BlogId           string    `json:"blog_id,omitempty"`
	OriginPostId     string    `json:"origin_post_id,omitempty"`
	ReplyToCommentId string    `json:"reply_to_comment_id,omitempty"`
	ReplyToCommenter string    `json:"reply_to_commenter,omitempty"`
	Content          string    `json:"content,omitempty"`
	CreateTime       time.Time `json:"create_time,omitempty"`
}

func (c *CommentDto) DtoFlag() string {
	return "CommentDto"
}

func (c *CommentDto) Name() string {
	return c.CommentId
}
