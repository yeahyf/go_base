package error

//自定义错误类型
type MyError struct {
	ErrCode int
	ErrMsg  string
}

func New(code int, msg string) *MyError {
	return &MyError{ErrCode: code, ErrMsg: msg}
}

func NewWrapper(code int, err error) *MyError {
	return &MyError{ErrCode: code, ErrMsg: err.Error()}
}

//实现错误接口
func (err *MyError) Error() string {
	return err.ErrMsg
}
