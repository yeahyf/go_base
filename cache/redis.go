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
	Set    = "SET"
	Get    = "GET"
	DEL    = "DEL"
	EXISTS = "EXISTS"

	MGet = "MGET"
	MSet = "MSET"

	HGetAll = "HGETALL"

	Expire = "EXPIRE"
	SetEx  = "SETEX"

	Multi  = "MULTI"
	Select = "SELECT"
	Exec   = "EXEC"
)

var (
	ErrGetConn  = errors.New("get redis conn error")
	ErrGetValue = errors.New("get redis data exception")
)

//type RedisPool = redis.Pool

type RedisPool struct {
	*redis.Pool // 创建redis连接池
	//DBIndex     int
}

// Config 配置参数
type Config struct {
	InitConnSize int    // 初始化连接数量
	MaxConnSize  int    // 最大连接数
	MaxIdleTime  int    // 连接最大空闲时间
	Address      string // 服务器地址
	Username     string // 账号名称，如没有设置为空
	Password     string // 密码
	DBIndex      int    // 使用的 DB ID
}

// newPool 初始化 Redis 连接池
func newPool(cfg *Config) *RedisPool {
	redisPool := &redis.Pool{
		// 实例化一个连接池
		MaxIdle:     cfg.InitConnSize,                             // 最初的连接数量
		MaxActive:   cfg.MaxConnSize,                              // 连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		Wait:        true,                                         // 没有连接可用需要等待
		IdleTimeout: time.Second * time.Duration(cfg.MaxIdleTime), // 连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { // 要连接的redis数据库
			if cfg.Password == immut.Blank {
				c, err := redis.Dial("tcp", cfg.Address, redis.DialDatabase(cfg.DBIndex))
				if err != nil {
					fmt.Println("couldn't create conn ", err)
					return nil, err
				}
				return c, nil
			} else {
				var err error
				var c redis.Conn
				if cfg.Username == immut.Blank {
					c, err = redis.Dial("tcp", cfg.Address, redis.DialPassword(cfg.Password), redis.DialDatabase(cfg.DBIndex))
				} else {
					c, err = redis.Dial("tcp", cfg.Address,
						redis.DialUsername(cfg.Username), redis.DialPassword(cfg.Password), redis.DialDatabase(cfg.DBIndex))
				}
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
		//cfg.DBIndex,
	}
}

// NewPool 根据 cfg 来设置, 推荐使用
func NewPool(cfg *Config) *RedisPool {
	return newPool(cfg)
}

// NewRedisPoolByDB 构建新的Redis连接池
func NewRedisPoolByDB(init, maxsize, idle int, address, password string, dbIndex int) *RedisPool {
	cfg := &Config{
		init,
		maxsize,
		idle,
		address,
		"",
		password,
		dbIndex,
	}
	return newPool(cfg)
}

// NewRedisPool 构建新的Redis连接池，放入默认的0号DB中
func NewRedisPool(init, maxsize, idle int, address, password string) *RedisPool {
	return NewRedisPoolByDB(init, maxsize, idle, address, password, 0) // 默认选择0号库
}

// SetValue expire的单位为秒 默认DB索引为DB初始化设置
func (p *RedisPool) SetValue(key string, value string, expire int) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)

	var err error
	if expire > 0 {
		_, err = c.Do(SetEx, key, expire, value)
	} else {
		_, err = c.Do(Set, key, value)
	}

	return err
}

// DeleteValues 删除多个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValues(keys []string) (int, error) {
	c := p.Get()
	if c == nil {
		return 0, ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	for _, key := range keys {
		RedisSend(c, DEL, key)
	}

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return 0, err
	}
	// 统计所有删除的数量
	total := 0
	length := len(replay)
	for i := range length {
		count, err := redis.Int(replay[i], nil)
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

// DeleteValue 删除一个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValue(key string) (int, error) {
	c := p.Get()
	if c == nil {
		return 0, errors.New("get redis conn error")
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	return redis.Int(c.Do(DEL, key))
}

// GetValue 从Redis中获取指定的值
func (p *RedisPool) GetValue(key string) (string, error) {
	c := p.Get()
	if c == nil {
		return "", ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	value, err := redis.String(c.Do(Get, key))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// ExistsValue 判断某个 Key 是否有缓存
func (p *RedisPool) ExistsValue(key string) (bool, error) {
	c := p.Get()
	if c == nil {
		return false, ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	replay, err := redis.Int(c.Do(EXISTS, key))
	if err != nil {
		return false, err
	}
	return replay == 1, nil
}

// HGetAllValue 从Redis中获取指定的值
func (p *RedisPool) HGetAllValue(key string) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, errors.New("can not get redis conn")
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	value, err := redis.Strings(c.Do(HGetAll, key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// 如果长度为0,说明没有值
	if len(value) == 0 {
		return nil, nil
	}
	return value, nil
}

// MGetValue 一次性获取多个Key的值
func (p *RedisPool) MGetValue(keys []any) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	value, err := redis.Strings(c.Do(MGet, keys...))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// 有可能都没有值,但是返回值仍旧会封装成slice
	// 需要判断slice里边的具体的值
	return value, nil
}

// MSetValue 批量设置
func (p *RedisPool) MSetValue(kv []any) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	_, err := c.Do(MSet, kv...)
	if err != nil {
		return err
	}
	return nil
}

// SetExpire 设置过期时间
func (p *RedisPool) SetExpire(key string, expire int) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	_, err := c.Do(Expire, key, expire)
	return err
}

// MSetValueWithExpire 批量K-V以及对应的过期时间
// kv中的值需要按照 k1,v1,k2,v2,k3,v3 ... 进行存储
func (p *RedisPool) MSetValueWithExpire(kv []any, expire int) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	length := len(kv)

	RedisSend(c, Multi)
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
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
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
	Send(commandName string, args ...any) error
}

func RedisSend(s Send, action string, args ...any) {
	var err error
	if len(args) > 0 {
		// 注意此处args的写法,args是切片,不是可变参数
		err = s.Send(action, args...)
	} else {
		err = s.Send(action)
	}
	if err != nil {
		log.Errorf("couldn't exec redis %s, param %v, %v",
			action, args, err)
	}
}
