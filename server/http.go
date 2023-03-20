package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yeahyf/go_base/utils"
)

// GetPostData 从Post请求中获取数据
func GetPostData(r *http.Request) ([]byte, error) {
	//拿到数据就关闭掉
	defer utils.CloseAction(r.Body)

	length := r.ContentLength
	var b bytes.Buffer
	b.Grow(int(length))
	_, err := io.Copy(&b, r.Body)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
