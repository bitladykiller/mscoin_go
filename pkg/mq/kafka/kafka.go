// Package kafka 为 `go-zero` 服务集中管理 Kafka 生产者构建。
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
	PushWithKey(ctx context.Context, key string, value string) error
	Close() error
}

type producer struct {
	raw *kafka.Writer
}

// NewProducer 创建一个由 `kafka-go` 支持的 Kafka 生产者。
//
// 为什么这个包装器保持项目本地化，而不是直接暴露 `kafka.Writer`：
//   - 服务代码只需要一个小的发布接口
//   - 测试可以轻松地用 fake 生产者替换 Kafka
//   - 未来的迁移步骤可以添加追踪、重试或发件箱集成，而无需更改业务层调用点
func NewProducer(cfg Config) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if cfg.Topic == "" {
		return nil, errors.New("kafka topic is required")
	}

	return &producer{
		raw: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
			Async:                  !cfg.Sync,
		},
	}, nil
}

func (p *producer) PushWithKey(ctx context.Context, key string, value string) error {
	if p == nil || p.raw == nil {
		return errors.New("kafka producer is not initialized")
	}
	return p.raw.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (p *producer) Close() error {
	if p == nil || p.raw == nil {
		return nil
	}
	return p.raw.Close()
}
