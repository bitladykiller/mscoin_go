// Package redisx centralizes Redis client creation.
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

// Config is the shared Redis configuration model used by the refactored
// services. The project will keep all Redis initialization in one place so each
// service only depends on a prebuilt client.
type Config struct {
	Addrs    []string
	Password string
	DB       int
}

// Client wraps the raw go-redis universal client.
type Client struct {
	raw goredis.UniversalClient
}

// New creates a shared Redis client.
func New(cfg Config) *Client {
	return &Client{
		raw: goredis.NewUniversalClient(&goredis.UniversalOptions{
			Addrs:    cfg.Addrs,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

// Raw returns the underlying go-redis client for advanced use cases.
func (c *Client) Raw() goredis.UniversalClient {
	return c.raw
}

// Get returns a cached value using the default background context.
func (c *Client) Get(key string, value any) error {
	return c.GetCtx(context.Background(), key, value)
}

// GetCtx loads a cached value and decodes it into the provided target.
//
// Why this helper exists:
//   - legacy MSCoin services store both raw strings and JSON documents in Redis
//   - register and withdraw verification flows expect string reads
//   - centralizing decode logic avoids reimplementing type switches in each
//     service
func (c *Client) GetCtx(ctx context.Context, key string, value any) error {
	bytes, err := c.raw.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return decode(bytes, value)
}

// Set writes a value without expiration using the default background context.
func (c *Client) Set(key string, value any) error {
	return c.SetCtx(context.Background(), key, value)
}

// SetCtx writes a value without expiration.
func (c *Client) SetCtx(ctx context.Context, key string, value any) error {
	return c.set(ctx, key, value, 0)
}

// SetWithExpireCtx writes a value with an explicit TTL.
func (c *Client) SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.set(ctx, key, value, ttl)
}

// SetJSON stores a value as JSON. This is intentionally explicit because most
// project cache payloads are structured objects rather than raw strings.
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
