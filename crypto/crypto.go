package crypto

import (
	"crypto/md5"

	"encoding/hex"

	"github.com/yeahyf/go_base/strutil"
)

///
/// 重要提示：
/// MD5 算法已被废弃，不建议使用。
/// 安全性较高的建议使用 SHA-256、SHA-3 等哈希算法
/// 密码保存的时候使用 Bcrypt、Argon2
/// 请参考 crypto/crypto_new.go中的相关函数
///

// MD5 获取字符串的MD5签名
// @deprecated MD5已经被废弃，不建议使用
func MD5(str *string) *string {
	h := md5.New()
	h.Write(strutil.String2bytes(str))
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

// MD5S 获取字符串的MD5签名
// @deprecated MD5已经被废弃，不建议使用
func MD5S(str string) string {
	h := md5.New()
	h.Write(strutil.Str2bytes(str))
	return hex.EncodeToString(h.Sum(nil))
}

// MD54Bytes 获取字节数组的md5值
// @deprecated MD5已经被废弃，不建议使用
func MD54Bytes(str []byte) *string {
	h := md5.New()
	h.Write(str)
	result := hex.EncodeToString(h.Sum(nil))
	return &result
}

// MD54Bs 获取字节数组的md5值
// @deprecated MD5已经被废弃，不建议使用
func MD54Bs(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}
