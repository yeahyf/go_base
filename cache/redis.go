package cache

import (
	"fmt"
	//"gobase/cfg"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	action_Set = "SET"
	action_Get = "GET"
	//action_Sadd      = "SADD"
	//action_SisMember = "SISMEMBER"
	action_Expire = "EXPIRE"
	action_SetEx  = "SETEX"
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
			if passwd == "" {
				return redis.Dial("tcp", address)
			} else {
				return redis.Dial("tcp", address, redis.DialPassword(passwd))
			}
		},
	}

	ANewRedisPool := &RedisPool{
		redisPool,
	}

	return ANewRedisPool
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

// //向集合中增加数据
// func Sadd(key, value *string, expire int) error {
// 	c := RedisPool.Get()
// 	defer c.Close() //函数运行结束 ，把连接放回连接池

// 	_, err := redis.Int(c.Do(action_Sadd, *key, *value))
// 	if err != nil {
// 		return err
// 	}
// 	_, err = c.Do(action_Expire, *key, expire)
// 	return err
// }

// //判断集合中是否有该数据
// func IsHasMember(key, value *string) (bool, error) {
// 	c := RedisPool.Get()
// 	defer c.Close() //函数运行结束 ，把连接放回连接池

// 	result, err := redis.Int(c.Do(action_SisMember, *key, *value))
// 	if err == nil {
// 		if result == 1 {
// 			return true, nil
// 		} else {
// 			return false, nil
// 		}
// 	}
// 	return false, err
// }

//设置过期时间
func (p *RedisPool) SetExpire(key *string, expire int) error {
	c := p.Get()
	defer c.Close() //函数运行结束 ，把连接放回连接池

	_, err := redis.Int(c.Do(action_Expire, *key, expire))
	return err
}

//方便连接池在系统退出的时候也能够优雅的退出
func (p *RedisPool) CloseRedisPool() {
	p.Close()
}
