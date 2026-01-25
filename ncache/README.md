# ncache - 新版 Redis 客户端模块

ncache 是 go_base 库中的新版 Redis 客户端模块，使用 `github.com/redis/go-redis/v9` 作为底层驱动，替代了旧版的 `github.com/gomodule/redigo/redis`。

## 特性

- 基于最新的 `github.com/redis/go-redis/v9` 客户端
- 提供与旧版 cache 模块兼容的 API
- 支持 Redis 的各种数据结构操作（字符串、哈希、列表、有序集合等）
- 上下文感知的操作，支持超时控制
- 更好的错误处理和连接管理

## 主要变化

相比旧版 cache 模块，ncache 模块的主要变化包括：

1. 使用 `RedisClient` 替代 `RedisPool`
2. 底层驱动从 redigo 更换为 go-redis/v9
3. 更好的管道和事务支持
4. 更完善的错误处理机制

## 使用示例

### 初始化客户端

```go
import "github.com/yeahyf/go_base/ncache"

// 使用配置对象创建客户端
cfg := &ncache.Config{
    InitConnSize: 1,           // 初始化连接数量
    MaxConnSize:  10,          // 最大连接数
    MaxIdleTime:  300,         // 连接最大空闲时间
    Address:      "127.0.0.1:6379", // 服务器地址
    Username:     "",          // 用户名（可选）
    Password:     "password",  // 密码（可选）
    DBIndex:      0,           // 数据库索引
}

client := ncache.NewClient(cfg)

// 或者使用便捷方法
client := ncache.NewRedisClient(1, 10, 300, "127.0.0.1:6379", "password")
```

### 基本操作

```go
// 设置值（带过期时间）
err := client.SetValue("key", "value", 60) // 60秒过期

// 获取值
value, err := client.GetValue("key")

// 检查键是否存在
exists, err := client.ExistsValue("key")

// 删除键
deleted, err := client.DeleteValue("key")
```

### 高级功能

```go
// 列表操作
err := client.LPush("list_key", "value1", "value2")
popped, err := client.LPop("list_key")

// 哈希操作
err := client.HSet("hash_key", "field", "value")
hashValues, err := client.HGetAllValue("hash_key")

// 有序集合操作
err := client.ZAdd("zset_key", "member", 10.5)
rangeValues, err := client.ZRange("zset_key", 0, -1)
```

## API 兼容性

ncache 模块提供与旧版 cache 模块相似的 API，便于迁移：

- `SetValue` / `GetValue` - 字符串操作
- `HSet` / `HGetAllValue` - 哈希操作
- `LPush` / `LPop` - 列表操作
- `ZAdd` / `ZRange` - 有序集合操作
- `ExistsValue` / `DeleteValue` - 通用操作

## 迁移指南

从旧版 cache 模块迁移到 ncache 模块：

1. 将导入路径从 `github.com/yeahyf/go_base/cache` 更改为 `github.com/yeahyf/go_base/ncache`
2. 将 `cache.RedisPool` 类型替换为 `ncache.RedisClient`
3. 调用 `client.CloseRedisClient()` 替代 `client.CloseRedisPool()`
4. 其他 API 调用保持不变