package webp

const (
	DownloadError = 0
	ConvertError  = 1
	UploadError   = 2
	DeleteError   = 3
)

type Err struct {
	Code int
	Msg  string
}

func NewWebErr(code int, msg string) error {
	return &Err{
		Code: code,
		Msg:  msg,
	}
}

func (e *Err) Error() string {
	return e.Msg
}
