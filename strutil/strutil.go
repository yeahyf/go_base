package strutil

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"sort"
	"unsafe"
)

// Bytes2str byte数组转化为字符串，返回字符串
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes2Str byte数组转化为字符串，返回字符串引用
func Bytes2Str(b []byte) *string {
	return (*string)(unsafe.Pointer(&b))
}

// Str2bytes 字符串转化为byte数组
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s)) // 获取s的起始地址开始后的两个uintptr指针
	h := [3]uintptr{x[0], x[1], x[1]}      // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

// String2bytes 字符串转化为byte数组
func String2bytes(s *string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}     // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

func SortString(list []string) {
	sort.Sort(sort.StringSlice(list))
}

// IsBase64String 如果是gzip压缩，就转base64
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

// ConvertBytes 将字符串指针转为byte数组
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

// ConvertString 将字byte数组转为字符串指针
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

// StrByXOR 对字符串进行异或操作
func StrByXOR(message []byte, keywords []byte) []byte {
	messageLen := len(message)
	keywordsLen := len(keywords)

	result := make([]byte, 0, messageLen)
	for i := 0; i < messageLen; i++ {
		result = append(result, message[i]^keywords[i%keywordsLen])
	}
	return result
}

// Gunzip 解压缩数据
func Gunzip(source []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(source))
	//如果解压缩失败
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, gzipReader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
