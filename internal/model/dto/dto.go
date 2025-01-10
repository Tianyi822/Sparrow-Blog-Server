package dto

// Dto 是一个接口，定义了所有数据传输对象（DTO）必须实现的方法
type Dto interface {
	// DtoFlag 方法返回一个字符串，用于标识 Dto 的某种特性或状态
	DtoFlag() string

	// Name 方法返回一个字符串，用于获取 Dto 的名称
	Name() string
}

// BlogInfoDto 用于表示博客信息的传输对象
type BlogInfoDto struct {
	BlogId string `json:"blog_id,omitempty"`
	Title  string `json:"title,omitempty"`
	Brief  string `json:"brief,omitempty"`
}

// DtoFlag 方法返回 BlogInfoDto 的标识字符串
func (bid *BlogInfoDto) DtoFlag() string {
	return "BlogInfoDto"
}

// Name 方法返回 BlogInfoDto 的名称
func (bid *BlogInfoDto) Name() string {
	return bid.Title
}

// ImgDto 图片数据
type ImgDto struct {
	ImgId   string `json:"img_id,omitempty"`
	ImgName string `json:"img_name,omitempty"`
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
