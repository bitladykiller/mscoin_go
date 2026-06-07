// Package kafka 为 `go-zero` 服务集中管理 Kafka 生产者构建。
//
// 本包提供简化的 Kafka 生产者实现，包括：
//   - 生产者配置管理
//   - 同步和异步发送支持
//   - 消息键支持（用于分区路由）
//   - 优雅关闭
//
// 架构设计：
//   - Producer 接口：抽象消息发送行为，便于测试
//   - producer 结构体：基于 segmentio/kafka-go 的实现
//
// 使用场景：
//   - 发送异步任务消息（如提现请求）
//   - 事件发布
//   - 日志收集
//
// 发送模式：
//   - 同步发送（Sync=true）：等待 Kafka 确认后返回，适合重要消息
//   - 异步发送（Sync=false）：立即返回，适合高吞吐场景
package kafka

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
)

// Config 定义重构后服务共享的最小生产者配置。
//
// 当前迁移只需要一个 `withdraw` 主题的生产者，但包装器保持配置通用，
// 以便后续服务可以复用相同的抽象。
//
// 字段说明：
//   - Brokers: Kafka 集群地址列表
//   - Topic: 发送到的主题名称
//   - Sync: 是否同步发送，true 表示等待 Kafka 确认
//   - AllowAutoTopicCreation: 是否允许自动创建主题（生产环境建议关闭）
type Config struct {
	Brokers                []string
	Topic                  string
	Sync                   bool
	AllowAutoTopicCreation bool
}

// Producer 是业务层依赖的小接口。
//
// 为什么服务层只能看到这个接口：
//   - 领域代码不应该知道 `kq.Pusher` 的细节
//   - 测试可以用 fake 发布者替换 Kafka，而无需启动 broker
//   - 未来的迁移可以更换生产者内部实现，而不触及业务编排代码
type Producer interface {
	// PushWithKey 发送带键的消息。
	//
	// 消息键的作用：
	//   - 用于分区路由：相同键的消息总是发送到同一分区
	//   - 保证顺序：同一键的消息按发送顺序被消费
	//   - 日志压缩：Kafka 可以基于键进行日志压缩
	//
	// 参数：
	//   - ctx: 上下文，用于超时和取消控制
	//   - key: 消息键，可以是用户 ID、订单号等
	//   - value: 消息值，通常是 JSON 格式的业务数据
	//
	// 返回值：
	//   - error: 发送失败时返回错误
	PushWithKey(ctx context.Context, key string, value string) error

	// Close 关闭生产者，释放资源。
	Close() error
}

// producer 是 Producer 接口的默认实现。
type producer struct {
	raw *kafka.Writer
}

// NewProducer 创建一个由 `kafka-go` 支持的 Kafka 生产者。
//
// 为什么这个包装器保持项目本地化，而不是直接暴露 `kafka.Writer`：
//   - 服务代码只需要一个小的发布接口
//   - 测试可以轻松地用 fake 生产者替换 Kafka
//   - 未来的迁移步骤可以添加追踪、重试或发件箱集成，而无需更改业务层调用点
//
// 参数：
//   - cfg: 生产者配置
//
// 返回值：
//   - Producer: 生产者实例
//   - error: 配置验证失败时返回错误
//
// 使用示例：
//
//	producer, err := NewProducer(Config{
//	    Brokers: []string{"localhost:9092"},
//	    Topic:   "withdraw",
//	    Sync:    true, // 同步发送，确保消息可靠
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	defer producer.Close()
//
//	err = producer.PushWithKey(ctx, "user:123", `{"amount": 100}`)
func NewProducer(cfg Config) (Producer, error) {
	// 配置验证
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if cfg.Topic == "" {
		return nil, errors.New("kafka topic is required")
	}

	return &producer{
		raw: &kafka.Writer{
			// Kafka 集群地址
			Addr: kafka.TCP(cfg.Brokers...),
			// 目标主题
			Topic: cfg.Topic,
			// 负载均衡策略：LeastBytes 选择当前消息量最少的分区
			Balancer: &kafka.LeastBytes{},
			// 确认级别：RequireAll 表示等待所有副本确认
			RequiredAcks: kafka.RequireAll,
			// 是否自动创建主题
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
			// 异步发送标志：false 表示同步发送
			Async: !cfg.Sync,
		},
	}, nil
}

// PushWithKey 发送带键的消息到 Kafka。
//
// 参数：
//   - ctx: 上下文
//   - key: 消息键，用于分区路由和顺序保证
//   - value: 消息值
//
// 返回值：
//   - error: 发送失败或生产者未初始化时返回错误
func (p *producer) PushWithKey(ctx context.Context, key string, value string) error {
	// 检查生产者是否已初始化
	if p == nil || p.raw == nil {
		return errors.New("kafka producer is not initialized")
	}

	// 写入消息
	return p.raw.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

// Close 关闭 Kafka 生产者，释放连接资源。
//
// 返回值：
//   - error: 关闭失败时返回错误，未初始化时返回 nil
func (p *producer) Close() error {
	if p == nil || p.raw == nil {
		return nil
	}
	return p.raw.Close()
}