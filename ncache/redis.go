package ncache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrGetConn  = errors.New("get redis conn error")
	ErrGetValue = errors.New("get redis data exception")
)

// RedisClient Redis客户端结构
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// Config 配置参数
type Config struct {
	InitConnSize int    // 初始化连接数量 (此参数在新客户端中不使用，但为了保持兼容性保留)
	MaxConnSize  int    // 最大连接数
	MaxIdleTime  int    // 连接最大空闲时间 (此参数在新客户端中不使用，但为了保持兼容性保留)
	Address      string // 服务器地址
	Username     string // 账号名称，如没有设置为空
	Password     string // 密码
	DBIndex      int    // 使用的 DB ID
}

// NewClient 根据 cfg 来设置
func NewClient(cfg *Config) *RedisClient {
	options := &redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DBIndex,
	}

	if cfg.Username != "" {
		options.Username = cfg.Username
	}

	if cfg.MaxConnSize > 0 {
		options.PoolSize = cfg.MaxConnSize
	}

	client := redis.NewClient(options)

	return &RedisClient{
		client: client,
		ctx:    context.Background(),
	}
}

// NewRedisClientByDB 构建新的Redis连接
func NewRedisClientByDB(init, maxsize, idle int, address, password string, dbIndex int) *RedisClient {
	cfg := &Config{
		InitConnSize: init,
		MaxConnSize:  maxsize,
		MaxIdleTime:  idle,
		Address:      address,
		Username:     "",
		Password:     password,
		DBIndex:      dbIndex,
	}
	return NewClient(cfg)
}

// NewRedisClient 构建新的Redis连接，放入默认的0号DB中
func NewRedisClient(init, maxsize, idle int, address, password string) *RedisClient {
	return NewRedisClientByDB(init, maxsize, idle, address, password, 0) // 默认选择0号库
}

// SetValue expire的单位为秒
func (r *RedisClient) SetValue(key string, value string, expire int) error {
	var err error
	if expire > 0 {
		err = r.client.Set(r.ctx, key, value, time.Duration(expire)*time.Second).Err()
	} else {
		err = r.client.Set(r.ctx, key, value, 0).Err()
	}

	return err
}

// DeleteValues 删除多个key
func (r *RedisClient) DeleteValues(keys []string) (int64, error) {
	result, err := r.client.Del(r.ctx, keys...).Result()
	return result, err
}

// DeleteValue 删除一个key
func (r *RedisClient) DeleteValue(key string) (int64, error) {
	result, err := r.client.Del(r.ctx, key).Result()
	return result, err
}

// GetValue 从Redis中获取指定的值
func (r *RedisClient) GetValue(key string) (string, error) {
	value, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// ExistsValue 判断某个 Key 是否有缓存
func (r *RedisClient) ExistsValue(key string) (bool, error) {
	count, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// HGetAllValue 从Redis中获取指定的值
func (r *RedisClient) HGetAllValue(key string) (map[string]string, error) {
	result, err := r.client.HGetAll(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// MGetValue 一次性获取多个Key的值
func (r *RedisClient) MGetValue(keys []string) ([]string, error) {
	result, err := r.client.MGet(r.ctx, keys...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	values := make([]string, len(result))
	for i, v := range result {
		if v != nil {
			switch val := v.(type) {
			case string:
				values[i] = val
			case int64:
				values[i] = fmt.Sprintf("%d", val)
			default:
				values[i] = ""
			}
		} else {
			values[i] = ""
		}
	}

	return values, nil
}

// MSetValue 批量设置
func (r *RedisClient) MSetValue(kv map[string]any) error {
	return r.client.MSet(r.ctx, kv).Err()
}

// MSetValueWithExpire 批量K-V以及对应的过期时间
// kv中的值需要按照 k1,v1,k2,v2,k3,v3 ... 进行存储
func (r *RedisClient) MSetValueWithExpire(kv map[string]any, expire int) error {
	// 使用事务来确保原子性
	pipe := r.client.TxPipeline()
	for k, v := range kv {
		pipe.Set(r.ctx, k, v, time.Duration(expire)*time.Second)
	}
	_, err := pipe.Exec(r.ctx)
	return err
}

// SetExpire 设置过期时间
func (r *RedisClient) SetExpire(key string, expire int) error {
	return r.client.Expire(r.ctx, key, time.Duration(expire)*time.Second).Err()
}

// MSetExpire 设置过期时间 keys 为key的列表 expire为对应的过期时间，单位为秒
func (r *RedisClient) MSetExpire(keys []string, expire int) error {
	pipe := r.client.Pipeline()
	for _, key := range keys {
		pipe.Expire(r.ctx, key, time.Duration(expire)*time.Second)
	}
	_, err := pipe.Exec(r.ctx)
	return err
}

// CloseRedisClient 关闭连接
func (r *RedisClient) CloseRedisClient() error {
	return r.client.Close()
}

// 以下是扩展功能函数
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
func (r *RedisClient) LPush(key string, values ...string) error {
	x := make([]any, len(values))
	for i, v := range values {
		x[i] = v
	}
	_, err := r.client.LPush(r.ctx, key, x...).Result()
	return err
}

// LPop 从队列尾部获取数据，一次获取一个
func (r *RedisClient) LPop(key string) (string, error) {
	return r.Pop(key, LPOP)
}

// Pop 从队列中获取一条数据
func (r *RedisClient) Pop(key, direct string) (string, error) {
	if direct == LPOP {
		return r.client.LPop(r.ctx, key).Result()
	} else {
		return r.client.RPop(r.ctx, key).Result()
	}
}

// LMPop 从队列尾部获取数据,一次获取多个
// 注意: 新版redis客户端提供了更直接的方法
func (r *RedisClient) LMPop(key string, start, stop uint32) ([]string, error) {
	values, err := r.client.LRange(r.ctx, key, int64(start), int64(stop)).Result()
	if err != nil {
		return nil, err
	}
	// 删除已经取出的元素
	if len(values) > 0 {
		err = r.client.LTrim(r.ctx, key, int64(stop+1), -1).Err()
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

// ZAdd 向排序集合中增加数据
func (r *RedisClient) ZAdd(key string, field string, value float64) error {
	_, err := r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  value,
		Member: field,
	}).Result()
	return err
}

// ZMAdd 向排序集合中批量增加数据，field为字段的列表，value为对应score的列表
func (r *RedisClient) ZMAdd(key string, field []string, value []float64) error {
	if len(field) != len(value) {
		return errors.New("field and value slices must have the same length")
	}

	members := make([]redis.Z, len(field))
	for i := 0; i < len(field); i++ {
		members[i] = redis.Z{
			Score:  value[i],
			Member: field[i],
		}
	}

	_, err := r.client.ZAdd(r.ctx, key, members...).Result()
	return err
}

// ZRange 获取有序集合中的指定位置的数据（不带分数）
func (r *RedisClient) ZRange(key string, start, stop uint32) ([]string, error) {
	result, err := r.client.ZRange(r.ctx, key, int64(start), int64(stop)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// ZRevRange  反向获取有序集合中的指定位置的数据（带分数）
func (r *RedisClient) ZRevRange(key string, start, stop uint32) ([]string, error) {
	result, err := r.client.ZRevRange(r.ctx, key, int64(start), int64(stop)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// ZRangeWithScore 获取有序集合中的指定位置的数据（带分数）
func (r *RedisClient) ZRangeWithScore(key string, start, stop uint32) ([]string, []float64, error) {
	result, err := r.client.ZRangeWithScores(r.ctx, key, int64(start), int64(stop)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	fields := make([]string, len(result))
	scores := make([]float64, len(result))
	for i, z := range result {
		fields[i] = z.Member.(string)
		scores[i] = z.Score
	}
	return fields, scores, nil
}

// ZRevRangeWithScore 反向获取有序集合中的指定位置的数据（带分数）
func (r *RedisClient) ZRevRangeWithScore(key string, start, stop uint32) ([]string, []float64, error) {
	result, err := r.client.ZRevRangeWithScores(r.ctx, key, int64(start), int64(stop)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	fields := make([]string, len(result))
	scores := make([]float64, len(result))
	for i, z := range result {
		fields[i] = z.Member.(string)
		scores[i] = z.Score
	}
	return fields, scores, nil
}

// ZRangeByScore 根据最大最小值获取列表
func (r *RedisClient) ZRangeByScore(key string, min, max float64) ([]string, error) {
	result, err := r.client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", min),
		Max: fmt.Sprintf("%f", max),
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

// ZRangeByScoreWithScore 根据最大最小值获取列表（带分数）
func (r *RedisClient) ZRangeByScoreWithScore(key string, min, max float64) ([]string, []float64, error) {
	result, err := r.client.ZRangeByScoreWithScores(r.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", min),
		Max: fmt.Sprintf("%f", max),
	}).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	fields := make([]string, len(result))
	scores := make([]float64, len(result))
	for i, z := range result {
		fields[i] = z.Member.(string)
		scores[i] = z.Score
	}
	return fields, scores, nil
}

// ZRem 根据key进行删除,返回删除的数量
func (r *RedisClient) ZRem(key string, fields ...string) (int64, error) {
	x := make([]any, len(fields))
	for i, v := range fields {
		x[i] = v
	}
	return r.client.ZRem(r.ctx, key, x...).Result()
}

// ZCard 获取有序集合成员个数
func (r *RedisClient) ZCard(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

// ZRemRangeByRank  移除指定索引的item
func (r *RedisClient) ZRemRangeByRank(key string, start, stop uint32) error {
	return r.client.ZRemRangeByRank(r.ctx, key, int64(start), int64(stop)).Err()
}

// HSet 存储Hash数据
func (r *RedisClient) HSet(key, field string, value string) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// HLen 获取Hash长度
func (r *RedisClient) HLen(key string) (int64, error) {
	return r.client.HLen(r.ctx, key).Result()
}

// HMSet 批量存储Hash数据
// values 按照 key1 value1 key2 value2 key3 value3 排列
func (r *RedisClient) HMSet(key string, values ...string) error {
	if len(values)%2 != 0 {
		return errors.New("values must be paired as key-value")
	}

	data := make(map[string]any)
	for i := 0; i < len(values); i += 2 {
		data[values[i]] = values[i+1]
	}
	return r.client.HMSet(r.ctx, key, data).Err()
}

// HMSetWithMap 批量存储Hash数据
func (r *RedisClient) HMSetWithMap(key string, m map[string]string) error {
	data := make(map[string]any)
	for k, v := range m {
		data[k] = v
	}
	return r.client.HMSet(r.ctx, key, data).Err()
}

// HMGet 批量获取Hash数据
func (r *RedisClient) HMGet(key string, fields []string) ([]string, error) {
	result, err := r.client.HMGet(r.ctx, key, fields...).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	values := make([]string, len(result))
	for i, v := range result {
		if v != nil {
			switch val := v.(type) {
			case string:
				values[i] = val
			case int64:
				values[i] = fmt.Sprintf("%d", val)
			default:
				values[i] = fmt.Sprintf("%v", val)
			}
		} else {
			values[i] = ""
		}
	}
	return values, nil
}

// HDel 删除Hash数据
func (r *RedisClient) HDel(key string, dataKeys ...string) error {
	_, err := r.client.HDel(r.ctx, key, dataKeys...).Result()
	return err
}

// ExecScript 执行 lua 脚本，返回结果为 any
func (r *RedisClient) ExecScript(script string, keys []string, args []any) (any, error) {
	s := redis.NewScript(script)
	return s.Run(r.ctx, r.client, keys, args...).Result()
}

// ExecScriptString 执行 lua 脚本，返回结果为 string
func (r *RedisClient) ExecScriptString(script string, keys []string, args []any) (string, error) {
	result, err := r.ExecScript(script, keys, args)
	if err != nil {
		return "", err
	}
	switch v := result.(type) {
	case string:
		return v, nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%f", v), nil
	case []byte:
		return string(v), nil
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}
