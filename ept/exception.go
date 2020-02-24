///自定义error异常封装
package ept

import (
	"fmt"
	"runtime/debug"
)

//自定义错误类型
type Error struct {
	Code    uint32 //错误代码采用1000-9999整形
	Message string //消息的说明
}

func New(code uint32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NewWrapper(code uint32, err error) *Error {
	return &Error{
		Code:    code,
		Message: err.Error(),
	}
}

//实现错误接口
func (err *Error) Error() string {
	return err.Message
}

func PanicHandle() {
	if r := recover(); r != nil {
		//判断是否是某种类型的错误
		err, ok := r.(error)
		if !ok {
			debug.PrintStack()
		} else {
			fmt.Println(err)
		}
	}
}
