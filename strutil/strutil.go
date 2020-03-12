///字符串工具接口封装
package strutil

import (
	"encoding/base64"
	"sort"
	"unsafe"
)

///处理字符串与[]byte数组的转换

//byte数组转化为字符串，返回字符串
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//byte数组转化为字符串，返回字符串引用
func Bytes2Str(b []byte) *string {
	return (*string)(unsafe.Pointer(&b))
}

//字符串转化为byte数组
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}      // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

//字符串转化为byte数组
func String2bytes(s *string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}     // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

func SortString(list []string) {
	sort.Sort(sort.StringSlice(list))
}

//如果是gzip压缩，就转base64
func IsBase64String(str *string) bool {
	length := len(*str)
	if length == 0 || len(*str)%4 != 0 {
		return false
	}
	//只是判断前面20位
	if length > 20 {
		length = 20
	}

	b := String2bytes(str)
	for i := 0; i < length-2; i++ {
		v := b[i]
		if v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' ||
			v == '+' || v == '/' {
			continue
		} else {
			return false
		}
	}
	return true
}

func ConvertBytes(src *string) []byte {
	if !IsBase64String(src) {
		return String2bytes(src)
	}
	r, err := base64.StdEncoding.DecodeString(*src)
	//判断是否是gzip压缩之后的数据
	if err == nil && len(r) >= 2 && r[0] == 0x1f && r[1] == 0x8b {
		return r
	}
	return String2bytes(src)
}

func ConvertString(b []byte) *string {
	//太短的情况下直接返回字符串，不可能是gzip压缩
	if len(b) < 2 {
		return Bytes2Str(b)
	}

	//gzip压缩之后二进制要转为base64
	if b[0] == 0x1f && b[1] == 0x8b {
		//以下通过直接调用内部代码，减少字符串生成
		enc := base64.StdEncoding
		buf := make([]byte, enc.EncodedLen(len(b)))
		enc.Encode(buf, b)
		return Bytes2Str(buf)
	} else {
		return Bytes2Str(b)
	}
}
