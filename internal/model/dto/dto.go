package dto

// Dto 是一个接口，定义了所有数据传输对象（DTO）必须实现的方法
type Dto interface {
	// DtoFlag 方法返回一个字符串，用于标识 Dto 的某种特性或状态
	DtoFlag() string

	// Name 方法返回一个字符串，用于获取 Dto 的名称
	Name() string
}

type BlogDto struct {
	BId        string      `json:"b_id,omitempty"`
	Title      string      `json:"title,omitempty"`
	Brief      string      `json:"brief,omitempty"`
	CategoryId string      `json:"category_id,omitempty"`
	Category   CategoryDto `json:"category,omitempty"`
	Tags       []TagDto    `json:"tags,omitempty"`
	State      bool        `json:"state,omitempty"`
	WordsNum   uint16      `json:"words_num,omitempty"`
	IsTop      bool        `json:"is_top,omitempty"`
	CreateTime string      `json:"create_time,omitempty"`
	UpdateTime string      `json:"update_time,omitempty"`
}

type TagDto struct {
	TId   string
	TName string
}

func (ht *TagDto) DtoFlag() string {
	return "TagDto"
}

func (ht *TagDto) Name() string {
	return ht.TName
}

type CategoryDto struct {
	CId   string `json:"c_id,omitempty"`
	CName string `json:"c_name,omitempty"`
}

func (hc *CategoryDto) DtoFlag() string {
	return "CategoryDto"
}

func (hc *CategoryDto) Name() string {
	return hc.CName
}

// ImgDto 图片数据
type ImgDto struct {
	ImgId   string `json:"img_id,omitempty"`
	ImgName string `json:"img_name,omitempty"`
	ImgType string `json:"img_type,omitempty"`
}

func (i *ImgDto) DtoFlag() string {
	return "ImgDto"
}

func (i *ImgDto) Name() string {
	return i.ImgName
}

// ImgsDto 图片数据
type ImgsDto struct {
	Imgs []ImgDto `json:"imgs,omitempty"`
}

func (is *ImgsDto) DtoFlag() string {
	return "ImgsDto"
}

func (is *ImgsDto) Name() string {
	return "ImgsDto"
}

type FriendLinkDto struct {
	FriendLinkId   string `json:"friend_link_id,omitempty"`
	FriendLinkName string `json:"friend_link_name,omitempty"`
	FriendLinkUrl  string `json:"friend_link_url,omitempty"`
}

func (fl *FriendLinkDto) DtoFlag() string {
	return "FriendLinkDto"
}

func (fl *FriendLinkDto) Name() string {
	return fl.FriendLinkName
}
