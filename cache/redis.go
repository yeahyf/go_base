/// 提供Reids的基本管理接口
package cache

import (
	"fmt"
	"gobase/immut"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	action_Set = "SET"
	action_Get = "GET"

	action_Expire = "EXPIRE"
	action_SetEx  = "SETEX"

	action_Zadd      = "ZADD"
	action_Zcard     = "ZCARD"
	action_Zcount    = "ZCOUNT"
	action_Zrevrange = "ZREVRANGE"
	action_Zrevrank  = "ZREVRANK"

	action_Zrank  = "ZRANK"
	action_Zrange = "ZRANGE"

	action_WithScores = "WITHSCORES"
)

type RedisPool struct {
	*redis.Pool //创建redis连接池
}

//构建新的Redis连接池
func NewRedisPool(init, maxsize, idle int, address, passwd string) *RedisPool {
	fmt.Println("Start init Redis ... ")
	redisPool := &redis.Pool{ //实例化一个连接池
		MaxIdle:     init,                              //最初的连接数量
		MaxActive:   maxsize,                           //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		Wait:        true,                              //没有连接可用需要等待
		IdleTimeout: time.Second * time.Duration(idle), //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			if passwd == immut.Blank_String {
				return redis.Dial("tcp", address)
			} else {
				return redis.Dial("tcp", address, redis.DialPassword(passwd))
			}
		},
	}
	return &RedisPool{
		redisPool,
	}
}

// Set 用法：Set("key", val, 60)，其中 expire 的单位为秒
func (p *RedisPool) SetValue(key *string, value *string, expire int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	var err error
	if expire > 0 {
		_, err = c.Do(action_SetEx, *key, expire, *value)
	} else {
		_, err = c.Do(action_Set, *key, *value)
	}
	return err
}

//从Redis中获取指定的值
func (p *RedisPool) GetValue(key *string) (*string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	replay, err := redis.String(c.Do(action_Get, *key))
	//说明没有值
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &replay, nil
}

//设置过期时间
func (p *RedisPool) SetExpire(key *string, expire int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	_, err := c.Do(action_Expire, *key, expire)
	return err
}

///向有序集合增加元素或修改元素
///注意：当不存在某个有序集合的时候直接使用zadd会创建这个有序集合
func (p *RedisPool) ZAdd(key, member *string, value float32) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	_, err := c.Do(action_Zadd, *key, value, *member)
	return err
}

///获取有序集合总成员数
func (p *RedisPool) Zcard(key *string) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Int(c.Do(action_Zcard, *key))
}

///计算指定区间分数成员
func (p *RedisPool) Zcount(key *string, min, max float32) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Int(c.Do(action_Zcount, *key, min, max))
}

///按照分数从高到低获取成员信息
func (p *RedisPool) Zrevrange(key *string, start, stop int) ([]string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Strings(c.Do(action_Zrevrange, *key, start, stop, action_WithScores))
}

///按照分数从低到高获取成员信息
func (p *RedisPool) Zrange(key *string, start, stop int) ([]string, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Strings(c.Do(action_Zrange, *key, start, stop, action_WithScores))
}

///按照分数从高低获取用户的排名信息
func (p *RedisPool) Zrevrank(key, member *string) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Int(c.Do(action_Zrevrank, *key, *member))
}

func (p *RedisPool) Zrank(key, member *string) (int, error) {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	return redis.Int(c.Do(action_Zrank, *key, *member))
}

//方便连接池在系统退出的时候也能够优雅的退出
func (p *RedisPool) CloseRedisPool() {
	p.Close()
}
