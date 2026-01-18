package cfg

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/yeahyf/go_base/log"
)

var p *Properties

// Properties 构建存储对象
type Properties struct {
	values map[string]string
}

// NewProperties 创建新的存储对象
func NewProperties() *Properties {
	p := &Properties{
		values: make(map[string]string),
	}
	return p
}

// Load 加载方法
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
	if p == nil {
		return ""
	}
	return p.values[key]
}

//=============================================

// Load 加载配置函数
func Load(configPath *string) {
	data, err := os.ReadFile(*configPath)
	if err != nil {
		panic(err)
	}
	p = NewProperties()
	err = p.Load(bytes.NewReader(data))
	if err != nil {
		log.Errorf("couldn't load config file, %v", err)
	}
}

// GetString 获取字符串
func GetString(key string) string {
	return p.Get(key)
}

// GetInt 获取整形
func GetInt(key string) int {
	s := p.Get(key)
	value, err := strconv.Atoi(s)
	if err != nil {
		log.Errorf("couldn't get key(%s) value, %v", key, err)
		return 0
	}
	return value
}

func GetInt64(key string) int64 {
	s := p.Get(key)
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Errorf("couldn't get key(%s) value, %v", key, err)
		return 0
	}
	return value
}

// GetBool 获取布尔型
func GetBool(key string) bool {
	value, err := strconv.ParseBool(p.Get(key))
	if err != nil {
		log.Errorf("couldn't get key(%s) value, %v", key, err)
		return false
	}
	return value
}

// GetIntArray 获取整形数组
func GetIntArray(key string) []int {
	s := p.Get(key)
	array := strings.Split(s, ",")
	r := make([]int, len(array))
	var err error
	for k, v := range array {
		r[k], err = strconv.Atoi(v)
		if err != nil {
			log.Errorf("couldn't get key(%s) value, %v", key, err)
		}
	}
	return r
}

// GetStringArray 获取字符串数组
func GetStringArray(key string) []string {
	s := p.Get(key)
	return strings.Split(s, ",")
}
