///特定的http处理工具
package httputil

import (
	"crypto/sha1"
	"fmt"
	"gobase/cache"
	"gobase/crypto"
	"gobase/ept"
	"gobase/immut"
	"gobase/log"
	"gobase/strutil"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
)

const (
	Head_ContentType = "Content-Type"
	Head_X_Vesion    = "X-Version"
	Head_X_Nonce     = "X-Nonce"
	Head_X_Timestamp = "X-Timestamp"
	Head_X_Signature = "X-Signature"
	Head_IP          = "X-Real-IP"
	Http_Post        = "POST"

	Head_Server_Ex = "X-Server-Ex"

	CT_Protobuf     = "application/x-protobuf"
	CT_Json         = "application/json"
	Head_User_Agent = "User-Agent"
)

///组合处理
func HttpHandle(w *http.ResponseWriter, r *http.Request) []byte {
	postData, err := ReqPreHandle(r, nil)
	if err != nil {
		ExRespHandle(w, err)
		return nil
	}
	return postData
}

//像客户端输出错误信息
func ExRespHandle(w *http.ResponseWriter, err error) {
	(*w).Header().Add(Head_Server_Ex, "1")
	resp := &ept.ErrorResponse{
		Code: err.(*ept.Error).Code,
		Info: err.(*ept.Error).Message,
	}
	data, _ := proto.Marshal(resp)
	(*w).Write(data)
}

///对http请求进行通用处理
func ReqPreHandle(r *http.Request, cache *cache.RedisPool) ([]byte, error) {
	// 对请求方法做判断
	if r.Method != Http_Post {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_Version,
			Message: "Req Head Version Error!!!",
		}
	}

	//判断请求头信息
	version := r.Header.Get(Head_X_Vesion)
	if version == immut.Blank_String {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_Version,
			Message: "Req Head Version Error!!!",
		}
	}

	nonce := r.Header.Get(Head_X_Nonce)
	if nonce == immut.Blank_String {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_Nonce,
			Message: "Req Head Nonce Error!!!",
		}
	}

	timestamp := r.Header.Get(Head_X_Timestamp)
	if timestamp == immut.Blank_String {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_TS,
			Message: "Req Head Timestamp Error!!!",
		}
	}

	signature := r.Header.Get(Head_X_Signature)
	if signature == immut.Blank_String {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_Signature,
			Message: "Req Head Signature Error!!!",
		}
	}
	//====================================

	//从nginx转发过来的ip地址
	addr := r.Header.Get(Head_IP)
	if addr == "" {
		//部分情况下是直接请求
		if r.RemoteAddr != "" {
			addr = strings.Split(r.RemoteAddr, ":")[0]
		}
	}

	if log.IsDebug() {
		log.Debug("ver=", version)
		log.Debug("nonce=", nonce)
		log.Debug("ts=", timestamp)
		log.Debug("sn=", signature)
		log.Debug("ip=", addr)
	}

	defer r.Body.Close()
	postData, err := ioutil.ReadAll(r.Body) //获取post的数据

	if err != nil {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_ReadIO,
			Message: "Read Post Data Error!!!",
		}
	}

	postDataMD5 := crypto.MD54Bytes(postData)

	l := make([]*string, 0, 3)
	l = append(l, postDataMD5)
	l = append(l, &nonce)
	l = append(l, &timestamp)

	//排序
	strutil.SortString(l)

	var builder strings.Builder
	builder.WriteString(*l[0])
	builder.WriteByte('&')
	builder.WriteString(*l[1])
	builder.WriteByte('&')
	builder.WriteString(*l[2])

	source := builder.String()

	if log.IsDebug() {
		log.Debugf("Befor Sginatrue str = %s", source)
	}

	//对数据进行SHA1摘要处理
	h := sha1.New()
	io.WriteString(h, source)
	Signature := fmt.Sprintf("%x", h.Sum(nil))
	h = nil
	builder.Reset()

	if log.IsDebug() {
		log.Debugf("After Singature str = %s", Signature)
	}

	//对比摘要
	if Signature != signature {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_Signature,
			Message: "Signatrue Data Error!!!",
		}
	}

	//对时间进行处理
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_TS,
			Message: "Timestampt Error!!!",
		}
	}
	tm := time.Unix(ts, 0)
	//超过3分钟,过期请求
	if time.Now().Sub(tm) > time.Duration(3*time.Minute) {
		return nil, &ept.Error{
			Code:    immut.Code_Ex_TS,
			Message: "Timestampt Error!!!",
		}
	}

	if cache != nil {
		//nonce
		value, err := cache.GetValue(&nonce)
		if err != nil {
			return nil, &ept.Error{
				Code:    immut.Code_Ex_Redis,
				Message: "Read Redis Data Error!!!",
			}
		}
		if value != nil { //存在值，说明已经提交过了
			return nil, &ept.Error{
				Code:    immut.Code_Ex_Repeat_Req,
				Message: "Req Repeat Error!!!",
			}
		} else { //说明里边没有值
			value := ""
			cache.SetValue(&nonce, &value, 5*60) //180秒
		}
	}
	return postData, nil
}
