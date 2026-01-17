package cache

import (
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/yeahyf/go_base/log"
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
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令
	RedisSend(c, Multi)
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
		return "", ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	var value string
	var err error
	if direct == LPOP {
		value, err = redis.String(c.Do(LPOP, key))
	} else {
		value, err = redis.String(c.Do(RPOP, key))
	}

	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// LMPop 从队列尾部获取数据,一次获取多个,尾部的序号为从0开始
// 注意: 一次获取多个 LPOP需要再高版本中实现，现在使用 LRANGE + LTRIM组合实现
func (p *RedisPool) LMPop(key string, start, stop uint32) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, ErrGetConn
	}
	defer CloseAction(c) // 函数运行结束 ，把连接放回连接池

	RedisSend(c, Multi)
	RedisSend(c, LRANGE, key, start, stop)
	RedisSend(c, LTRIM, key, stop+1, -1)
	replay, err := redis.Values(c.Do(Exec))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// replay[0] 是 LRANGE 的结果，replay[1] 是 LTRIM 的结果
	if len(replay) >= 1 {
		if source, ok := replay[0].([]any); ok {
			result := make([]string, 0, len(source))
			for _, v := range source {
				if s, ok := v.([]uint8); ok {
					result = append(result, string(s))
				} else if s, ok := v.(string); ok {
					result = append(result, s)
				}
			}
			return result, nil
		}
	}
	return nil, ErrGetValue
}

// ZAdd 向排序集合中增加数据
func (p *RedisPool) ZAdd(key, field string, value float64) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	//RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	_, err := c.Do(ZADD, key, value, field) // 直接覆盖

	//_, err := c.Do(Exec)
	return err
}

// ZMAdd 向排序集合中批量增加数据，field为字段的列表，value为对应score的列表
func (p *RedisPool) ZMAdd(key string, field []string, value []float64) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	length := len(field)
	for i := range length {
		RedisSend(c, ZADD, key, value[i], field[i]) // 直接覆盖
	}
	_, err := c.Do(Exec)
	return err
}

// zFetch 实现ZRange 与 ZRevRange的通用方法
func (p *RedisPool) zFetch(key string, start, stop uint32, isRev bool) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, ErrGetConn
	}
	defer CloseAction(c)

	var err error
	var value interface{}
	if isRev {
		value, err = c.Do(ZREVRANGE, key, start, stop)
	} else {
		value, err = c.Do(ZRANGE, key, start, stop)
	}
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// c.Do 直接返回 []any，转换为 []string
	if values, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(values))
		for _, v := range values {
			if s, ok := v.([]uint8); ok {
				result = append(result, string(s))
			} else if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result, nil
	}
	return nil, ErrGetValue
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
		return nil, nil, ErrGetConn
	}
	defer CloseAction(c)

	var err error
	var value interface{}
	if isRev {
		value, err = c.Do(ZREVRANGE, key, start, stop, WITHSCORES)
	} else {
		value, err = c.Do(ZRANGE, key, start, stop, WITHSCORES)
	}
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	// c.Do 直接返回 []any，转换为 []string 和 []float64
	if values, ok := value.([]any); ok {
		length := len(values)
		f := make([]string, 0, length/2)
		s := make([]float64, 0, length/2)
		for i := 0; i < length; i += 2 {
			var field string
			var score float64
			if v, ok := values[i].([]uint8); ok {
				field = string(v)
			} else if v, ok := values[i].(string); ok {
				field = v
			}
			if i+1 < length {
				if v, ok := values[i+1].([]uint8); ok {
					temp, err := strconv.ParseFloat(string(v), 64)
					if err == nil {
						score = temp
					}
				} else if v, ok := values[i+1].(string); ok {
					temp, err := strconv.ParseFloat(v, 64)
					if err == nil {
						score = temp
					}
				}
			}
			if field != "" {
				f = append(f, field)
				s = append(s, score)
			}
		}
		return f, s, nil
	}
	return nil, nil, ErrGetValue
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
		return nil, ErrGetConn
	}
	defer CloseAction(c)

	value, err := c.Do(ZRANGEBYSCORE, key, min, max)
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// c.Do 直接返回 []any，转换为 []string
	if values, ok := value.([]interface{}); ok {
		f := make([]string, 0, len(values))
		for _, val := range values {
			if v, ok := val.([]uint8); ok {
				f = append(f, string(v))
			} else if v, ok := val.(string); ok {
				f = append(f, v)
			}
		}
		return f, nil
	}
	return nil, ErrGetValue
}

// ZRangeByScoreWithScore 根据最大最小值获取列表（带分数）
func (p *RedisPool) ZRangeByScoreWithScore(key string, min, max float64) ([]string, []float64, error) {
	c := p.Get()
	if c == nil {
		return nil, nil, ErrGetConn
	}
	defer CloseAction(c)

	value, err := c.Do(ZRANGEBYSCORE, key, min, max, WITHSCORES)
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	// c.Do 直接返回 []any，转换为 []string 和 []float64
	if values, ok := value.([]interface{}); ok {
		length := len(values)
		f := make([]string, 0, length/2)
		s := make([]float64, 0, length/2)
		for i := 0; i < length; i += 2 {
			var field string
			var score float64
			if v, ok := values[i].([]uint8); ok {
				field = string(v)
			} else if v, ok := values[i].(string); ok {
				field = v
			}
			if i+1 < length {
				if v, ok := values[i+1].([]uint8); ok {
					temp, err := strconv.ParseFloat(string(v), 64)
					if err == nil {
						score = temp
					}
				} else if v, ok := values[i+1].(string); ok {
					temp, err := strconv.ParseFloat(v, 64)
					if err == nil {
						score = temp
					}
				}
			}
			if field != "" {
				f = append(f, field)
				s = append(s, score)
			}
		}
		return f, s, nil
	}
	return nil, nil, ErrGetValue
}

// ZRem 根据key进行删除,返回删除的数量
func (p *RedisPool) ZRem(key string, fields ...string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, ErrGetConn
	}
	defer CloseAction(c)

	args := make([]any, 0, len(fields)+1)
	args = append(args, key)
	for _, field := range fields {
		args = append(args, field)
	}

	return redis.Int64(c.Do(ZREM, args...))
}

// ZCard 获取有序集合成员个数
func (p *RedisPool) ZCard(key string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, ErrGetConn
	}
	defer CloseAction(c)

	result, err := redis.Int64(c.Do(ZCARD, key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}

// ZRemRangeByRank  移除指定索引的item
func (p *RedisPool) ZRemRangeByRank(key string, start, stop uint32) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	//RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	//RedisSend(c, ZREMRANGEBYRANK, key, start, stop)

	_, err := c.Do(ZREMRANGEBYRANK, key, start, stop)
	return err
}

// HSet 存储Hash数据
func (p *RedisPool) HSet(key, filed string, value string) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	//RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	//RedisSend(c, HSET, key, filed, value)

	_, err := c.Do(HSET, key, filed, value)
	return err
}

func (p *RedisPool) HLen(key string) (int64, error) {
	c := p.Get()
	if c == nil {
		return 0, ErrGetConn
	}
	defer CloseAction(c)

	result, err := redis.Int64(c.Do(HLEN, key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, err
	}
	return result, nil
}

// HMSet 批量存储Hash数据
// values 按照 key1 value1 key2 value2 key3 value3 排列
func (p *RedisPool) HMSet(key string, values ...string) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令
	//RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	args := make([]any, 0, len(values)+1)
	args = append(args, key)
	for _, v := range values {
		args = append(args, v)
	}
	//RedisSend(c, HSET, args...)
	_, err := c.Do(HSET, args...)
	return err
}

// HMSetWithMap 批量存储Hash数据
func (p *RedisPool) HMSetWithMap(key string, m map[string]string) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)
	for k, v := range m {
		RedisSend(c, HMSET, key, k, v)
	}
	_, err := c.Do(Exec)
	return err
}

func (p *RedisPool) HMGet(key string, fields []string) ([]string, error) {
	c := p.Get()
	if c == nil {
		return nil, ErrGetConn
	}
	defer CloseAction(c)

	args := make([]any, 0, len(fields)+1)
	args = append(args, key)
	for _, v := range fields {
		args = append(args, v)
	}

	value, err := c.Do(HMGET, args...)
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	// c.Do 直接返回 []any，转换为 []string
	if values, ok := value.([]interface{}); ok {
		result := make([]string, 0, len(values))
		for _, v := range values {
			if s, ok := v.([]uint8); ok {
				result = append(result, string(s))
			} else if s, ok := v.(string); ok {
				result = append(result, s)
			} else {
				result = append(result, "")
			}
		}
		return result, nil
	}
	return nil, ErrGetValue
}

// HDel 删除Hash数据
func (p *RedisPool) HDel(key string, dataKeys ...string) error {
	c := p.Get()
	if c == nil {
		return ErrGetConn
	}
	defer CloseAction(c)
	// 使用Do命令执行缓存区的命令

	//RedisSend(c, Multi)
	//RedisSend(c, Select, p.DBIndex)

	args := make([]any, 0, len(dataKeys)+1)
	args = append(args, key)
	for _, v := range dataKeys {
		args = append(args, v)
	}
	//RedisSend(c, HDEL, args...)
	_, err := c.Do(HDEL, args...)
	return err
}

// ExecScript 执行 lua 脚本，返回结果为 interface{}
func (p *RedisPool) ExecScript(script string, param ...any) (any, error) {
	c := p.Get()
	if c == nil {
		return "", ErrGetConn
	}
	defer CloseAction(c)
	// 不能在 lua 脚本中执行 select 操作，只能单独处理
	//RedisSend(c, Select, p.DBIndex)
	// 执行 Lua 脚本
	scriptSHA := redis.NewScript(1, script) // 1 表示 KEYS 的数量
	result, err := scriptSHA.Do(c, param...)
	if err != nil {
		log.Error("execute lua script error", err)
		return "", err
	}
	return result, nil
}

// ExecScriptString 执行 lua 脚本，返回结果为 string
func (p *RedisPool) ExecScriptString(script string, param ...any) (string, error) {
	return redis.String(p.ExecScript(script, param...))
}
