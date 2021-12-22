/// 提供Redis的基本管理接口
package cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/yeahyf/go_base/immut"
)

const (
	Set = "SET"
	Get = "GET"
	DEL = "DEL"

	MGet = "MGET"

	Expire = "EXPIRE"
	SetEx  = "SETEX"

	Multi  = "MULTI"
	Select = "SELECT"
	Exec   = "EXEC"
)

type RedisPool struct {
	*redis.Pool //创建redis连接池
	DBIndex     int
}

//构建新的Redis连接池
func NewRedisPoolByDB(init, maxsize, idle int, address, password string, dbIndex int) *RedisPool {
	redisPool := &redis.Pool{
		//实例化一个连接池
		MaxIdle:     init,                              //最初的连接数量
		MaxActive:   maxsize,                           //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		Wait:        true,                              //没有连接可用需要等待
		IdleTimeout: time.Second * time.Duration(idle), //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			if password == immut.BlankString {
				c, err := redis.Dial("tcp", address)
				if err != nil {
					fmt.Println("Create Connection Error!")
					return nil, err
				}
				return c, nil
			} else {
				c, err := redis.Dial("tcp", address, redis.DialPassword(password))
				if err != nil {
					fmt.Println("Create Connection Error!")
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

//构建新的Redis连接池
func NewRedisPool(init, maxsize, idle int, address, password string) *RedisPool {
	return NewRedisPoolByDB(init, maxsize, idle, address, password, 0) //默认选择0号库
}

// SetValueForDBIdx  expire的单位为秒 dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) SetValueForDBIdx(key *string, value *string, expire, dbIdx int) error {
	c := p.Get()
	if c == nil {
		return errors.New("get redis conn error")
	}
	defer c.Close()

	c.Send(Multi)
	c.Send(Select, dbIdx)
	if expire > 0 {
		_ = c.Send(SetEx, *key, expire, *value)
	} else {
		_ = c.Send(Set, *key, *value)
	}
	_, err := c.Do(Exec)
	return err
}

//DeleteValueForDBIdx 删除一个key dbIdx 为所使用的DB的索引(默认0-15)
func (p *RedisPool) DeleteValueForDBIdx(key *string, dbIdx int) (int, error) {
	c := p.Get()
	if c == nil {
		return 0, errors.New("get redis conn error")
	}
	defer c.Close() //函数运行结束 ，把连接放回连接池

	c.Send(Multi)
	c.Send(Select, dbIdx)
	c.Send(DEL, *key)

	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		return 0, err
	}
	return redis.Int(replay[1], err)
}

//GetValueForDBIdx 从Redis中获取指定的值
func (p *RedisPool) GetValueForDBIdx(key *string, dbIdx int) (*string, error) {
	c := p.Get()
	if c == nil {
		return nil, errors.New("can not get redis conn")
	}
	defer c.Close() //函数运行结束 ，把连接放回连接池

	c.Send("MULTI")
	c.Send("SELECT", dbIdx)
	c.Send(Get, *key)
	replay, err := redis.Values(c.Do("EXEC"))
	if err != nil {
		return nil, err
	}
	var value string
	value, err = redis.String(replay[1], err)
	if err != nil {
		if err ==  redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &value, err
}

// SetValue 用法：Set("key", val, 60)，其中 expire 的单位为秒
func (p *RedisPool) SetValue(key *string, value *string, expire int) error {
	if p.DBIndex == 0 {
		c := p.Get()
		if c == nil {
			return errors.New("get redis conn error")
		}
		defer c.Close()

		var err error
		if expire > 0 {
			_, err = c.Do(SetEx, *key, expire, *value)
		} else {
			_, err = c.Do(Set, *key, *value)
		}
		return err
	}
	return p.SetValueForDBIdx(key, value, expire, p.DBIndex)
}

//DeleteValue 删除值,默认为dbIdx库中的
func (p *RedisPool) DeleteValue(key *string) (int, error) {
	if p.DBIndex == 0 {
		c := p.Get()
		if c == nil {
			return 0, errors.New("get redis conn error")
		}
		defer c.Close()

		return redis.Int(c.Do(DEL, *key))
	}
	return p.DeleteValueForDBIdx(key, p.DBIndex)
}

//从Redis中获取指定的值
func (p *RedisPool) GetValue(key *string) (*string, error) {
	if p.DBIndex == 0 {
		c := p.Get()
		if c == nil {
			return nil, errors.New("get redis conn error")
		}
		defer c.Close() //函数运行结束 ，把连接放回连接池

		replay, err := redis.String(c.Do(Get, *key))
		//说明没有值
		if err == redis.ErrNil {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		return &replay, nil
	}
	return p.GetValueForDBIdx(key, p.DBIndex)
}

//MGetValue 一次性获取多个Key的值,不支持选择库!!
func (p *RedisPool) MGetValue(keys []string) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, errors.New("get redis conn error")
	}
	defer c.Close() //函数运行结束 ，把连接放回连接池

	s := make([]interface{}, len(keys))
	for i, v := range keys {
		s[i] = v
	}

	replay, err := redis.Strings(c.Do(MGet, s...))
	if err != nil {
		return nil, err
	}
	return replay, nil
}

//SetExpire 设置过期时间 不支持选择库!!
func (p *RedisPool) SetExpire(key *string, expire int) error {
	c := p.Get()
	if c == nil {
		return errors.New("get redis conn error")
	}
	defer c.Close() //函数运行结束 ，把连接放回连接池

	_, err := c.Do(Expire, *key, expire)
	return err
}

//方便连接池在系统退出的时候也能够优雅的退出
func (p *RedisPool) CloseRedisPool() {
	p.Close()
}
