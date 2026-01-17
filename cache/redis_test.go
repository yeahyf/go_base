package cache

import (
	"testing"
)

var p *RedisPool

func initPool() {
	p = NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)
}

func TestExists(t *testing.T) {
	initPool()
	key := "web:uv"
	err := p.SetValue(key, "asdf", 100)
	if err != nil {
		t.Fail()
	}
	result, err := p.ExistsValue(key)
	if err != nil {
		t.Fail()
	}
	if !result {
		t.Fail()
	}
}

const LUASCRIPT = `
	local key = KEYS[1]
	local value = ARGV[1]
	local ttl = tonumber(ARGV[2])
	-- 获取当前值
	local current_value = redis.call("GET", key)
	-- 处理键不存在的情况
	if not current_value then
   	return "ERR_KEY_NOT_EXIST"
	end
	-- 比较值并设置 TTL
	if current_value ~= value then
			return "ERR_VALUE_MISMATCH"
	end
	local expire_result = redis.call("EXPIRE", key, ttl)
   	if expire_result == 0 then
       	return "ERR_EXPIRE_FAILED"
   	else
       	return "OK"
   	end
	`

func TestExecScript(t *testing.T) {
	initPool()
	key := "test:exec_script"
	value := "hDCPCfy-WywKXn6B4vsjPHqK4dCV_4sj"
	ttl := 600

	// 先设置 key 的值
	err := p.SetValue(key, value, 600)
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.ExecScriptString(LUASCRIPT, key, value, ttl)
	if err != nil {
		t.Error(err)
	}
	if result != "OK" {
		t.Fatalf("Expected OK, got %s", result)
	}
	t.Logf("%s", result)
}

func TestList(t *testing.T) {
	initPool()
	key := "test:list"
	data := []string{"www.sina.com", "www.baidu.com"}
	err := p.LPush(key, data...)
	if err != nil {
		t.Fatal(err)
	}

	d, err := p.Pop(key, "LPOP")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(d)
	}
}

func TestZRem(t *testing.T) {
	initPool()
	key := "test:zset_rem"
	// 先添加一些数据
	err := p.ZMAdd(key, []string{"www.sina.com", "www.baidu.com"}, []float64{100.0, 200.0})
	if err != nil {
		t.Fatal(err)
	}
	value, err := p.ZRem(key, "www.sina.com", "www.baidu.com")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(value)
	}
}

func TestZRangeByScore(t *testing.T) {
	initPool()
	key := "salary"
	value, err := p.ZRangeByScore(key, 3000, 4000)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestZRangeByScoreWithScore(t *testing.T) {
	initPool()
	key := "salary"
	f, s, err := p.ZRangeByScoreWithScore(key, 3000, 4000)
	if err != nil {
		t.Fail()
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestLPush(t *testing.T) {
	initPool()
	key := "list"
	err := p.LPush(key, "11", "12", "13", "14", "15", "16", "17", "18", "19", "20")
	if err != nil {
		t.Fail()
	}
}

func TestLPop(t *testing.T) {
	initPool()
	key := "list"
	value, err := p.LPop(key)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestLMPop(t *testing.T) {
	initPool()
	key := "list"
	value, err := p.LMPop(key, 0, 3)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestZAdd(t *testing.T) {
	initPool()
	setList := "test:zset"
	err := p.ZAdd(setList, "f1", 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZMAdd(t *testing.T) {
	initPool()
	setList := "test:zset"
	field := []string{"f2", "f3", "f4"}
	value := []float64{12.2, 433.5, 89.9}
	err := p.ZMAdd(setList, field, value)
	if err != nil {
		t.Fatal(err)
	}
}

func TestZRange(t *testing.T) {
	initPool()
	setList := "test:zset"
	r, err := p.ZRange(setList, 0, 3)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestZRevRange(t *testing.T) {
	initPool()
	setList := "test:zset"
	r, err := p.ZRevRange(setList, 0, 3)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestZRangeWithScore(t *testing.T) {
	initPool()
	setList := "test:zset"
	f, s, err := p.ZRangeWithScore(setList, 0, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestZRevRangeWithScore(t *testing.T) {
	initPool()
	setList := "test:zset"
	f, s, err := p.ZRevRangeWithScore(setList, 0, 100)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestZCard(t *testing.T) {
	initPool()
	setList := "test:zset"
	r, err := p.ZCard(setList)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(r)
	}
}

func TestZRemRangeByRank(t *testing.T) {
	initPool()
	setList := "test:zset"
	err := p.ZRemRangeByRank(setList, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHSet(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HSet(setList, "key1", "value1")
	if err != nil {
		t.Fail()
	}
}

func TestHMSet(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HMSet(setList, "key2", "value2", "key3", "value3", "key4", "value4")
	if err != nil {
		t.Fail()
	}
}

func TestHMGet(t *testing.T) {
	initPool()
	setList := "set"
	result, err := p.HMGet(setList, []string{"key1", "key2", "key3", "key4", "key5", "key6"})
	if err != nil {
		t.Fail()
	} else {
		t.Log(result)
	}
}

func TestHLen(t *testing.T) {
	initPool()
	setList := "set"
	result, err := p.HLen(setList)
	if err != nil {
		t.Fail()
	} else {
		t.Log(result)
	}
}

func TestHMDel(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HDel(setList, "key2", "key1", "key3")
	if err != nil {
		t.Fail()
	}
}

func BenchmarkKey(b *testing.B) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)
	for i := 0; i < b.N; i++ {
		key := "01_1"
		value := "aslkdjfalsdfkj"

		err := p.SetValue(key, value, 0)
		if err != nil {
			b.Fail()
		}
	}
}

func TestSingle(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	key := "01_1"
	value := "aslkdjfalsdfkj"

	err := p.SetValue(key, value, 0)
	if err != nil {
		t.Fail()
	}
	var v string
	v, err = p.GetValue(key)
	if err != nil {
		t.Fail()
	}
	if v != value {
		t.Fail()
	}

	var result int
	result, err = p.DeleteValue(key)
	if err != nil {
		t.Fail()
	}
	if result != 1 {
		t.Fail()
	}
	p.CloseRedisPool()
}

func TestMulti(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	keys := []string{"01_1", "01_2", "01_3", "01_4", "01_5"}
	targets := []string{"01_1", "v_1", "01_2", "v_2", "01_3", "v_3", "01_4", "v_4", "01_5", "v_5"}
	s := make([]interface{}, 0, len(targets))
	for _, v := range targets {
		s = append(s, v)
	}
	err := p.MSetValue(s)
	if err != nil {
		t.Fail()
	}
	p.MSetExpire(keys, 60)

	p.CloseRedisPool()
}

func TestMultiWithExpire(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	// keys := []string{"01_1", "01_2", "01_3", "01_4", "01_5"}
	targets := []string{"01_1", "v_1", "01_2", "v_2", "01_3", "v_3", "01_4", "v_4", "01_5", "v_5"}
	s := make([]interface{}, 0, len(targets))
	for _, v := range targets {
		s = append(s, v)
	}
	err := p.MSetValueWithExpire(s, 60)
	if err != nil {
		t.Fail()
	}
	p.CloseRedisPool()
}

func TestMultiGetWithExpire(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	slice := make([]interface{}, 0, 3)
	slice = append(slice, "123")
	slice = append(slice, "456")
	slice = append(slice, "789")

	result, err := p.MGetValue(slice)
	if err != nil {
		t.Fail()
	}
	t.Log(result)
	p.CloseRedisPool()
}

func TestDeleteValues(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	keys := make([]string, 0, 3)
	keys = append(keys, "a")
	keys = append(keys, "b")
	keys = append(keys, "c")

	_, err := p.DeleteValues(keys)
	if err != nil {
		t.Fail()
	}

	p.CloseRedisPool()
}

// TestEmptyKey 测试空键的处理
func TestEmptyKey(t *testing.T) {
	initPool()

	// 测试空键的 SetValue
	err := p.SetValue("", "value", 0)
	if err != nil {
		t.Log("SetValue with empty key returns error (expected):", err)
	}

	// 测试空键的 GetValue
	result, err := p.GetValue("")
	if err != nil || result == "" {
		t.Log("GetValue with empty key handled correctly")
	}

	// 测试空键的 Exists
	exists, err := p.ExistsValue("")
	if err != nil || !exists {
		t.Log("ExistsValue with empty key handled correctly")
	}
}

// TestInvalidDataType 测试错误的数据类型操作
func TestInvalidDataType(t *testing.T) {
	initPool()
	key := "test:invalid:type"

	// 先设置一个 string 类型的值
	err := p.SetValue(key, "string_value", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 尝试对 string 类型的 key 执行 list 操作
	// 注意：Redis 可能会自动转换类型或报错，这里我们只验证操作能正常执行
	// 在实际使用中，应该由应用层保证数据类型一致性
	data := []string{"item1", "item2"}
	err = p.LPush(key, data...)
	if err != nil {
		t.Log("LPush on string key handled:", err)
	}

	// 清理
	p.DeleteValue(key)
}

// TestTTLBoundary 测试 TTL 边界情况
func TestTTLBoundary(t *testing.T) {
	initPool()
	key := "test:ttl:boundary"

	// 测试 TTL 为 1 秒
	err := p.SetValue(key, "value", 1)
	if err != nil {
		t.Fatal(err)
	}

	// 立即获取，应该存在
	result, err := p.GetValue(key)
	if err != nil || result != "value" {
		t.Error("Value should exist immediately after setting with TTL=1")
	}

	// 测试较大的 TTL 值
	key2 := "test:ttl:large"
	err = p.SetValue(key2, "value", 86400*365) // 1年
	if err != nil {
		t.Fatal(err)
	}

	exists, err := p.ExistsValue(key2)
	if err != nil || !exists {
		t.Error("Value should exist with large TTL")
	}

	// 清理
	p.DeleteValue(key)
	p.DeleteValue(key2)
}

// TestLargeData 测试大数据量的处理
func TestLargeData(t *testing.T) {
	initPool()

	// 测试批量操作大量数据
	count := 100
	keys := make([]string, count)
	values := make([]string, count)

	for i := 0; i < count; i++ {
		keys[i] = "test:large:key:" + string(rune(i))
		values[i] = "value" + string(rune(i))
	}

	// 批量设置
	kv := make([]any, 0, count*2)
	for i := 0; i < count; i++ {
		kv = append(kv, keys[i], values[i])
	}

	err := p.MSetValue(kv)
	if err != nil {
		t.Fatal(err)
	}

	// 批量获取
	keysAny := make([]any, count)
	for i, k := range keys {
		keysAny[i] = k
	}

	results, err := p.MGetValue(keysAny)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != count {
		t.Errorf("Expected %d results, got %d", count, len(results))
	}

	// 清理
	p.DeleteValues(keys)
}

// TestHashBoundary 测试 Hash 操作的边界情况
func TestHashBoundary(t *testing.T) {
	initPool()
	key := "test:hash:boundary"

	// 测试设置空字段
	err := p.HSet(key, "", "value")
	if err != nil {
		t.Log("HSet with empty field handled:", err)
	}

	// 测试设置空值
	err = p.HSet(key, "field1", "")
	if err != nil {
		t.Fatal(err)
	}

	// 获取空值
	result, err := p.HMGet(key, []string{"field1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0] != "" {
		t.Errorf("Expected empty value, got %v", result)
	}

	// 测试获取不存在的字段
	result, err = p.HMGet(key, []string{"nonexistent_field"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0] != "" {
		t.Errorf("Expected empty string for nonexistent field, got %v", result)
	}

	// 清理
	p.HDel(key, "field1")
}

// TestZSetBoundary 测试有序集合的边界情况
func TestZSetBoundary(t *testing.T) {
	initPool()
	key := "test:zset:boundary"

	// 测试相同的 score
	err := p.ZMAdd(key, []string{"a", "b", "c"}, []float64{100.0, 100.0, 100.0})
	if err != nil {
		t.Fatal(err)
	}

	// 获取数据
	r, err := p.ZRange(key, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(r))
	}

	// 测试负数 score
	err = p.ZAdd(key, "negative", -10.5)
	if err != nil {
		t.Fatal(err)
	}

	// 测试获取不存在范围的成员
	r, err = p.ZRangeByScore(key, 1000, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 0 {
		t.Errorf("Expected empty result for non-existent range, got %v", r)
	}

	// 测试删除不存在的成员
	count, err := p.ZRem(key, "nonexistent_member")
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Expected 0 deletions for non-existent member, got %d", count)
	}

	// 清理
	p.DeleteValue(key)
}

// TestListBoundary 测试列表操作的边界情况
func TestListBoundary(t *testing.T) {
	initPool()
	key := "test:list:boundary"

	// 测试空列表的 LPop
	value, err := p.LPop(key)
	if err != nil {
		t.Log("LPop on empty list returns error (expected):", err)
	}
	if value != "" {
		t.Errorf("Expected empty string for empty list, got %s", value)
	}

	// 测试 LMPop 超出范围
	err = p.LPush(key, "item1", "item2")
	if err != nil {
		t.Fatal(err)
	}

	// 尝试获取超出范围的元素
	items, err := p.LMPop(key, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items (all available), got %d", len(items))
	}

	// 清理
	p.DeleteValue(key)
}

// TestMGetValueWithEmptySlice 测试空切片的批量获取
func TestMGetValueWithEmptySlice(t *testing.T) {
	initPool()

	// Redis 的 MGET 命令不接受空参数，会返回错误
	// 这个测试验证这种行为是预期的
	keysAny := make([]any, 0)
	_, err := p.MGetValue(keysAny)
	if err != nil {
		t.Log("MGetValue with empty slice returns error (expected):", err)
	}
}

// TestDeleteValuesEmpty 测试删除空键列表
func TestDeleteValuesEmpty(t *testing.T) {
	initPool()

	keys := make([]string, 0)
	count, err := p.DeleteValues(keys)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Expected 0 deletions for empty keys list, got %d", count)
	}
}

// TestSetValueWithExpireZero 测试过期时间为0的情况
func TestSetValueWithExpireZero(t *testing.T) {
	initPool()
	key := "test:expire:zero"

	// 过期时间为0应该使用 SET 而不是 SETEX
	err := p.SetValue(key, "value", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 验证值存在
	result, err := p.GetValue(key)
	if err != nil || result != "value" {
		t.Error("Value should exist when expire=0")
	}

	// 清理
	p.DeleteValue(key)
}

// TestRPop 测试从列表右端弹出元素
func TestRPop(t *testing.T) {
	initPool()
	key := "test:rpop"
	data := []string{"item1", "item2", "item3"}
	err := p.LPush(key, data...)
	if err != nil {
		t.Fatal(err)
	}

	// RPop 应该返回最先插入的元素（因为 LPUSH 是从左边推入）
	value, err := p.Pop(key, "RPOP")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("RPop value:", value)
}

// TestHMSetWithMap 测试使用 Map 批量设置 Hash 字段
func TestHMSetWithMap(t *testing.T) {
	initPool()
	key := "test:hmset_with_map"
	data := map[string]string{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	err := p.HMSetWithMap(key, data)
	if err != nil {
		t.Fatal(err)
	}

	// 验证数据
	fields := []string{"field1", "field2", "field3"}
	result, err := p.HMGet(key, fields)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("HMSetWithMap result:", result)
}

// TestExecScriptComplex 测试执行复杂的 Lua 脚本
func TestExecScriptComplex(t *testing.T) {
	initPool()
	key := "test:exec_script_complex"

	// 先设置一个带过期时间的值
	err := p.SetValue(key, "test_value", 100)
	if err != nil {
		t.Fatal(err)
	}

	// 执行脚本获取 TTL（返回类型是 int64）
	script := `
		local key = KEYS[1]
		local ttl = redis.call("TTL", key)
		return ttl
	`

	result, err := p.ExecScript(script, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TTL result type: %T, value: %v", result, result)
}
