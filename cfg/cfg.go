///提供基本的配置管理接口
package cfg

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/yeahyf/go_base/log"
)

var p *Properties

//构建存储对象
type Properties struct {
	values map[string]string
}

//创建新的存储对象
func NewProperties() *Properties {
	p := &Properties{
		values: make(map[string]string),
	}
	return p
}

//加载方法
func (p *Properties) Load(r io.Reader) error {
	buf := bufio.NewReader(r)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		line = bytes.TrimSpace(line)
		lenLine := len(line)

		if lenLine == 0 {
			continue
		}
		first := line[0]
		if first == byte('#') || first == byte('!') {
			continue
		}

		sep := []byte{'='}
		index := bytes.Index(line, sep)
		if index < 0 {
			sep = []byte{':'}
		}
		kv := bytes.SplitN(line, sep, 2)
		if kv == nil {
			continue
		}
		lenKV := len(kv)
		if lenKV == 2 {
			p.values[string(bytes.TrimSpace(kv[0]))] = string(bytes.TrimSpace(kv[1]))
		}
	}
}

func (p *Properties) Get(key string) string {
	return p.values[key]
}

//=============================================

//加载配置函数
func Load(configPath *string) {
	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	p = NewProperties()
	err = p.Load(bytes.NewReader(data))
}

//获取字符串
func GetString(key string) string {
	return p.Get(key)
}

//获取整形
func GetInt(key string) int {
	s := p.Get(key)
	value, err := strconv.Atoi(s)
	if err != nil {
		log.Error(err)
		return 0
	}
	return value
}

//获取布尔型
func GetBool(key string) bool {
	value, err := strconv.ParseBool(p.Get(key))
	if err != nil {
		log.Error(err)
		return false
	}
	return value
}

//获取整形数组
func GetIntArray(key string) []int {
	s := p.Get(key)
	array := strings.Split(s, ",")
	r := make([]int, len(array))
	var err error
	for k, v := range array {
		r[k], err = strconv.Atoi(v)
		if err != nil {
			log.Error("Parse Int Array Error,key = ", key, err)
		}
	}
	return r
}

//获取整形数组
func GetStringArray(key string) []string {
	s := p.Get(key)
	return strings.Split(s, ",")
}

//判断appkey是否在白名单
func CheckAppKey(appkey string) bool {
	s := p.Get("appkey.list")
	array := strings.Split(s, ",")
	for _, v := range array {
		if appkey == v {
			return true
		}
	}
	return false
}
