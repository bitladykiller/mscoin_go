// Package mongox 集中管理 MongoDB 初始化和生命周期管理。
//
// 本包封装了 MongoDB Go Driver，提供：
//   - 统一的客户端创建方式
//   - 连接验证（启动时快速失败）
//   - 默认数据库访问
//   - 优雅关闭
//
// 使用场景：
//   - 存储时间序列数据（如交易记录）
//   - 存储弱结构化数据（如用户配置）
//   - 高吞吐量的日志存储
//
// 注意事项：
//   - MongoDB 适合存储灵活的文档数据
//   - 事务支持较弱，复杂事务建议使用 MySQL
//   - 建议为不同服务使用不同的数据库以隔离数据
package mongox

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config 描述需要时间序列或弱结构化存储的服务所使用的 MongoDB 连接配置。
//
// 字段说明：
//   - URI: MongoDB 连接 URI，格式：mongodb://host:port 或 mongodb+srv://host
//   - Username: 认证用户名，如果 MongoDB 未启用认证则留空
//   - Password: 认证密码，如果 MongoDB 未启用认证则留空
//   - Database: 默认数据库名称，后续操作会使用此数据库
type Config struct {
	URI      string // MongoDB 连接 URI
	Username string // 认证用户名（可选）
	Password string // 认证密码（可选）
	Database string // 默认数据库名称
}

// Client 同时保留原始客户端和选择的默认数据库。
//
// 这种设计允许：
//   - 通过 Database() 方法快速访问默认数据库
//   - 通过 Disconnect() 方法优雅关闭连接
//   - 未来可以扩展添加连接池监控等功能
type Client struct {
	raw *mongo.Client    // 原始 MongoDB 客户端
	db  *mongo.Database  // 默认数据库引用
}

// New 创建一个 Mongo 客户端并立即验证连接，
// 以便服务在启动期间快速失败，而不是在第一个请求时失败。
//
// 连接验证步骤：
//  1. 解析连接 URI
//  2. 配置认证（如果提供了凭据）
//  3. 建立连接
//  4. 发送 Ping 命令验证连接可用性
//
// 参数：
//   - cfg: MongoDB 配置
//
// 返回值：
//   - *Client: MongoDB 客户端实例
//   - error: 连接或验证失败时返回错误
//
// 使用示例：
//
//	client, err := mongox.New(mongox.Config{
//	    URI:      "mongodb://localhost:27017",
//	    Username: "admin",
//	    Password: "password",
//	    Database: "mscoin",
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	defer client.Disconnect(context.Background())
func New(cfg Config) (*Client, error) {
	// 创建带超时的上下文，避免启动时阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 配置客户端选项
	clientOptions := options.Client().ApplyURI(cfg.URI)

	// 如果提供了认证信息，设置凭据
	if cfg.Username != "" || cfg.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	// 建立连接
	raw, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	// 验证连接可用性
	// 这确保服务启动时快速失败，而不是在第一个请求时才发现连接问题
	if err := raw.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return &Client{
		raw: raw,
		db:  raw.Database(cfg.Database),
	}, nil
}

// Database 返回为服务配置的默认数据库。
//
// 返回值：
//   - *mongo.Database: 默认数据库实例
//
// 使用示例：
//
//	collection := client.Database().Collection("users")
//	_, err := collection.InsertOne(ctx, user)
func (c *Client) Database() *mongo.Database {
	return c.db
}

// Disconnect 优雅地关闭 Mongo 客户端。
//
// 该方法会：
//   - 关闭所有连接
//   - 释放资源
//   - 确保所有进行中的操作完成
//
// 参数：
//   - ctx: 上下文，用于超时控制
//
// 返回值：
//   - error: 关闭失败时返回错误
//
// 使用示例：
//
//	defer func() {
//	    if err := client.Disconnect(context.Background()); err != nil {
//	        log.Printf("disconnect mongo failed: %v", err)
//	    }
//	}()
func (c *Client) Disconnect(ctx context.Context) error {
	return c.raw.Disconnect(ctx)
}