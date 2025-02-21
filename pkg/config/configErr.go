package config

const (
	NoConfigFileErr = 0
)

type Err struct {
	Code int
	Msg  string
}

func NewNoConfigFileErr(msg string) *Err {
	return &Err{
		Code: NoConfigFileErr,
		Msg:  msg,
	}
}

func (e *Err) IsNoConfigFileErr() bool {
	return e.Code == NoConfigFileErr
}

func (e *Err) Error() string {
	return e.Msg
}
