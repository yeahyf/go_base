package crypto

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/yeahyf/go_base/strutil"
)

// MD5 获取字符串的MD5签名
func MD5(str *string) *string {
	h := md5.New()
	h.Write(strutil.String2bytes(str))
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

// MD5S 获取字符串的MD5签名
func MD5S(str string) string {
	h := md5.New()
	h.Write(strutil.Str2bytes(str))
	return hex.EncodeToString(h.Sum(nil))
}

// MD54Bytes 获取字节数组的md5值
func MD54Bytes(str []byte) *string {
	h := md5.New()
	h.Write(str)
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

// MD54Bs 获取字节数组的md5值
func MD54Bs(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}
