package dto

type Dto interface {
	// DtoFlag 用例标志
	DtoFlag() string

	// Name 用例名称
	Name() string
}

type BlogInfoDto struct {
	BlogId string `json:"blog_id,omitempty"`
	Title  string `json:"title,omitempty"`
	Brief  string `json:"brief,omitempty"`
}

func (bid *BlogInfoDto) DtoFlag() string {
	return "BlogInfoDto"
}

func (bid *BlogInfoDto) Name() string {
	return bid.Title
}
