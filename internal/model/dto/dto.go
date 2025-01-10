package dto

// Dto is an interface that defines a contract for data transfer objects (DTOs).
// It includes two methods that any type implementing this interface must provide.
type Dto interface {
	// DtoFlag returns a string that represents a flag or identifier for the DTO.
	// This method is used to distinguish different types of DTOs or to provide
	// additional metadata about the DTO.
	DtoFlag() string

	// Name returns a string that represents the name of the DTO.
	// This method is typically used to get a human-readable name for the DTO,
	// which can be useful for logging, debugging, or display purposes.
	Name() string
}

// BlogInfoDto is a struct that holds information about a blog post.
// It includes fields for the blog ID, title, and a brief summary.
type BlogInfoDto struct {
	BlogId string `json:"blog_id,omitempty"`
	Title  string `json:"title,omitempty"`
	Brief  string `json:"brief,omitempty"`
}

// DtoFlag returns a string identifier for the type of DTO.
// This method is typically used to identify the type of data transfer object.
func (bid *BlogInfoDto) DtoFlag() string {
	return "BlogInfoDto"
}

// Name returns the title of the blog post.
// This method provides a way to access the title field of the BlogInfoDto struct.
func (bid *BlogInfoDto) Name() string {
	return bid.Title
}

// ImgDto is a struct that represents an image data transfer object.
// This struct is used to encapsulate image-related data for transfer between different layers
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
