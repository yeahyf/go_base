/// 提供Reids的基本管理接口
package cache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/yeahyf/go_base/immut"
)

const (
	actionSet = "SET"
	actionGet = "GET"
	actionDEL = "DEL"

	actionMGet = "MGET"

	actionExpire = "EXPIRE"
	actionSetEx  = "SETEX"
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

// Set 用法：Set("key", val, 60)，其中 expire 的单位为秒
func (p *RedisPool) SetValueForDBIdx(key *string, value *string, expire ,index int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	c.Do("SELECT",index)

	var err error
	if expire > 0 {
		_, err = c.Do(actionSetEx, *key, expire, *value)
	} else {
		_, err = c.Do(actionSet, *key, *value)
	}
	return err
}

func (p *RedisPool) DeleteValueForDBIdx(key *string,index int) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	c.Do("SELECT", index)

	return redis.Int(c.Do(actionDEL, *key))
}

//从Redis中获取指定的值
func (p *RedisPool) GetValueForDBIdx(key *string,index int) (*string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	c.Do("SELECT", index)

	replay, err := redis.String(c.Do(actionGet, *key))
	//说明没有值
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &replay, nil
}

// Set 用法：Set("key", val, 60)，其中 expire 的单位为秒
func (p *RedisPool) SetValue(key *string, value *string, expire int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	if p.DBIndex != 0 {
		c.Do("SELECT", p.DBIndex)
	}
	var err error
	if expire > 0 {
		_, err = c.Do(actionSetEx, *key, expire, *value)
	} else {
		_, err = c.Do(actionSet, *key, *value)
	}
	return err
}

func (p *RedisPool) DeleteValue(key *string) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	if p.DBIndex != 0 {
		c.Do("SELECT", p.DBIndex)
	}
	return redis.Int(c.Do(actionDEL, *key))
}

//从Redis中获取指定的值
func (p *RedisPool) GetValue(key *string) (*string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	if p.DBIndex != 0 {
		c.Do("SELECT", p.DBIndex)
	}
	replay, err := redis.String(c.Do(actionGet, *key))
	//说明没有值
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &replay, nil
}

func (p *RedisPool) MGetValue(keys []string) ([]string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	if p.DBIndex != 0 {
		c.Do("SELECT", p.DBIndex)
	}
	s := make([]interface{}, len(keys))
	for i, v := range keys {
		s[i] = v
	}

	replay, err := redis.Strings(c.Do(actionMGet, s...))
	if err != nil {
		return nil, err
	}
	return replay, nil
}

//设置过期时间
func (p *RedisPool) SetExpire(key *string, expire int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	if p.DBIndex != 0 {
		c.Do("SELECT", p.DBIndex)
	}
	_, err := c.Do(actionExpire, *key, expire)
	return err
}

//方便连接池在系统退出的时候也能够优雅的退出
func (p *RedisPool) CloseRedisPool() {
	p.Close()
}
