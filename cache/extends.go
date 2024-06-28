package cache

import (
	"strconv"

	"github.com/gomodule/redigo/redis"
)

const (
	LPUSH           = "LPUSH"
	LPOP            = "LPOP"
	RPOP            = "RPOP"
	ZADD            = "ZADD"
	ZRANGE          = "ZRANGE"
	ZREVRANGE       = "ZREVRANGE"
	WITHSCORES      = "WITHSCORES"
	ZCARD           = "ZCARD"
	ZREMRANGEBYRANK = "ZREMRANGEBYRANK"
	ZRANGEBYSCORE   = "ZRANGEBYSCORE"
	ZREM            = "ZREM"
	HSET            = "HSET"
	HMSET           = "HMSET"
	HDEL            = "HDEL"
	HMGET           = "HMGET"
	LRANGE          = "LRANGE"
	LTRIM           = "LTRIM"
	HLEN            = "HLEN"
)

// LPush 向队列头部插入字符串数据，value为可变参数，一次可以插入多个
func (p *RedisPool) LPush(key string, values ...string) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令
	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	for _, v := range values {
		RedisSend(c, LPUSH, key, v)
	}

	_, err := c.Do(Exec)
	return err
}

// LPop 从队列尾部获取数据，一次获取一个
func (p *RedisPool) LPop(key string) (string, error) {
	return p.Pop(key, LPOP)
}

// Pop 从队列中获取一条数据
func (p *RedisPool) Pop(key, direct string) (string, error) {
	c := p.Get()
	if c == nil {
		return "", getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	if direct == LPOP {
		RedisSend(c, LPOP, key)
	} else {
		RedisSend(c, RPOP, key)
	}

	value, err := redis.Strings(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		} else {
			return "", err
		}
	}
	//长度必须为2
	if len(value) == 2 {
		return value[1], nil
	}
	return "", valueErr
}

// LMPop 从队列尾部获取数据,一次获取多个,尾部的序号为从0开始
// 注意: 一次获取多个 LPOP需要再高版本中实现，现在使用 LRANGE + LTRIM组合实现
func (p *RedisPool) LMPop(key string, start, stop uint32) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, getConnErr
	}
	defer CloseAction(c) //函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, LRANGE, key, start, stop)
	RedisSend(c, LTRIM, key, stop+1, -1)
	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if len(value) == 3 {
		if source, ok := value[1].([]interface{}); ok {
			result := make([]string, 0, len(source))
			for _, v := range source {
				if s, ok := v.([]uint8); ok {
					result = append(result, string(s))
				}
			}
			return result, nil
		}
	}
	return nil, valueErr
}

// ZAdd 向排序集合中增加数据
func (p *RedisPool) ZAdd(key, field string, value float64) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, ZADD, key, value, field) //直接覆盖

	_, err := c.Do(Exec)
	return err
}

// ZMAdd 向排序集合中批量增加数据，field为字段的列表，value为对应score的列表
func (p *RedisPool) ZMAdd(key string, field []string, value []float64) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	length := len(field)
	for i := 0; i < length; i++ {
		RedisSend(c, ZADD, key, value[i], field[i]) //直接覆盖
	}
	_, err := c.Do(Exec)
	return err
}

// zFetch 实现ZRange 与 ZRevRange的通用方法
func (p *RedisPool) zFetch(key string, start, stop uint32, isRev bool) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	if isRev {
		RedisSend(c, ZREVRANGE, key, start, stop)
	} else {
		RedisSend(c, ZRANGE, key, start, stop)
	}
	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if len(value) == 2 {
		if source, ok := value[1].([]interface{}); ok {
			result := make([]string, 0, len(source))
			for _, v := range source {
				if s, ok := v.([]uint8); ok {
					result = append(result, string(s))
				}
			}
			return result, nil
		}
	}
	return nil, valueErr
}

// ZRange 获取有序集合中的指定位置的数据（不带分数）
func (p *RedisPool) ZRange(key string, start, stop uint32) ([]string, error) {
	return p.zFetch(key, start, stop, false)
}

// ZRevRange  反向获取有序集合中的指定位置的数据（带分数）
func (p *RedisPool) ZRevRange(key string, start, stop uint32) ([]string, error) {
	return p.zFetch(key, start, stop, true)
}

// ZRangeWithScore 获取有序集合中的指定位置的数据（带分数）
func (p *RedisPool) zFetchWithScore(key string, start, stop uint32, isRev bool) ([]string, []float64, error) {
	c := p.Get()
	if c == nil {
		return nil, nil, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	if isRev {
		RedisSend(c, ZREVRANGE, key, start, stop, WITHSCORES)
	} else {
		RedisSend(c, ZRANGE, key, start, stop, WITHSCORES)
	}

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil, nil
		} else {
			return nil, nil, err
		}
	}
	if len(value) == 2 {
		if source, ok := value[1].([]interface{}); ok {
			length := len(source)
			f := make([]string, 0, length/2)
			s := make([]float64, 0, length/2)
			for i := 0; i < length; i += 2 {
				var field string
				var score float64
				if v, ok := source[i].([]uint8); ok {
					field = string(v)
				}
				if v, ok := source[i+1].([]uint8); ok {
					temp, err := strconv.ParseFloat(string(v), 64)
					if err == nil {
						score = temp
					}
				}
				if field != "" {
					f = append(f, field)
					s = append(s, score)
				}
			}
			return f, s, nil
		}
	}
	return nil, nil, valueErr
}

func (p *RedisPool) ZRangeWithScore(key string, start, stop uint32) ([]string, []float64, error) {
	return p.zFetchWithScore(key, start, stop, false)
}

func (p *RedisPool) ZRevRangeWithScore(key string, start, stop uint32) ([]string, []float64, error) {
	return p.zFetchWithScore(key, start, stop, true)
}

// ZRangeByScore 根据最大最小值获取列表
func (p *RedisPool) ZRangeByScore(key string, min, max float64) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, ZRANGEBYSCORE, key, min, max)

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if len(value) == 2 {
		if source, ok := value[1].([]interface{}); ok {
			f := make([]string, 0, len(source))
			for _, v := range source {
				if v, ok := v.([]uint8); ok {
					field := string(v)
					f = append(f, field)
				}
			}
			return f, nil
		}
	}
	return nil, valueErr
}

// ZRangeByScoreWithScore 根据最大最小值获取列表（带分数）
func (p *RedisPool) ZRangeByScoreWithScore(key string, min, max float64) ([]string, []float64, error) {
	c := p.Get()
	if c == nil {
		return nil, nil, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, ZRANGEBYSCORE, key, min, max, WITHSCORES)

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil, nil
		} else {
			return nil, nil, err
		}
	}
	if len(value) == 2 {
		if source, ok := value[1].([]interface{}); ok {
			length := len(source)
			f := make([]string, 0, length/2)
			s := make([]float64, 0, length/2)
			for i := 0; i < length; i += 2 {
				var field string
				var score float64
				if v, ok := source[i].([]uint8); ok {
					field = string(v)
				}
				if v, ok := source[i+1].([]uint8); ok {
					temp, err := strconv.ParseFloat(string(v), 64)
					if err == nil {
						score = temp
					}
				}
				if field != "" {
					f = append(f, field)
					s = append(s, score)
				}
			}
			return f, s, nil
		}
	}
	return nil, nil, valueErr
}

// ZRem 根据key进行删除,返回删除的数量
func (p *RedisPool) ZRem(key string, fields ...string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, key)
	for _, field := range fields {
		args = append(args, field)
	}
	RedisSend(c, ZREM, args...)

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		return 0, err
	}
	if len(value) == 2 {
		if source, ok := value[1].(int64); ok {
			return source, nil
		}
	}
	return 0, valueErr
}

// ZCard 获取有序集合成员个数
func (p *RedisPool) ZCard(key string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, ZCARD, key)

	result, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		} else {
			return 0, err
		}
	}

	if len(result) == 2 {
		if v, ok := result[1].(int64); ok {
			return v, nil
		}
	}
	return 0, err
}

// ZRemRangeByRank  移除指定索引的item
func (p *RedisPool) ZRemRangeByRank(key string, start, stop uint32) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, ZREMRANGEBYRANK, key, start, stop)

	_, err := c.Do(Exec)
	return err
}

// HSet 存储Hash数据
func (p *RedisPool) HSet(key, filed string, value string) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, HSET, key, filed, value)

	_, err := c.Do(Exec)
	return err
}

func (p *RedisPool) HLen(key string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令
	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	RedisSend(c, HLEN, key)

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		} else {
			return 0, err
		}
	}
	//长度必须为2
	if len(value) == 2 {
		if length, ok := value[1].(int64); ok {
			if err != nil {
				return 0, valueErr
			} else {
				return length, nil
			}
		}
	}
	return 0, valueErr
}

// HMSet 批量存储Hash数据
// values 按照 key1 value1 key2 value2 key3 value3 排列
func (p *RedisPool) HMSet(key string, values ...string) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令
	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	args := make([]interface{}, 0, len(values)+1)
	args = append(args, key)
	for _, v := range values {
		args = append(args, v)
	}
	RedisSend(c, HSET, args...)
	_, err := c.Do(Exec)
	return err
}

// HMSetWithMap 批量存储Hash数据
func (p *RedisPool) HMSetWithMap(key string, m map[string]string) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)
	for k, v := range m {
		RedisSend(c, HMSET, key, k, v)
	}
	_, err := c.Do(Exec)
	return err
}

func (p *RedisPool) HMGet(key string, fields []string) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)

	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, key)
	for _, v := range fields {
		args = append(args, v)
	}
	RedisSend(c, HMGET, args...)

	value, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	if len(value) == 2 {
		if source, ok := value[1].([]interface{}); ok {
			length := len(source)
			result := make([]string, 0, length)
			for _, v := range source {
				if s, ok := v.([]uint8); ok {
					result = append(result, string(s))
				} else {
					result = append(result, "")
				}
			}
			return result, nil

		}
	}
	return nil, valueErr
}

// HDel 删除Hash数据
func (p *RedisPool) HDel(key string, dataKeys ...string) error {
	c := p.Get()
	if c == nil {
		return getConnErr
	}
	defer CloseAction(c)
	//使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	RedisSend(c, Select, p.DBIndex)

	args := make([]interface{}, 0, len(dataKeys)+1)
	args = append(args, key)
	for _, v := range dataKeys {
		args = append(args, v)
	}
	RedisSend(c, HDEL, args...)
	_, err := c.Do(Exec)
	return err
}
