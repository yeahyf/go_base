package httputil

import (
	"net/http"
	"strconv"

	"github.com/yeahyf/go_base/immut"

	"github.com/gogo/protobuf/proto"
	"github.com/yeahyf/go_base/ept"
	"github.com/yeahyf/go_base/log"
)

// Wrapper 以下方法是一种对错误统一处理的封装
type Wrapper func(w http.ResponseWriter, r *http.Request) (proto.Message, error)

func Handler(httpWrapper Wrapper) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer ept.PanicHandle()
		if pb, err := httpWrapper(w, r); err != nil {
			ExRespHandler(w, err)
		} else if pb != nil {
			RespHandler(w, pb)
		}
	}
}

// ExRespHandler 像客户端输出错误信息
func ExRespHandler(w http.ResponseWriter, err error) {
	log.Error("Code="+strconv.Itoa(int(err.(*ept.Error).Code)), ", Info="+err.(*ept.Error).Message)
	w.Header().Add(HeadServerEx, "1")
	resp := &ept.ErrorResponse{
		Code: err.(*ept.Error).Code,
		Info: err.(*ept.Error).Message,
	}
	data, _ := proto.Marshal(resp)
	w.Write(data)
}

// ReqHandle 组合处理
func ReqHandle(w *http.ResponseWriter, r *http.Request, commonCache *CommonCache, pb proto.Message) error {
	postData, err := ReqHeadHandle(r, commonCache)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(postData, pb)
	if err != nil {
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufUn,
			Message: "AtAdLoginRequest Unmarshal Error!!!",
		}
		return aErr
	}
	if log.IsDebug() {
		log.Debugf("req = %s", pb)
	}
	return nil
}

func RespHandler(w http.ResponseWriter, pb proto.Message) {
	if log.IsDebug() {
		log.Debugf("Resp = %s", pb)
	}

	result, err := proto.Marshal(pb)
	if err != nil {
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufMa,
			Message: "Protobuf Ma Failed!!!",
		}
		ExRespHandler(w, aErr)
		return
	}
	w.Write(result)
}
