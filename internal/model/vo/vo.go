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

func (bv *BlogInfoVo) VoFlag() string {
	return "BlogInfoVo"
}

type ImgInfoVo struct {
	ImgId   string `json:"img_id,omitempty"`
	ImgName string `json:"img_name,omitempty"`
	Err     string `json:"err,omitempty"`
}

func (iv *ImgInfoVo) VoFlag() string {
	return "ImgInfoVo"
}

type ImgInfosVo struct {
	Success []ImgInfoVo `json:"success,omitempty"`
	Failure []ImgInfoVo `json:"failure,omitempty"`
}

func (isv *ImgInfosVo) VoFlag() string {
	return "ImgInfosVo"
}
