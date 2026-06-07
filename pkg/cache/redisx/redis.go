// Package redisx 集中管理 Redis 客户端创建。
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
// 项目将所有 Redis 初始化集中在一个地方，以便每个服务只依赖于预构建的客户端。
type Config struct {
	Addrs    []string
	Password string
	DB       int
}

// Client 包装原始的 go-redis 通用客户端。
type Client struct {
	raw goredis.UniversalClient
}

// New 创建一个共享的 Redis 客户端。
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
func (c *Client) Raw() goredis.UniversalClient {
	return c.raw
}

// Get 使用默认后台上下文返回缓存的值。
func (c *Client) Get(key string, value any) error {
	return c.GetCtx(context.Background(), key, value)
}

// GetCtx 加载缓存的值并将其解码到提供的目标中。
//
// 为什么需要这个辅助函数：
//   - 传统 MSCoin 服务在 Redis 中同时存储原始字符串和 JSON 文档
//   - 注册和提现验证流程期望读取字符串
//   - 集中解码逻辑可以避免在每个服务中重复实现类型判断
func (c *Client) GetCtx(ctx context.Context, key string, value any) error {
	bytes, err := c.raw.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return decode(bytes, value)
}

// Set 使用默认后台上下文写入一个值，不设置过期时间。
func (c *Client) Set(key string, value any) error {
	return c.SetCtx(context.Background(), key, value)
}

// SetCtx 写入一个值，不设置过期时间。
func (c *Client) SetCtx(ctx context.Context, key string, value any) error {
	return c.set(ctx, key, value, 0)
}

// SetWithExpireCtx 写入一个值并设置显式的 TTL。
func (c *Client) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.set(ctx, key, value, ttl)
}

// SetJSON 将值存储为 JSON。这是故意显式的，因为大多数项目缓存负载是结构化对象而非原始字符串。
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal redis value: %w", err)
	}
	return c.raw.Set(ctx, key, bytes, ttl).Err()
}

func (c *Client) set(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := encode(value)
	if err != nil {
		return err
	}
	return c.raw.Set(ctx, key, encoded, ttl).Err()
}

func encode(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	}

	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return "", nil
	}
	switch rv.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(value), nil
	}

	bytes, err := json.Marshal(value)
	return string(bytes), err
}

func decode(data []byte, value any) error {
	switch v := value.(type) {
	case *string:
		*v = string(data)
		return nil
	case *[]byte:
		*v = append((*v)[:0], data...)
		return nil
	}
	if err := json.Unmarshal(data, value); err == nil {
		return nil
	}
	return errors.New("redis decode failed")
}
