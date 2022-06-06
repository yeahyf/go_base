package crypto

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/yeahyf/go_base/strutil"
)

//MD5 获取字符串的MD5签名
func MD5(str *string) *string {
	h := md5.New()
	h.Write(strutil.String2bytes(str))
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

func MD54Bytes(str []byte) *string {
	h := md5.New()
	h.Write(str)
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}
