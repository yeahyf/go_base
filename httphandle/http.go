package httphandle

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/yeahyf/go_base/cache"
	"github.com/yeahyf/go_base/crypto"
	"github.com/yeahyf/go_base/ept"
	"github.com/yeahyf/go_base/immut"
	"github.com/yeahyf/go_base/log"
	"github.com/yeahyf/go_base/strutil"
	"github.com/yeahyf/go_base/utils"
	"google.golang.org/protobuf/proto"
)

const (
	HeadContentEncoding = "Content-Encoding"
	HeadXVersion        = "X-Version" //使用版本1.0，但是不做检查
	HeadXNonce          = "X-Nonce"
	HeadXTimestamp      = "X-Timestamp"
	HeadXSignature      = "X-Signature"
	HttpPost            = "POST"
	HeadServerEx        = "X-Server-Ex"
	EncodingType        = "gzip"
	HeadXAppKey         = "X-AppKey"
)

type CommonCache struct {
	ReadCache  *cache.RedisPool
	WriteCache *cache.RedisPool
}

type ReqData struct {
	ReqBody []byte
	Nonce   string
	Appkey  string
}

// Wrapper 对基本护理逻辑的封装
type Wrapper func(pb proto.Message) (proto.Message, error)

// IsRepeatReq 对请求进行重复检查,返回true表示重复请求,false表示无重复
type IsRepeatReq func(nonce string) bool

// IsValidAppKey 对Appkey进行检查,true为有效,false为无效
type IsValidAppKey func(appkey string) bool

// AbstractHandler 对业务逻辑的基本封装
func AbstractHandler(httpWrapper Wrapper, repeatCheck IsRepeatReq, appKeyCheck IsValidAppKey,
	reqPb proto.Message) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer ept.PanicHandle()

		reqData, err := ReqBaseCheck(r)
		if err != nil {
			ExRespHandler(w, err)
			return
		}

		if repeatCheck != nil {
			if result := repeatCheck(reqData.Nonce); result {
				err = &ept.Error{
					Code:    immut.CodeExRepeatReq,
					Message: "repeat req error",
				}
				ExRespHandler(w, err)
				return
			}
		}

		if appKeyCheck != nil {
			if !appKeyCheck(reqData.Appkey) {
				err := &ept.Error{
					Code:    immut.CodeExAppKey,
					Message: "wrongful appkey",
				}
				ExRespHandler(w, err)
				return
			}
		}
		err = proto.Unmarshal(reqData.ReqBody, reqPb)
		if err != nil {
			log.Errorf("couldn't unmarshal type = %s info = %v", reflect.TypeOf(reqPb).Name(), err)
			aErr := &ept.Error{
				Code:    immut.CodeExProtobufUn,
				Message: "unmarshal error",
			}
			ExRespHandler(w, aErr)
			return
		}
		if log.IsDebug() {
			log.Debugf("req = %s", reqPb)
		}
		if respPb, err := httpWrapper(reqPb); err != nil {
			ExRespHandler(w, err)
		} else if respPb != nil {
			if log.IsDebug() {
				log.Debugf("resp = %s", respPb)
			}
			RespHandler(w, respPb)
		}
	}
}

// ExRespHandler 异常响应处理
func ExRespHandler(w http.ResponseWriter, err error) {
	log.Error("code="+strconv.Itoa(int(err.(*ept.Error).Code)), ", info="+err.(*ept.Error).Message)
	w.Header().Add(HeadServerEx, "1")
	resp := &ept.ErrorResponse{
		Code: err.(*ept.Error).Code,
		Info: err.(*ept.Error).Message,
	}
	data, _ := proto.Marshal(resp)
	w.Write(data)
}

func ReqBaseCheck(r *http.Request) (*ReqData, error) {
	// 对请求方法做判断
	if r.Method != HttpPost {
		return nil, &ept.Error{
			Code:    immut.CodeExHttpMethod,
			Message: "must http post",
		}
	}

	//判断请求头信息
	version := r.Header.Get(HeadXVersion)
	if version == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExVersion,
			Message: "couldn't read head x-version",
		}
	} else {
		//
		if _, err := strconv.ParseFloat(version, 32); err != nil {
			return nil, &ept.Error{
				Code:    immut.CodeExVersion,
				Message: "x-version error",
			}
		}
	}

	var reqData ReqData
	//appkey不能为空
	appkey := r.Header.Get(HeadXAppKey)
	if appkey == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExAppKey,
			Message: "couldn't read req head appkey",
		}
	}

	reqData.Appkey = appkey

	nonce := r.Header.Get(HeadXNonce)
	if nonce == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExNonce,
			Message: "couldn't req head nonce",
		}
	}

	reqData.Nonce = nonce

	timestamp := r.Header.Get(HeadXTimestamp)
	if timestamp == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "couldn't read req head ts",
		}
	}

	signature := r.Header.Get(HeadXSignature)
	if signature == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExSignature,
			Message: "couldn't read head signature",
		}
	}

	encoding := r.Header.Get(HeadContentEncoding)
	if log.IsDebug() {
		log.Debug("ts=", timestamp)
		log.Debug("sn=", signature)
		log.Debug("encoding=", encoding)
	}

	//获取post的数据
	buffer := getPostData(r)
	postDataMD5 := crypto.MD54Bytes(buffer.Bytes())
	l := make([]string, 0, 3)
	l = append(l, *postDataMD5)
	l = append(l, nonce)
	l = append(l, timestamp)
	//排序
	strutil.SortString(l)
	var builder strings.Builder
	builder.WriteString(l[0])
	builder.WriteByte('&')
	builder.WriteString(l[1])
	builder.WriteByte('&')
	builder.WriteString(l[2])
	source := builder.String()
	if log.IsDebug() {
		log.Debugf("before signature str = %s", source)
	}

	//对数据进行SHA1摘要处理
	h := sha1.New()
	_, _ = io.WriteString(h, source)
	sha1Value := fmt.Sprintf("%x", h.Sum(nil))
	h = nil

	if log.IsDebug() {
		log.Debugf("after signature str = %s", sha1Value)
	}
	//对比摘要
	if sha1Value != signature {
		return nil, &ept.Error{
			Code:    immut.CodeExSignature,
			Message: "signature data error!",
		}
	}

	//对时间进行处理
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "ts format Error!!!",
		}
	}

	tm := time.Unix(ts, 0)
	//超过3分钟,过期请求
	duration := time.Since(tm)
	if duration > 3*time.Minute {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "ts duration error!!! duration=" + duration.String(),
		}
	}

	//无压缩
	if encoding != EncodingType {
		reqData.ReqBody = buffer.Bytes()
		return &reqData, nil
	}

	//解压缩
	var gzipReader *gzip.Reader
	gzipReader, err = gzip.NewReader(buffer)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExReadIO,
			Message: "couldn't gunzip data" + err.Error(),
		}
	}
	defer utils.CloseAction(gzipReader)

	var buf bytes.Buffer
	compressSize := buffer.Len()
	//默认为5倍的压缩大小
	buf.Grow(compressSize * 5)
	//判断缓存区的大小
	cacheSize := 4096
	if compressSize > 1024 {
		cacheSize = 8092
	}
	//建立缓存区
	p := make([]byte, cacheSize)
	for {
		n, err := gzipReader.Read(p)
		if err != nil {
			//请注意读取到unexpected EOF也是可以将数据读取完整的
			if strings.Contains(err.Error(), "EOF") {
				if n != 0 {
					buf.Write(p[:n])
					continue
				} else {
					break
				}
			} else {
				return nil, &ept.Error{
					Code:    immut.CodeExReadIO,
					Message: "couldn't read gzip data" + err.Error(),
				}
			}
		}
		//读取到的数据如果不满，不一定代表结束,需要接收数据之后，继续读取
		if n != cacheSize {
			buf.Write(p[:n])
			continue
		}
		buf.Write(p)
	}
	reqData.ReqBody = buf.Bytes()
	return &reqData, nil
}

func getPostData(r *http.Request) *bytes.Buffer {
	//拿到数据就关闭掉
	defer utils.CloseAction(r.Body)

	length := r.ContentLength
	var b bytes.Buffer
	b.Grow(int(length))
	_, _ = io.Copy(&b, r.Body)
	return &b
}

func RespHandler(w http.ResponseWriter, pb proto.Message) {
	if pb == nil {
		w.Write([]byte(""))
		return
	}
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
