// Package redisx 集中管理 Redis 客户端创建和常用缓存操作。
//
// 本包封装了 github.com/go-redis/redis/v8 库，提供：
//   - 统一的客户端创建方式
//   - 简化的 Get/Set 操作，支持多种数据类型自动编码
//   - JSON 格式存储支持
//   - 上下文支持，便于超时控制
//
// 使用场景：
//   - 存储用户 Session 信息
//   - 缓存热点数据
//   - 分布式锁实现
//   - 验证码存储
package redisx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

// Config 是重构后服务使用的共享 Redis 配置模型。
//
// 项目将所有 Redis 初始化集中在一个地方，以便每个服务只依赖于预构建的客户端。
//
// 字段说明：
//   - Addrs: Redis 服务器地址列表，支持单机和集群模式
//   - Password: Redis 认证密码，如果 Redis 未启用认证则留空
//   - DB: Redis 数据库编号，范围 0-15，不同服务可使用不同 DB 避免键冲突
type Config struct {
	Addrs    []string
	Password string
	DB       int
}

// Client 包装原始的 go-redis 通用客户端。
// 它提供了更简洁的 API，自动处理常见的数据编码问题。
type Client struct {
	raw goredis.UniversalClient
}

// New 创建一个共享的 Redis 客户端。
//
// 该函数使用 UniversalClient，可以自动适应单机或集群模式，
// 由配置中的 Addrs 数量决定。
//
// 参数：
//   - cfg: Redis 配置
//
// 返回值：
//   - *Client: Redis 客户端实例
//
// 使用示例：
//
//	client := New(Config{
//	    Addrs:    []string{"127.0.0.1:6379"},
//	    Password: "",
//	    DB:       0,
//	})
func New(cfg Config) *Client {
	return &Client{
		raw: goredis.NewUniversalClient(&goredis.UniversalOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

// Raw 返回底层的 go-redis 客户端，用于高级使用场景。
//
// 当需要使用本包未封装的高级功能（如 Lua 脚本、Pipeline、Pub/Sub）时，
// 可以通过此方法获取原始客户端。
//
// 返回值：
//   - goredis.UniversalClient: 原始 Redis 客户端
func (c *Client) Raw() goredis.UniversalClient {
	return c.raw
}

// Get 使用默认后台上下文返回缓存的值。
// 这是一个便捷方法，适用于不需要超时控制的场景。
//
// 参数：
//   - key: 缓存键
//   - value: 目标值指针，用于接收解码后的数据
//
// 返回值：
//   - error: 缓存不存在或解码失败时返回错误
func (c *Client) Get(key string, value any) error {
	return c.GetCtx(context.Background(), key, value)
}

// GetCtx 加载缓存的值并将其解码到提供的目标中。
//
// 为什么需要这个辅助函数：
//   - 传统 MSCoin 服务在 Redis 中同时存储原始字符串和 JSON 文档
//   - 注册和提现验证流程期望读取字符串
//   - 集中解码逻辑可以避免在每个服务中重复实现类型判断
//
// 参数：
//   - ctx: 上下文，用于超时和取消控制
//   - key: 缓存键
//   - value: 目标值指针，可以是 string、[]byte 或任意结构体
//
// 返回值：
//   - error: 缓存不存在或解码失败时返回错误（包括 redis.Nil 表示键不存在）
//
// 使用示例：
//
//	var token string
//	if err := client.GetCtx(ctx, "user:123:token", &token); err != nil {
//	    // 处理错误
//	}
func (c *Client) GetCtx(ctx context.Context, key string, value any) error {
	bytes, err := c.raw.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return decode(bytes, value)
}

// Set 使用默认后台上下文写入一个值，不设置过期时间。
//
// 参数：
//   - key: 缓存键
//   - value: 要存储的值，可以是字符串、数字或结构体
//
// 返回值：
//   - error: 写入失败时返回错误
func (c *Client) Set(key string, value any) error {
	return c.SetCtx(context.Background(), key, value)
}

// SetCtx 写入一个值，不设置过期时间。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 要存储的值
//
// 返回值：
//   - error: 写入失败时返回错误
func (c *Client) SetCtx(ctx context.Context, key string, value any) error {
	return c.set(ctx, key, value, 0)
}

// SetWithExpireCtx 写入一个值并设置显式的 TTL。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 要存储的值
//   - ttl: 过期时间，0 表示永不过期
//
// 返回值：
//   - error: 写入失败时返回错误
//
// 使用示例：
//
//	err := client.SetWithExpireCtx(ctx, "verify:13800138000", "123456", 5*time.Minute)
func (c *Client) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.set(ctx, key, value, ttl)
}

// SetJSON 将值存储为 JSON。这是故意显式的，因为大多数项目缓存负载是结构化对象而非原始字符串。
//
// 参数：
//   - ctx: 上下文
//   - key: 缓存键
//   - value: 要存储的值（会被 JSON 编码）
//   - ttl: 过期时间
//
// 返回值：
//   - error: JSON 编码或写入失败时返回错误
//
// 使用示例：
//
//	user := User{ID: 123, Name: "test"}
//	err := client.SetJSON(ctx, "user:123", user, 1*time.Hour)
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal redis value: %w", err)
	}
	return c.raw.Set(ctx, key, bytes, ttl).Err()
}

// set 是内部通用写入方法。
func (c *Client) set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := encode(value)
	if err != nil {
		return err
	}
	return c.raw.Set(ctx, key, encoded, ttl).Err()
}

// encode 将值编码为适合 Redis 存储的字符串格式。
//
// 编码规则：
//   - string: 直接存储
//   - []byte: 转换为字符串
//   - fmt.Stringer: 调用 String() 方法
//   - 基本数值类型: 使用 fmt.Sprint 格式化
//   - 其他类型: JSON 编码
//
// 参数：
//   - value: 要编码的值
//
// 返回值：
//   - string: 编码后的字符串
//   - error: 编码失败时返回错误
func encode(value any) (string, error) {
	// 特殊类型快速路径
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	}

	// 反射检查基本数值类型
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		// nil 值返回空字符串
		return "", nil
	}
	switch rv.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(value), nil
	}

	// 其他类型使用 JSON 编码
	bytes, err := json.Marshal(value)
	return string(bytes), err
}

// decode 将 Redis 存储的字节解码到目标值。
//
// 解码规则：
//   - *string: 直接转换为字符串
//   - *[]byte: 复制字节（避免共享底层数组）
//   - 其他类型: 尝试 JSON 解码
//
// 参数：
//   - data: Redis 存储的原始字节
//   - value: 目标值指针
//
// 返回值：
//   - error: 解码失败时返回错误
func decode(data []byte, value any) error {
	// 特殊类型快速路径
	switch v := value.(type) {
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		// 复制字节，避免与 Redis 内部缓冲区共享
		*v = append((*v)[:0], data...)
		return nil
	}

	// 尝试 JSON 解码
	if err := json.Unmarshal(data, value); err == nil {
		return nil
	}

	// 解码失败
	return errors.New("redis decode failed")
}