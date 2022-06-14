package httputil

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

	"github.com/yeahyf/go_base/cfg"
	"github.com/yeahyf/go_base/utils"

	"github.com/yeahyf/go_base/cache"
	"github.com/yeahyf/go_base/crypto"
	"github.com/yeahyf/go_base/ept"
	"github.com/yeahyf/go_base/immut"
	"github.com/yeahyf/go_base/log"
	"github.com/yeahyf/go_base/strutil"
	"google.golang.org/protobuf/proto"
)

const (
	HeadContentEncoding = "Content-Encoding"
	HeadXVersion        = "X-Version"
	HeadXNonce          = "X-Nonce"
	HeadXTimestamp      = "X-Timestamp"
	HeadXSignature      = "X-Signature"
	HeadIp              = "X-Real-IP"
	HttpPost            = "POST"
	HeadServerEx        = "X-Server-Ex"
	EncodingType        = "gzip"
	HeadXAppKey         = "X-AppKey"

	HeadUserAgent = "User-Agent"
)

type CommonCache struct {
	ReadCache  *cache.RedisPool
	WriteCache *cache.RedisPool
}

func HttpReqHandle(w http.ResponseWriter, r *http.Request,
	commonCache *CommonCache, pb proto.Message) bool {
	postData, err := ReqHeadHandle(r, commonCache)
	if err != nil {
		ExceptionRespHandle(w, err)
		return false
	}

	err = proto.Unmarshal(postData, pb)
	if err != nil {
		log.Errorf("proto couldn't unmarshal type = %s info = %v"+
			reflect.TypeOf(pb).Name(), err)
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufUn,
			Message: "unmarshal error!!!",
		}
		ExceptionRespHandle(w, aErr)
		return false
	}
	if log.IsDebug() {
		log.Debugf("req = %s", pb)
	}
	return true
}

//ReqHeadHandle 从Http请求中获取上报数据，只支持Post
func ReqHeadHandle(r *http.Request, commonCache *CommonCache) ([]byte, error) {
	// 对请求方法做判断
	if r.Method != HttpPost {
		return nil, &ept.Error{
			Code:    immut.CodeExHttpMethod,
			Message: "only support post method",
		}
	}

	//判断请求头信息
	version := r.Header.Get(HeadXVersion)
	if version == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExVersion,
			Message: "couldn't read head x-version",
		}
	}

	if ver, err := strconv.ParseFloat(version, 32); err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExVersion,
			Message: "x-version error",
		}
	} else if ver >= 2.0 { //判断版本大于等于2.0 开启appkey白名单校验
		//appkey不能为空
		appkey := r.Header.Get(HeadXAppKey)
		if appkey == immut.Blank {
			return nil, &ept.Error{
				Code:    immut.CodeExAppKey,
				Message: "could read req head appkey",
			}
		}

		//白名单校验
		if !cfg.CheckAppKey(appkey) {
			if !strings.HasPrefix(appkey, "5ska3upf") {
				log.Errorf("wrongful appkey: %s", appkey)
			}
			return nil, &ept.Error{
				Code:    immut.CodeExAppKey,
				Message: "wrongful appkey",
			}
		}
	}

	nonce := r.Header.Get(HeadXNonce)
	if nonce == immut.Blank {
		return nil, &ept.Error{
			Code:    immut.CodeExNonce,
			Message: "couldn't req head nonce",
		}
	}

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
	Signature := fmt.Sprintf("%x", h.Sum(nil))
	h = nil

	if log.IsDebug() {
		log.Debugf("after signature str = %s", Signature)
	}
	//对比摘要
	if Signature != signature {
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
	duration := time.Now().Sub(tm)
	if duration > 3*time.Minute {
		return nil, &ept.Error{
			Code:    immut.CodeExTs,
			Message: "ts duration error!!! duration=" + duration.String(),
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
			_ = commonCache.WriteCache.SetValue(&nonce, &value, 5*60) //180秒
		}
	}
	//无压缩
	if encoding != EncodingType {
		return buffer.Bytes(), nil
	}
	var gzipReader *gzip.Reader
	gzipReader, err = gzip.NewReader(buffer)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.CodeExReadIO,
			Message: "couldn't gunzip data" + err.Error(),
		}
	}
	defer utils.CloseAction(gzipReader)
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
				Message: "couldn't read gzip data" + err.Error(),
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
	postData := buf.Bytes()
	return postData, nil
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

func HttpRespHandle(w http.ResponseWriter, pb proto.Message) {
	if log.IsDebug() {
		log.Debugf("resp = %s", pb)
	}
	result, err := proto.Marshal(pb)
	if err != nil {
		aErr := &ept.Error{
			Code:    immut.CodeExProtobufMa,
			Message: "proto couldn't marshal",
		}
		ExceptionRespHandle(w, aErr)
		return
	}
	_, _ = w.Write(result)
}

//ExceptionRespHandle 向客户端输出错误信息
func ExceptionRespHandle(w http.ResponseWriter, err error) {
	w.Header().Add(HeadServerEx, "1")
	var resp proto.Message
	if eptError, ok := err.(*ept.Error); ok {
		//1008 为appkey错误
		if eptError.Code != 1008 {
			log.Errorf("code:%s, %s", strconv.Itoa(int(eptError.Code)), eptError.Message)
		}
		resp = &ept.ErrorResponse{
			Code: err.(*ept.Error).Code,
			Info: err.(*ept.Error).Message,
		}
	} else {
		resp = &ept.ErrorResponse{
			Code: 1,
			Info: err.Error(),
		}
		log.Errorf("other exp info=%v", err.Error())
	}
	data, _ := proto.Marshal(resp)
	_, _ = w.Write(data)
}
