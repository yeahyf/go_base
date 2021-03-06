///部分加解密的简单封装
package crypto

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/yeahyf/go_base/strutil"
)

// // AES for CBC
// func padding(src []byte, blocksize int) []byte {
// 	padnum := blocksize - len(src)%blocksize
// 	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
// 	return append(src, pad...)
// }

// func unpadding(src []byte) []byte {
// 	n := len(src)
// 	unpadnum := int(src[n-1])
// 	return src[:n-unpadnum]
// }

// func EncryptCBCAES(src []byte, key []byte) []byte {
// 	block, _ := aes.NewCipher(key)
// 	src = padding(src, block.BlockSize())
// 	blockmode := cipher.NewCBCEncrypter(block, key)
// 	blockmode.CryptBlocks(src, src)
// 	return src
// }

// func DecryptCBCAES(src []byte, key []byte) []byte {
// 	block, _ := aes.NewCipher(key)
// 	blockmode := cipher.NewCBCDecrypter(block, key)
// 	blockmode.CryptBlocks(src, src)
// 	src = unpadding(src)
// 	return src
// }

//获取字符串的MD5签名
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
