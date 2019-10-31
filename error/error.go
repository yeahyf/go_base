package error

//自定义错误类型
type Error struct {
	ErrCode int
	ErrMsg  string
}

func New(code int, msg string) *Error {
	return &Error{ErrCode: code, ErrMsg: msg}
}

//实现错误接口
func (err *Error) Error() string {
	return err.ErrMsg
}
