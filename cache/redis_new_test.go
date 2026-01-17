package cache

import (
	"testing"
)

var testPool *RedisPool

func initTestPool() {
	testPool = NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
}

func cleanupTestPool() {
	if testPool != nil {
		testPool.CloseRedisPool()
	}
}

// TestNewPool 测试使用 Config 创建连接池
func TestNewPool(t *testing.T) {
	cfg := &Config{
		InitConnSize: 1,
		MaxConnSize:  2,
		MaxIdleTime:  30,
		Address:      "127.0.0.1:6379",
		Username:     "",
		Password:     "",
		DBIndex:      0,
	}
	pool := NewPool(cfg)
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}
	pool.CloseRedisPool()
}

// TestNewRedisPoolByDB 测试使用参数创建连接池
func TestNewRedisPoolByDB(t *testing.T) {
	pool := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
	if pool == nil {
		t.Fatal("NewRedisPoolByDB returned nil")
	}
	pool.CloseRedisPool()
}

// TestNewRedisPool 测试创建默认数据库连接池
func TestNewRedisPool(t *testing.T) {
	pool := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")
	if pool == nil {
		t.Fatal("NewRedisPool returned nil")
	}
	pool.CloseRedisPool()
}

// TestSetValue 测试设置值（带过期时间）
func TestSetValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:set:with:expire"
	value := "test_value_with_expire"
	expire := 60

	err := testPool.SetValue(key, value, expire)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 验证值是否正确设置
	result, err := testPool.GetValue(key)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if result != value {
		t.Fatalf("Expected %s, got %s", value, result)
	}

	// 清理
	testPool.DeleteValue(key)
}

// TestSetValueWithoutExpire 测试设置值（不带过期时间）
func TestSetValueWithoutExpire(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:set:without:expire"
	value := "test_value_without_expire"

	err := testPool.SetValue(key, value, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 验证值是否正确设置
	result, err := testPool.GetValue(key)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if result != value {
		t.Fatalf("Expected %s, got %s", value, result)
	}

	// 清理
	testPool.DeleteValue(key)
}

// TestGetValue 测试获取值
func TestGetValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:get"
	value := "test_get_value"

	// 先设置值
	err := testPool.SetValue(key, value, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 获取值
	result, err := testPool.GetValue(key)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if result != value {
		t.Fatalf("Expected %s, got %s", value, result)
	}

	// 测试不存在的key
	nonExistent, err := testPool.GetValue("test:nonexistent")
	if err != nil {
		t.Fatalf("GetValue for non-existent key should not return error, got: %v", err)
	}
	if nonExistent != "" {
		t.Fatalf("Expected empty string for non-existent key, got: %s", nonExistent)
	}

	// 清理
	testPool.DeleteValue(key)
}

// TestDeleteValue 测试删除单个key
func TestDeleteValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:delete"
	value := "test_delete_value"

	// 先设置值
	err := testPool.SetValue(key, value, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 验证值存在
	exists, err := testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist before deletion")
	}

	// 删除值
	count, err := testPool.DeleteValue(key)
	if err != nil {
		t.Fatalf("DeleteValue failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 deleted key, got %d", count)
	}

	// 验证值已删除
	exists, err = testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if exists {
		t.Fatal("Key should not exist after deletion")
	}
}

// TestDeleteValuesNew 测试批量删除key
func TestDeleteValuesNew(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	keys := []string{"test:delete:1", "test:delete:2", "test:delete:3"}
	values := []string{"value1", "value2", "value3"}

	// 先设置值
	for i, key := range keys {
		err := testPool.SetValue(key, values[i], 0)
		if err != nil {
			t.Fatalf("SetValue failed for key %s: %v", key, err)
		}
	}

	// 验证值存在
	for _, key := range keys {
		exists, err := testPool.ExistsValue(key)
		if err != nil {
			t.Fatalf("ExistsValue failed: %v", err)
		}
		if !exists {
			t.Fatalf("Key %s should exist before deletion", key)
		}
	}

	// 批量删除
	count, err := testPool.DeleteValues(keys)
	if err != nil {
		t.Fatalf("DeleteValues failed: %v", err)
	}
	if count != len(keys) {
		t.Fatalf("Expected %d deleted keys, got %d", len(keys), count)
	}

	// 验证值已删除
	for _, key := range keys {
		exists, err := testPool.ExistsValue(key)
		if err != nil {
			t.Fatalf("ExistsValue failed: %v", err)
		}
		if exists {
			t.Fatalf("Key %s should not exist after deletion", key)
		}
	}
}

// TestExistsValue 测试判断key是否存在
func TestExistsValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:exists"
	value := "test_exists_value"

	// 测试不存在的key
	exists, err := testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if exists {
		t.Fatal("Key should not exist")
	}

	// 设置值
	err = testPool.SetValue(key, value, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 测试存在的key
	exists, err = testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}

	// 清理
	testPool.DeleteValue(key)
}

// TestHGetAllValue 测试获取Hash的所有字段和值
func TestHGetAllValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	// 注意：HGetAllValue 需要先通过其他方式设置hash值
	// 由于当前代码中没有 HSet 方法，这里只测试空值的情况
	key := "test:hash:nonexistent"

	result, err := testPool.HGetAllValue(key)
	if err != nil {
		t.Fatalf("HGetAllValue failed: %v", err)
	}
	if len(result) > 0 {
		t.Fatalf("Expected nil or empty slice for non-existent hash, got: %v", result)
	}
}

// TestMGetValue 测试批量获取值
func TestMGetValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	keys := []string{"test:mget:1", "test:mget:2", "test:mget:3"}
	values := []string{"value1", "value2", "value3"}

	// 先设置值
	for i, key := range keys {
		err := testPool.SetValue(key, values[i], 0)
		if err != nil {
			t.Fatalf("SetValue failed for key %s: %v", key, err)
		}
	}

	// 转换为 []any
	keysAny := make([]any, len(keys))
	for i, k := range keys {
		keysAny[i] = k
	}

	// 批量获取
	results, err := testPool.MGetValue(keysAny)
	if err != nil {
		t.Fatalf("MGetValue failed: %v", err)
	}
	if len(results) != len(keys) {
		t.Fatalf("Expected %d results, got %d", len(keys), len(results))
	}

	// 验证值
	for i, result := range results {
		if result != values[i] {
			t.Fatalf("Expected %s for key %s, got %s", values[i], keys[i], result)
		}
	}

	// 清理
	testPool.DeleteValues(keys)
}

// TestMSetValue 测试批量设置值
func TestMSetValue(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	// 准备键值对：k1,v1,k2,v2,k3,v3
	kv := []any{"test:mset:1", "value1", "test:mset:2", "value2", "test:mset:3", "value3"}

	// 批量设置
	err := testPool.MSetValue(kv)
	if err != nil {
		t.Fatalf("MSetValue failed: %v", err)
	}

	// 验证值
	keys := []string{"test:mset:1", "test:mset:2", "test:mset:3"}
	expectedValues := []string{"value1", "value2", "value3"}

	for i, key := range keys {
		result, err := testPool.GetValue(key)
		if err != nil {
			t.Fatalf("GetValue failed for key %s: %v", key, err)
		}
		if result != expectedValues[i] {
			t.Fatalf("Expected %s for key %s, got %s", expectedValues[i], key, result)
		}
	}

	// 清理
	testPool.DeleteValues(keys)
}

// TestSetExpire 测试设置过期时间
func TestSetExpire(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:expire"
	value := "test_expire_value"

	// 先设置值（不带过期时间）
	err := testPool.SetValue(key, value, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 设置过期时间
	err = testPool.SetExpire(key, 60)
	if err != nil {
		t.Fatalf("SetExpire failed: %v", err)
	}

	// 验证值仍然存在
	result, err := testPool.GetValue(key)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if result != value {
		t.Fatalf("Expected %s, got %s", value, result)
	}

	// 清理
	testPool.DeleteValue(key)
}

// TestMSetValueWithExpire 测试批量设置值并设置过期时间
func TestMSetValueWithExpire(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	// 准备键值对：k1,v1,k2,v2,k3,v3
	kv := []any{"test:mset:expire:1", "value1", "test:mset:expire:2", "value2", "test:mset:expire:3", "value3"}
	expire := 60

	// 批量设置并设置过期时间
	err := testPool.MSetValueWithExpire(kv, expire)
	if err != nil {
		t.Fatalf("MSetValueWithExpire failed: %v", err)
	}

	// 验证值
	keys := []string{"test:mset:expire:1", "test:mset:expire:2", "test:mset:expire:3"}
	expectedValues := []string{"value1", "value2", "value3"}

	for i, key := range keys {
		result, err := testPool.GetValue(key)
		if err != nil {
			t.Fatalf("GetValue failed for key %s: %v", key, err)
		}
		if result != expectedValues[i] {
			t.Fatalf("Expected %s for key %s, got %s", expectedValues[i], key, result)
		}
	}

	// 清理
	testPool.DeleteValues(keys)
}

// TestMSetExpire 测试批量设置过期时间
func TestMSetExpire(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	keys := []string{"test:mset:expire:batch:1", "test:mset:expire:batch:2", "test:mset:expire:batch:3"}
	values := []string{"value1", "value2", "value3"}

	// 先设置值（不带过期时间）
	for i, key := range keys {
		err := testPool.SetValue(key, values[i], 0)
		if err != nil {
			t.Fatalf("SetValue failed for key %s: %v", key, err)
		}
	}

	// 批量设置过期时间
	err := testPool.MSetExpire(keys, 60)
	if err != nil {
		t.Fatalf("MSetExpire failed: %v", err)
	}

	// 验证值仍然存在
	for i, key := range keys {
		result, err := testPool.GetValue(key)
		if err != nil {
			t.Fatalf("GetValue failed for key %s: %v", key, err)
		}
		if result != values[i] {
			t.Fatalf("Expected %s for key %s, got %s", values[i], key, result)
		}
	}

	// 清理
	testPool.DeleteValues(keys)
}

// TestCloseRedisPool 测试关闭连接池
func TestCloseRedisPool(t *testing.T) {
	pool := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
	if pool == nil {
		t.Fatal("NewRedisPoolByDB returned nil")
	}

	// 关闭连接池
	pool.CloseRedisPool()

	// 尝试使用已关闭的连接池应该会失败
	_, err := pool.GetValue("test")
	if err == nil {
		t.Fatal("Expected error when using closed pool")
	}
}

// TestSetValueGetValueFlow 测试完整的设置和获取流程
func TestSetValueGetValueFlow(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	key := "test:flow"
	value := "test_flow_value"
	expire := 100

	// 设置值
	err := testPool.SetValue(key, value, expire)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 验证存在
	exists, err := testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}

	// 获取值
	result, err := testPool.GetValue(key)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if result != value {
		t.Fatalf("Expected %s, got %s", value, result)
	}

	// 删除值
	count, err := testPool.DeleteValue(key)
	if err != nil {
		t.Fatalf("DeleteValue failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 deleted key, got %d", count)
	}

	// 验证已删除
	exists, err = testPool.ExistsValue(key)
	if err != nil {
		t.Fatalf("ExistsValue failed: %v", err)
	}
	if exists {
		t.Fatal("Key should not exist after deletion")
	}
}

// TestMGetValueWithNonExistentKeys 测试批量获取包含不存在的key
func TestMGetValueWithNonExistentKeys(t *testing.T) {
	initTestPool()
	defer cleanupTestPool()

	// 设置一个存在的key
	existingKey := "test:mget:existing"
	existingValue := "existing_value"
	err := testPool.SetValue(existingKey, existingValue, 0)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	// 批量获取包含存在和不存在的key
	keysAny := []any{existingKey, "test:mget:nonexistent1", "test:mget:nonexistent2"}

	results, err := testPool.MGetValue(keysAny)
	if err != nil {
		t.Fatalf("MGetValue failed: %v", err)
	}
	if len(results) != len(keysAny) {
		t.Fatalf("Expected %d results, got %d", len(keysAny), len(results))
	}

	// 验证存在的key
	if results[0] != existingValue {
		t.Fatalf("Expected %s, got %s", existingValue, results[0])
	}

	// 清理
	testPool.DeleteValue(existingKey)
}

// BenchmarkSetValue 性能测试：设置值
func BenchmarkSetValue(b *testing.B) {
	pool := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
	defer pool.CloseRedisPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "bench:set"
		value := "bench_value"
		err := pool.SetValue(key, value, 0)
		if err != nil {
			b.Fatalf("SetValue failed: %v", err)
		}
	}
}

// BenchmarkGetValue 性能测试：获取值
func BenchmarkGetValue(b *testing.B) {
	pool := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
	defer pool.CloseRedisPool()

	// 先设置值
	key := "bench:get"
	value := "bench_value"
	pool.SetValue(key, value, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.GetValue(key)
		if err != nil {
			b.Fatalf("GetValue failed: %v", err)
		}
	}
}

// BenchmarkMGetValue 性能测试：批量获取值
func BenchmarkMGetValue(b *testing.B) {
	pool := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
	defer pool.CloseRedisPool()

	// 先设置值
	keys := []string{"bench:mget:1", "bench:mget:2", "bench:mget:3"}
	for _, key := range keys {
		pool.SetValue(key, "value", 0)
	}

	keysAny := make([]any, len(keys))
	for i, k := range keys {
		keysAny[i] = k
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.MGetValue(keysAny)
		if err != nil {
			b.Fatalf("MGetValue failed: %v", err)
		}
	}
}
