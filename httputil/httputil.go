///特定的http处理工具
package httputil

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"github.com/yeahyf/go_base/cfg"
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

	"github.com/gogo/protobuf/proto"
)

const (
	HeadContentType     = "Content-Type"
	HeadContentEncoding = "Content-Encoding"
	HeadXVesion         = "X-Version"
	HeadXNonce          = "X-Nonce"
	HeadXTimestamp      = "X-Timestamp"
	HeadXSignature      = "X-Signature"
	HeadIp              = "X-Real-IP"
	HttpPost            = "POST"
	HeadServerEx        = "X-Server-Ex"
	EncodingType        = "gzip"
	HeadXAppKey         = "X-AppKey"

	CtProtobuf    = "application/x-protobuf"
	CtJson        = "application/json"
	HeadUserAgent = "User-Agent"
)

type CommonCache struct {
	ReadCache  *cache.RedisPool
	WriteCache *cache.RedisPool
}

func HttpReqHandle(w http.ResponseWriter, r *http.Request, commonCache *CommonCache, pb proto.Message) bool {
	postData, err := ReqHeadHandle(r, commonCache)
	if err != nil {
		ExRespHandle(w, err)
		return false
	}

	err = proto.Unmarshal(postData, pb)

	if err != nil {
		log.Errorf("Proto Unmarshal Exception type = %s info = %s !!!"+reflect.TypeOf(pb).Name(), err)
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufUn,
			Message: "Unmarshal Error!!!",
		}
		ExRespHandle(w, aErr)
		return false
	}
	if log.IsDebug() {
		log.Debugf("req = %s", pb)
	}
	return true
}

///对http请求进行通用处理
func ReqHeadHandle(r *http.Request, commonCache *CommonCache) ([]byte, error) {
	// 对请求方法做判断
	if r.Method != HttpPost {
		return nil, &ept.Error{
			Code:    immut.CodeExHttpMethod,
			Message: "Req Method Error!!!",
		}
	}

	//判断请求头信息
	version := r.Header.Get(HeadXVesion)
	if version == immut.BlankString {
		return nil, &ept.Error{
			Code:    immut.CodeExVersion,
			Message: "Req Head Version is NULL Error!!!",
		}
	}

	if ver, err := strconv.ParseFloat(version, 32); err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExVersion,
			Message: "Req Head Version Error!!!",
		}
	} else if ver >= 2.0 { //判断版本大于等于2.0 开启appkey白名单校验
		//appkey不能为空
		appkey := r.Header.Get(HeadXAppKey)
		if appkey == immut.BlankString {
			return nil, &ept.Error{
				Code:    immut.CodeExAppKey,
				Message: "Req Head AppKey is NULL Error!!!",
			}
		}

		//白名单校验
		if !cfg.CheckAppKey(appkey) {
			return nil, &ept.Error{
				Code:    immut.CodeExAppKey,
				Message: "Req Head AppKey not publish Error!!!",
			}
		}
	}

	nonce := r.Header.Get(HeadXNonce)
	if nonce == immut.BlankString {
		return nil, &ept.Error{
			Code:    immut.CodeExNonce,
			Message: "Req Head Nonce Error!!!",
		}
	}

	timestamp := r.Header.Get(HeadXTimestamp)
	if timestamp == immut.BlankString {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "Req Head Timestamp Error!!!",
		}
	}

	signature := r.Header.Get(HeadXSignature)
	if signature == immut.BlankString {
		return nil, &ept.Error{
			Code:    immut.CodeExSignature,
			Message: "Req Head Signature Error!!!",
		}
	}

	encoding := r.Header.Get(HeadContentEncoding)

	if log.IsDebug() {
		//从nginx转发过来的ip地址
		addr := r.Header.Get(HeadIp)
		if addr == "" {
			//部分情况下是直接请求
			if r.RemoteAddr != "" {
				addr = strings.Split(r.RemoteAddr, ":")[0]
			}
		}
		log.Debug("ts=", timestamp)
		log.Debug("sn=", signature)
		log.Debug("ip=", addr)
		log.Debug("encoding=", encoding)
		log.Debug("UserAgent=", r.Header.Get(HeadUserAgent))
		log.Debug("ContentLength=", r.ContentLength)
	}

	//获取post的数据
	buffer := getPostData(r)
	//postData, err := ioutil.ReadAll(r.Body)
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
		log.Debugf("Befor Sginatrue str = %s", source)
	}

	//对数据进行SHA1摘要处理
	h := sha1.New()
	io.WriteString(h, source)
	Signature := fmt.Sprintf("%x", h.Sum(nil))
	h = nil
	//builder.Reset()

	if log.IsDebug() {
		log.Debugf("After Singature str = %s", Signature)
	}

	//对比摘要
	if Signature != signature {
		return nil, &ept.Error{
			Code:    immut.CodeExSignature,
			Message: "Signatrue Data Error!!!",
		}
	}

	//对时间进行处理
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "Timestampt format Error!!!",
		}
	}
	tm := time.Unix(ts, 0)
	//超过3分钟,过期请求
	duration := time.Now().Sub(tm)
	if duration > time.Duration(3*time.Minute) {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "Timestampt duration Error!!! duration=" + duration.String(),
		}
	}

	if commonCache != nil {
		//nonce
		value, err := commonCache.ReadCache.GetValue(&nonce)
		if err != nil {
			return nil, &ept.Error{
				Code:    immut.CodeExRedis,
				Message: "Read Redis Data Error!!!",
			}
		}
		if value != nil { //存在值，说明已经提交过了
			return nil, &ept.Error{
				Code:    immut.CodeExRepeatReq,
				Message: "Req Repeat Error!!!",
			}
		} else { //说明里边没有值
			value := ""
			commonCache.WriteCache.SetValue(&nonce, &value, 5*60) //180秒
		}
	}

	//无压缩
	if encoding != EncodingType {
		return buffer.Bytes(), nil
	}

	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExReadIO,
			Message: "Gunzip Error!!!" + err.Error(),
		}
	}
	defer gzipReader.Close()

	//return ioutil.ReadAll(gzipReader)
	//return postData
	//========================================
	var buf bytes.Buffer
	compressSize := buffer.Len()
	//默认为5倍的压缩大小
	buf.Grow(compressSize * 5)
	//选择开设读取缓存区
	cacheSize := 4096
	if compressSize > 1024 {
		cacheSize = 8092
	}
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
			}
			return nil, &ept.Error{
				Code:    immut.CodeExReadIO,
				Message: "Read Gzip Error!!!" + err.Error(),
			}
		}
		//读取到的数据如果不满，不一定代表结束
		//需要接收数据之后，继续读取
		if n != cacheSize {
			buf.Write(p[:n])
			continue
		}
		buf.Write(p)
	}
	//io.Copy(&buf, reader)
	//postData, err := ioutil.ReadAll(gzipReader)
	postData := buf.Bytes()
	return postData, nil
}

func getPostData(r *http.Request) *bytes.Buffer {
	//拿到数据就关闭掉
	defer r.Body.Close()
	length := r.ContentLength
	var b bytes.Buffer
	b.Grow(int(length))
	io.Copy(&b, r.Body)
	return &b
}

func HttpRespHandle(w http.ResponseWriter, pb proto.Message) {
	if log.IsDebug() {
		log.Debugf("Resp = %s", pb)
	}

	result, err := proto.Marshal(pb)
	if err != nil {
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufMa,
			Message: "Protobuf Ma Failed!!!",
		}
		ExRespHandle(w, aErr)
		return
	}
	w.Write(result)
}

//像客户端输出错误信息
func ExRespHandle(w http.ResponseWriter, err error) {
	w.Header().Add(HeadServerEx, "1")
	var resp proto.Message
	if eptError, ok := err.(*ept.Error); ok {
		log.Error("Code="+strconv.Itoa(int(eptError.Code)), ", Info="+eptError.Message)
		resp = &ept.ErrorResponse{
			Code: err.(*ept.Error).Code,
			Info: err.(*ept.Error).Message,
		}
	} else {
		resp = &ept.ErrorResponse{
			Code: 1,
			Info: err.Error(),
		}
		log.Error("Info=" + err.Error())
	}
	data, _ := proto.Marshal(resp)
	w.Write(data)
}
