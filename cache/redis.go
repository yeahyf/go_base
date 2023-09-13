package cache

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/yeahyf/go_base/log"

	"github.com/yeahyf/go_base/immut"
)

const (
	Set = "SET"
	Get = "GET"
	DEL = "DEL"

	MGet = "MGET"
	MSet = "MSET"

	HGetAll = "HGETALL"

	Expire = "EXPIRE"
	SetEx  = "SETEX"

	Multi  = "MULTI"
	Select = "SELECT"
	Exec   = "EXEC"
)

var getConnErr = errors.New("get redis conn error")
var valueErr = errors.New("get redis data exception")

type RedisPool struct {
	*redis.Pool //创建redis连接池
	DBIndex     int
}

// NewRedisPoolByDB 构建新的Redis连接池
func NewRedisPoolByDB(init, maxsize, idle int, address, password string, dbIndex int) *RedisPool {
	redisPool := &redis.Pool{
		//实例化一个连接池
		MaxIdle:     init,                              //最初的连接数量
		MaxActive:   maxsize,                           //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		Wait:        true,                              //没有连接可用需要等待
		IdleTimeout: time.Second * time.Duration(idle), //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			if password == immut.Blank {
				c, err := redis.Dial("tcp", address)
				if err != nil {
					fmt.Println("couldn't create conn ", err)
					return nil, err
				}
				return c, nil
			} else {
				c, err := redis.Dial("tcp", address, redis.DialPassword(password))
				if err != nil {
					fmt.Println("couldn't create conn ", err)
					return nil, err
				}
				return c, nil
			}
		},
	}
	return &RedisPool{
		redisPool,
		dbIndex,
	}
}

// NewRedisPool 构建新的Redis连接池，放入默认的0号DB中
func NewRedisPool(init, maxsize, idle int, address, password string) *RedisPool {
	return NewRedisPoolByDB(init, maxsize, idle, address, password, 0) //默认选择0号库
}

// SetValue expire的单位为秒 默认DB索引为DB初始化设置
func (p *RedisPool) SetValue(key string, value string, expire int) error {
	return p.SetValueWithDbIdx(key, value, expire, p.DBIndex)
}

// SetValueWithDbIdx expire的单位为秒 dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) SetValueWithDbIdx(key string, value string, expire, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)

	//使用Send发送指令到缓存区
	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	if expire > 0 {
		RedisSend(c, SetEx, key, expire, value)
	} else {
		RedisSend(c, Set, key, value)
	}
	//使用Do命令执行缓存区的命令
	_, err := c.Do(Exec)
	return err
}

// DeleteValues 删除多个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValues(keys []string) (int, error) {
	return p.DeleteValuesWithDbIdx(keys, p.DBIndex)
}

// DeleteValuesWithDbIdx 删除多个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValuesWithDbIdx(keys []string, dbIdx int) (int, error) {
	c := p.Get()
	if c == nil {
		return 0, getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	for _, key := range keys {
		RedisSend(c, DEL, key)
	}

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return 0, err
	}
	return redis.Int(replay[1], err)
}

// DeleteValue 删除一个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValue(key string) (int, error) {
	return p.DeleteValueWithDbIdx(key, p.DBIndex)
}

// DeleteValueWithDbIdx 删除一个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValueWithDbIdx(key string, dbIdx int) (int, error) {
	c := p.Get()
	if c == nil {
		return 0, errors.New("get redis conn error")
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, DEL, key)

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return 0, err
	}
	return redis.Int(replay[1], err)
}

// GetValue 从Redis中获取指定的值
func (p *RedisPool) GetValue(key string) (string, error) {
	return p.GetValueWithDbIdx(key, p.DBIndex)
}

// GetValueWithDbIdx 从Redis中获取指定的值
func (p *RedisPool) GetValueWithDbIdx(key string, dbIdx int) (string, error) {
	c := p.Get()
	if c == nil {
		return "", getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, Get, key)

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return "", err
	}
	var value string
	value, err = redis.String(replay[1], err)
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		} else {
			return "", err
		}
	}
	return value, err
}

// HGetAllValue 从Redis中获取指定的值
func (p *RedisPool) HGetAllValue(key string) ([]string, error) {
	return p.HGetValueWithDbIdx(key, p.DBIndex)
}

// HGetValueWithDbIdx 从Redis中获取指定的值
func (p *RedisPool) HGetValueWithDbIdx(key string, dbIdx int) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, errors.New("can not get redis conn")
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, HGetAll, key)

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return nil, err
	}
	var value []string
	value, err = redis.Strings(replay[1], err)
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	//如果长度为0,说明没有值
	if len(value) == 0 {
		return nil, nil
	}
	return value, err
}

// MGetValue 一次性获取多个Key的值
func (p *RedisPool) MGetValue(keys []interface{}) ([]string, error) {
	return p.MGetValueWithDbIdx(keys, p.DBIndex)
}

// MGetValueWithDbIdx 一次性获取多个Key的值
func (p *RedisPool) MGetValueWithDbIdx(keys []interface{}, dbIdx int) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, MGet, keys...)

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return nil, err
	}
	var value []string
	value, err = redis.Strings(replay[1], err)
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	//有可能都没有值,但是返回值仍旧会封装成slice
	//需要判断slice里边的具体的值
	return value, err
}

// MSetValue 批量设置
func (p *RedisPool) MSetValue(kv []interface{}) error {
	return p.MSetValueWithDbIdx(kv, p.DBIndex)
}

// MSetValueWithDbIdx 批量设置
func (p *RedisPool) MSetValueWithDbIdx(kv []interface{}, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, MSet, kv...)

	_, err := c.Do(Exec)
	if err != nil {
		return err
	}
	return nil
}

// SetExpire 设置过期时间
func (p *RedisPool) SetExpire(key string, expire int) error {
	return p.SetExpireWithDbIdx(key, expire, p.DBIndex)
}

// SetExpireWithDbIdx 设置过期时间
func (p *RedisPool) SetExpireWithDbIdx(key string, expire, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, Expire, key, expire)

	_, err := c.Do(Exec)
	return err
}

// MSetValueWithExpire 批量K-V以及对应的过期时间
// kv中的值需要按照 k1,v1,k2,v2,k3,v3 ... 进行存储
func (p *RedisPool) MSetValueWithExpire(kv []interface{}, expire int) error {
	return p.MSetValueWithExpireDbIdx(kv, expire, p.DBIndex)
}

// MSetValueWithExpireDbIdx 批量K-V以及对应的过期时间
// kv中的值需要按照 k1,v1,k2,v2,k3,v3 ... 进行存储
func (p *RedisPool) MSetValueWithExpireDbIdx(kv []interface{}, expire, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	length := len(kv)

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	RedisSend(c, MSet, kv...)
	for i := 0; i < length; i += 2 {
		RedisSend(c, Expire, kv[i], expire)
	}

	_, err := c.Do(Exec)
	if err != nil {
		return err
	}
	return nil
}

// MSetExpire 设置过期时间 keys 为key的列表 expire为对应的过期时间，单位为秒
func (p *RedisPool) MSetExpire(keys []string, expire int) error {
	return p.MSetExpireWithDbIdx(keys, expire, p.DBIndex)
}

// MSetExpireWithDbIdx 设置过期时间 keys 为key的列表 expire为对应的过期时间，单位为秒
func (p *RedisPool) MSetExpireWithDbIdx(keys []string, expire, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, dbIdx)
	for _, key := range keys {
		RedisSend(c, Expire, key, expire)
	}

	_, err := c.Do(Exec)
	return err
}

// CloseRedisPool 方便连接池在系统退出的时候也能够优雅的退出
func (p *RedisPool) CloseRedisPool() {
	CloseAction(p)
}

func CloseAction(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Errorf("couldn't close %v", err)
	}
}

type Send interface {
	Send(commandName string, args ...interface{}) error
}

func RedisSend(s Send, action string, args ...interface{}) {
	var err error
	if len(args) > 0 {
		//注意此处args的写法,args是切片,不是可变参数
		err = s.Send(action, args...)
	} else {
		err = s.Send(action)
	}
	if err != nil {
		log.Errorf("couldn't exec redis %s, param %v, %v",
			action, args, err)
	}
}
