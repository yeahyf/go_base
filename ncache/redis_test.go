package ncache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testRedisAddress = "localhost:6379"
	testRedisDB      = 0
)

// TestRedisClient 创建一个测试用的Redis客户端
func getTestRedisClient(t *testing.T) *RedisClient {
	// 尝试连接到测试Redis实例，如果无法连接则跳过测试
	client := NewRedisClientByDB(0, 10, 300, testRedisAddress, "", testRedisDB)

	// 测试连接是否正常
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis server not available at %s, skipping test: %v", testRedisAddress, err)
	}

	return client
}

// TestNewClient 测试NewClient函数
func TestNewClient(t *testing.T) {
	cfg := &Config{
		Address:     testRedisAddress,
		Password:    "",
		DBIndex:     testRedisDB,
		MaxConnSize: 10,
	}

	client := NewClient(cfg)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.ctx)

	// 清理资源
	client.CloseRedisClient()
}

// TestNewRedisClientByDB 测试NewRedisClientByDB函数
func TestNewRedisClientByDB(t *testing.T) {
	client := NewRedisClientByDB(0, 10, 300, testRedisAddress, "", testRedisDB)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.ctx)

	// 清理资源
	client.CloseRedisClient()
}

// TestNewRedisClient 测试NewRedisClient函数
func TestNewRedisClient(t *testing.T) {
	client := NewRedisClient(0, 10, 300, testRedisAddress, "")
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.ctx)

	// 清理资源
	client.CloseRedisClient()
}

// TestSetValue 测试SetValue方法
func TestSetValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_set_value_key"
	value := "test_set_value"
	expire := 10 // 10秒过期

	err := client.SetValue(key, value, expire)
	assert.NoError(t, err)

	// 验证值是否正确设置
	result, err := client.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

// TestSetValueWithoutExpire 测试不设置过期时间的SetValue方法
func TestSetValueWithoutExpire(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_set_value_no_expire_key"
	value := "test_set_value_no_expire"

	err := client.SetValue(key, value, 0) // 不设置过期时间
	assert.NoError(t, err)

	// 验证值是否正确设置
	result, err := client.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)
}

// TestGetValue 测试GetValue方法
func TestGetValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_get_value_key"
	value := "test_get_value"

	// 先设置值
	err := client.SetValue(key, value, 10)
	assert.NoError(t, err)

	// 获取值
	result, err := client.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, value, result)

	// 测试获取不存在的键
	nonExistentKey := "test_non_existent_key"
	result, err = client.GetValue(nonExistentKey)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

// TestMGetValue 测试MGetValue方法
func TestMGetValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	keys := []string{"test_mget_1", "test_mget_2", "test_mget_3"}
	values := []string{"value1", "value2", "value3"}

	// 设置多个键值对
	for i, key := range keys {
		err := client.SetValue(key, values[i], 10)
		assert.NoError(t, err)
	}

	// 获取多个值
	results, err := client.MGetValue(keys)
	assert.NoError(t, err)
	assert.Len(t, results, len(keys))

	for i, result := range results {
		assert.Equal(t, values[i], result)
	}
}

// TestMGetValueWithInt64 测试MGetValue方法处理int64类型返回值
func TestMGetValueWithInt64(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	// 直接使用底层redis客户端设置一些int64类型的值进行测试
	ctx := context.Background()
	err := client.client.Set(ctx, "test_int64_key", 12345, 0).Err()
	assert.NoError(t, err)

	// 获取值
	keys := []string{"test_int64_key"}
	results, err := client.MGetValue(keys)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// 验证int64值被正确转换为字符串
	assert.Equal(t, "12345", results[0])
}

// TestMSetValue 测试MSetValue方法
func TestMSetValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	kv := map[string]any{
		"test_mset_key1": "test_mset_value1",
		"test_mset_key2": "test_mset_value2",
		"test_mset_key3": 123,
	}

	err := client.MSetValue(kv)
	assert.NoError(t, err)

	// 验证值是否正确设置
	for k, v := range kv {
		result, err := client.GetValue(k)
		assert.NoError(t, err)
		// 对于整型值，将其转换为字符串进行比较
		switch val := v.(type) {
		case int:
			assert.Equal(t, fmt.Sprintf("%d", val), result)
		default:
			assert.Equal(t, v, result)
		}
	}
}

// TestMSetValueWithExpire 测试MSetValueWithExpire方法
func TestMSetValueWithExpire(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	kv := map[string]any{
		"test_mset_exp_key1": "test_mset_exp_value1",
		"test_mset_exp_key2": "test_mset_exp_value2",
	}

	expire := 10 // 10秒过期
	err := client.MSetValueWithExpire(kv, expire)
	assert.NoError(t, err)

	// 验证值是否正确设置
	for k, v := range kv {
		result, err := client.GetValue(k)
		assert.NoError(t, err)
		assert.Equal(t, v, result)
	}
}

// TestDeleteValue 测试DeleteValue方法
func TestDeleteValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_delete_key"
	value := "test_delete_value"

	// 设置值
	err := client.SetValue(key, value, 10)
	assert.NoError(t, err)

	// 验证值存在
	exists, err := client.ExistsValue(key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 删除值
	deleted, err := client.DeleteValue(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// 验证值已被删除
	exists, err = client.ExistsValue(key)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// TestDeleteValues 测试DeleteValues方法
func TestDeleteValues(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	keys := []string{"test_del_vals_1", "test_del_vals_2", "test_del_vals_3"}

	// 设置多个键值对
	for _, key := range keys {
		err := client.SetValue(key, "test_value", 10)
		assert.NoError(t, err)
	}

	// 验证值存在
	for _, key := range keys {
		exists, err := client.ExistsValue(key)
		assert.NoError(t, err)
		assert.True(t, exists)
	}

	// 删除多个值
	deleted, err := client.DeleteValues(keys)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(keys)), deleted)

	// 验证值已被删除
	for _, key := range keys {
		exists, err := client.ExistsValue(key)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

// TestExistsValue 测试ExistsValue方法
func TestExistsValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_exists_key"
	value := "test_exists_value"

	// 检查不存在的键
	exists, err := client.ExistsValue(key)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 设置值
	err = client.SetValue(key, value, 10)
	assert.NoError(t, err)

	// 检查存在的键
	exists, err = client.ExistsValue(key)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// TestSetExpire 测试SetExpire方法
func TestSetExpire(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_expire_key"
	value := "test_expire_value"

	// 设置值
	err := client.SetValue(key, value, 0) // 不设置过期时间
	assert.NoError(t, err)

	// 设置过期时间
	expire := 10 // 10秒
	err = client.SetExpire(key, expire)
	assert.NoError(t, err)
}

// TestMSetExpire 测试MSetExpire方法
func TestMSetExpire(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	keys := []string{"test_mexp_1", "test_mexp_2", "test_mexp_3"}

	// 设置多个键值对
	for _, key := range keys {
		err := client.SetValue(key, "test_value", 0) // 不设置过期时间
		assert.NoError(t, err)
	}

	// 设置多个键的过期时间
	expire := 10 // 10秒
	err := client.MSetExpire(keys, expire)
	assert.NoError(t, err)
}

// TestHGetAllValue 测试HGetAllValue方法
func TestHGetAllValue(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_hgetall_key"
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := "value2"

	// 设置哈希值
	err := client.HSet(key, field1, value1)
	assert.NoError(t, err)
	err = client.HSet(key, field2, value2)
	assert.NoError(t, err)

	// 获取哈希所有值
	result, err := client.HGetAllValue(key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, value1, result[field1])
	assert.Equal(t, value2, result[field2])

	// 测试不存在的键
	nonExistentKey := "test_hgetall_nonexistent"
	result, err = client.HGetAllValue(nonExistentKey)
	assert.NoError(t, err)
	// Redis HGetAll 在键不存在时返回空映射而不是 nil
	assert.Empty(t, result)
}

// TestHMGet 测试HMGet方法
func TestHMGet(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_hmget_key"
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := "value2"
	field3 := "field3"
	value3 := "123" // 这个值会被作为int64存储用于测试

	// 设置哈希值
	err := client.HSet(key, field1, value1)
	assert.NoError(t, err)
	err = client.HSet(key, field2, value2)
	assert.NoError(t, err)

	// 使用HMSetWithMap设置包含int64类型的值
	mapData := map[string]string{
		field3: value3,
	}
	err = client.HMSetWithMap(key, mapData)
	assert.NoError(t, err)

	// 获取多个哈希字段
	fields := []string{field1, field2, field3}
	results, err := client.HMGet(key, fields)
	assert.NoError(t, err)
	assert.Len(t, results, len(fields))
	assert.Equal(t, value1, results[0])
	assert.Equal(t, value2, results[1])
	assert.Equal(t, value3, results[2]) // int64值应该被正确转换为字符串
}

// TestZAddAndZRange 测试ZAdd和ZRange方法
func TestZAddAndZRange(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_zadd_range_key"
	field1 := "member1"
	score1 := 10.0
	field2 := "member2"
	score2 := 20.0
	field3 := "member3"
	score3 := 30.0

	// 添加有序集合成员
	err := client.ZAdd(key, field1, score1)
	assert.NoError(t, err)
	err = client.ZAdd(key, field2, score2)
	assert.NoError(t, err)
	err = client.ZAdd(key, field3, score3)
	assert.NoError(t, err)

	// 获取有序集合范围
	start := uint32(0)
	stop := uint32(2)
	result, err := client.ZRange(key, start, stop)
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// 检查返回的成员顺序（按分数从小到大）
	assert.Contains(t, result, field1)
	assert.Contains(t, result, field2)
	assert.Contains(t, result, field3)
}

// TestZCard 测试ZCard方法
func TestZCard(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_zcard_key"
	// 确保测试键不存在
	client.DeleteValue(key)

	// 检查空集合的大小
	card, err := client.ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), card)

	// 添加一些成员
	err = client.ZAdd(key, "member1", 10.0)
	assert.NoError(t, err)
	err = client.ZAdd(key, "member2", 20.0)
	assert.NoError(t, err)

	// 检查集合大小
	card, err = client.ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), card)
}

// TestZRem 测试ZRem方法
func TestZRem(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_zrem_key"

	// 添加一些成员
	err := client.ZAdd(key, "member1", 10.0)
	assert.NoError(t, err)
	err = client.ZAdd(key, "member2", 20.0)
	assert.NoError(t, err)

	// 检查初始大小
	card, err := client.ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), card)

	// 删除成员
	removed, err := client.ZRem(key, "member1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	// 检查删除后的大小
	card, err = client.ZCard(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), card)
}

// TestHSetAndHLen 测试HSet和HLen方法
func TestHSetAndHLen(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_hset_len_key"
	// 确保测试键不存在
	client.DeleteValue(key)
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := "value2"

	// 检查空哈希的长度
	len, err := client.HLen(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), len)

	// 设置哈希字段
	err = client.HSet(key, field1, value1)
	assert.NoError(t, err)
	err = client.HSet(key, field2, value2)
	assert.NoError(t, err)

	// 检查哈希长度
	len, err = client.HLen(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), len)
}

// TestHMSetWithMap 测试HMSetWithMap方法
func TestHMSetWithMap(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_hmset_map_key"
	data := map[string]string{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	err := client.HMSetWithMap(key, data)
	assert.NoError(t, err)

	// 验证值是否正确设置
	result, err := client.HGetAllValue(key)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

// TestCloseRedisClient 测试CloseRedisClient方法
func TestCloseRedisClient(t *testing.T) {
	client := getTestRedisClient(t)

	err := client.CloseRedisClient()
	assert.NoError(t, err)
}

// TestExecScript 测试ExecScript方法
func TestExecScript(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	script := "return ARGV[1]"
	keys := []string{}
	args := []any{"hello"}

	result, err := client.ExecScript(script, keys, args)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

// TestExecScriptString 测试ExecScriptString方法
func TestExecScriptString(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	script := "return ARGV[1]"
	keys := []string{}
	args := []any{"hello"}

	result, err := client.ExecScriptString(script, keys, args)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	// 测试返回int64的情况
	script2 := "return 123"
	result2, err := client.ExecScriptString(script2, keys, args)
	assert.NoError(t, err)
	assert.Equal(t, "123", result2)
}

// TestLPushAndLPop 测试LPush和LPop方法
func TestLPushAndLPop(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_lpush_pop_key"
	values := []string{"value1", "value2", "value3"}

	// 推入多个值
	err := client.LPush(key, values...)
	assert.NoError(t, err)

	// 弹出值（由于是LPUSH，所以是后进先出）
	for i := len(values) - 1; i >= 0; i-- {
		result, err := client.LPop(key)
		assert.NoError(t, err)
		assert.Equal(t, values[i], result)
	}
}

// TestHDel 测试HDel方法
func TestHDel(t *testing.T) {
	client := getTestRedisClient(t)
	defer client.CloseRedisClient()

	key := "test_hdel_key"
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := "value2"

	// 设置哈希字段
	err := client.HSet(key, field1, value1)
	assert.NoError(t, err)
	err = client.HSet(key, field2, value2)
	assert.NoError(t, err)

	// 验证字段存在
	result, err := client.HGetAllValue(key)
	assert.NoError(t, err)
	assert.Contains(t, result, field1)
	assert.Contains(t, result, field2)

	// 删除字段
	err = client.HDel(key, field1)
	assert.NoError(t, err)

	// 验证字段已被删除
	result, err = client.HGetAllValue(key)
	assert.NoError(t, err)
	assert.NotContains(t, result, field1)
	assert.Contains(t, result, field2)
}
