package vo

type Vo interface {
	// VoFlag 用例标志
	VoFlag() string
}

type BlogInfoVo struct {
	BlogId string `json:"blog_id,omitempty"`
	Title  string `json:"title,omitempty"`
	Brief  string `json:"brief,omitempty"`
}

func (bid *BlogInfoVo) VoFlag() string {
	return "BlogInfoVo"
}
