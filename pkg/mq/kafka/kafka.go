// Package kafka centralizes Kafka producer construction for `go-zero` services.
package kafka

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
)

// Config defines the minimal producer configuration shared by refactored
// services.
//
// The current migration only needs one producer for the `withdraw` topic, but
// the wrapper keeps the configuration generic so later services can reuse the
// same abstraction.
type Config struct {
	Brokers                []string
	Topic                  string
	Sync                   bool
	AllowAutoTopicCreation bool
}

// Producer is the small interface the business layer depends on.
//
// Why the service layer only sees this interface:
//   - domain code should not know about `kq.Pusher` details
//   - tests can replace Kafka with a fake publisher without booting brokers
//   - future migrations can swap producer internals without touching business
//     orchestration code
type Producer interface {
	PushWithKey(ctx context.Context, key string, value string) error
	Close() error
}

type producer struct {
	raw *kafka.Writer
}

// NewProducer creates a Kafka producer backed by `kafka-go`.
//
// Why this wrapper stays project-local instead of exposing `kafka.Writer`
// directly:
//   - service code only needs a tiny publish interface
//   - tests can replace Kafka with a fake producer trivially
//   - future migration steps can add tracing, retries, or outbox integration
//     without changing business-layer call sites
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
