package vo

import "time"

type Vo interface {
	// VoFlag 用例标志
	VoFlag() string
}

type BlogVo struct {
	BlogId       string      `json:"blog_id,omitempty"`
	BlogTitle    string      `json:"blog_title,omitempty"`
	BlogImageId  string      `json:"blog_image_id,omitempty"`
	BlogBrief    string      `json:"blog_brief,omitempty"`
	Category     *CategoryVo `json:"category,omitempty"`
	Tags         []TagVo     `json:"tags,omitempty"`
	BlogState    bool        `json:"blog_state"`
	BlogWordsNum uint16      `json:"blog_words_num,omitempty"`
	BlogIsTop    bool        `json:"blog_is_top"`
	CreateTime   time.Time   `json:"create_time,omitempty"`
	UpdateTime   time.Time   `json:"update_time,omitempty"`
}

func (bv *BlogVo) VoFlag() string {
	return "BlogVo"
}

type TagVo struct {
	TagId   string `json:"tag_id,omitempty"`
	TagName string `json:"tag_name,omitempty"`
}

func (tv *TagVo) VoFlag() string {
	return "TagVo"
}

type CategoryVo struct {
	CategoryId   string `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

func (cv *CategoryVo) VoFlag() string {
	return "CategoryVo"
}

type ImgVo struct {
	ImgId      string    `json:"img_id,omitempty"`
	ImgName    string    `json:"img_name,omitempty"`
	ImgType    string    `json:"img_type,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

func (iv *ImgVo) VoFlag() string {
	return "ImgVo"
}

type CommentVo struct {
	CommentId        string      `json:"comment_id,omitempty"`
	CommenterEmail   string      `json:"commenter_email,omitempty"`
	BlogId           string      `json:"blog_id,omitempty"`
	BlogTitle        string      `json:"blog_title,omitempty"`
	OriginPostId     string      `json:"origin_post_id,omitempty"`
	ReplyToCommenter string      `json:"reply_to_commenter,omitempty"`
	Content          string      `json:"content,omitempty"`
	CreateTime       time.Time   `json:"create_time,omitempty"`
	SubComments      []CommentVo `json:"sub_comments,omitempty"`
}

func (cv *CommentVo) VoFlag() string {
	return "CommentVo"
}

type FriendLinkVo struct {
	FriendLinkId    string `json:"friend_link_id,omitempty"`
	FriendLinkName  string `json:"friend_link_name,omitempty"`
	FriendLinkUrl   string `json:"friend_link_url,omitempty"`
	FriendAvatarUrl string `json:"friend_avatar_url,omitempty"`
	FriendDescribe  string `json:"friend_describe,omitempty"`
	Display         bool   `json:"display"`
}

func (flv *FriendLinkVo) VoFlag() string {
	return "FriendLinkVo"
}
