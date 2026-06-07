// Package mongox 集中管理 MongoDB 初始化和生命周期管理。
package mongox

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config 描述需要时间序列或弱结构化存储的服务所使用的 MongoDB 连接配置。
type Config struct {
	URI      string
	Username string
	Password string
	Database string
}

// Client 同时保留原始客户端和选择的默认数据库。
type Client struct {
	raw *mongo.Client
	db  *mongo.Database
}

// New 创建一个 Mongo 客户端并立即验证连接，
// 以便服务在启动期间快速失败，而不是在第一个请求时失败。
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	if cfg.Username != "" || cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	raw, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := raw.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return &Client{
		raw: raw,
		db:  raw.Database(cfg.Database),
	}, nil
}

// Database 返回为服务配置的默认数据库。
func (c *Client) Database() *mongo.Database {
	return c.db
}

// Disconnect 优雅地关闭 Mongo 客户端。
func (c *Client) Disconnect(ctx context.Context) error {
	return c.raw.Disconnect(ctx)
}
