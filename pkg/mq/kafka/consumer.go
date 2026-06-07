// Package kafka 为重构后的 go-zero 服务集中管理 Kafka 生产者和消费者的构建。
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rawkafka "github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultConsumerMinBytes     = 10 * 1024
	defaultConsumerMaxBytes     = 10 * 1024 * 1024
	defaultConsumerMaxWait      = 3 * time.Second
	defaultConsumerRetryBackoff = 2 * time.Second
)

// ConsumeAction 描述消费者循环应如何处理一个处理结果。
//
// 为什么消费者循环使用显式动作而非仅检查错误是否存在：
//   - 异步工作流必须区分可重试失败和毒消息
//   - 死信应该是框架关注点，而非在每个业务消费者中重复
//   - 处理器专注于领域行为，而队列层拥有 offset 提交策略
type ConsumeAction int

const (
	// ConsumeAck 立即提交当前 offset。
	ConsumeAck ConsumeAction = iota
	// ConsumeRetry 保持 offset 未提交，在退避延迟后本地重试同一条消息。
	ConsumeRetry
	// ConsumeDeadLetter 将原始消息发布到配置的死信主题，然后提交 offset。
	ConsumeDeadLetter
)

// Message 是暴露给业务处理器的队列无关负载。
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Partition int
	Offset    int64
	Time      time.Time
}

// Handler 处理一条 Kafka 消息。
type Handler func(ctx context.Context, message Message) error

// ErrorClassifier 将处理器结果转换为一种队列动作。
type ErrorClassifier func(err error) ConsumeAction

// ConsumerConfig 包含共享的 Kafka 消费者设置。
//
// 配置故意只保留迁移当前使用的字段。后续可以从一个地方添加更多 Kafka 调优参数，
// 而无需重写各个消费者。
type ConsumerConfig struct {
	Brokers              []string
	Topic                string
	GroupID              string
	MinBytes             int
	MaxBytes             int
	MaxWaitMs            int
	RetryBackoffMs       int
	StartOffset          int64
	DeadLetterTopic      string
	AllowAutoTopicCreate bool
}

type reader interface {
	FetchMessage(ctx context.Context) (rawkafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...rawkafka.Message) error
	Close() error
}

type consumerService struct {
	cfg           ConsumerConfig
	reader        reader
	deadLetter    Producer
	handler       Handler
	classifier    ErrorClassifier
	ctx           context.Context
	cancel        context.CancelFunc
	commitTimeout time.Duration
}

type deadLetterPayload struct {
	SourceTopic string            `json:"sourceTopic"`
	Key         []byte            `json:"key"`
	Value       []byte            `json:"value"`
	Headers     map[string][]byte `json:"headers"`
	Partition   int               `json:"partition"`
	Offset      int64             `json:"offset"`
	Error       string            `json:"error"`
	FailedAt    time.Time         `json:"failedAt"`
}

// NewConsumerService 创建一个兼容 go-zero 的后台消费者服务。
//
// 返回值实现 `service.Service`，因此调用者可以将它与其他 worker 和定时任务
// 一起注册到一个 `service.ServiceGroup` 中。
func NewConsumerService(cfg ConsumerConfig, handler Handler, classifier ErrorClassifier) (*consumerService, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if cfg.Topic == "" {
		return nil, errors.New("kafka topic is required")
	}
	if cfg.GroupID == "" {
		return nil, errors.New("kafka group id is required")
	}
	if handler == nil {
		return nil, errors.New("kafka handler is required")
	}
	if classifier == nil {
		classifier = defaultClassifier
	}

	ctx, cancel := context.WithCancel(context.Background())

	readerConfig := rawkafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    firstPositive(cfg.MinBytes, defaultConsumerMinBytes),
		MaxBytes:    firstPositive(cfg.MaxBytes, defaultConsumerMaxBytes),
		MaxWait:     time.Duration(firstPositive(cfg.MaxWaitMs, int(defaultConsumerMaxWait/time.Millisecond))) * time.Millisecond,
		StartOffset: cfg.StartOffset,
	}

	var deadLetter Producer
	var err error
	if cfg.DeadLetterTopic != "" {
		deadLetter, err = NewProducer(Config{
			Brokers:                cfg.Brokers,
			Topic:                  cfg.DeadLetterTopic,
			Sync:                   true,
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreate,
		})
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create dead letter producer: %w", err)
		}
	}

	return &consumerService{
		cfg:           cfg,
		reader:        rawkafka.NewReader(readerConfig),
		deadLetter:    deadLetter,
		handler:       handler,
		classifier:    classifier,
		ctx:           ctx,
		cancel:        cancel,
		commitTimeout: 5 * time.Second,
	}, nil
}

// Start 阻塞并持续消费消息，直到调用 Stop。
func (s *consumerService) Start() {
	logx.Infof("starting kafka consumer, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)

	for {
		message, err := s.reader.FetchMessage(s.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logx.Infof("kafka consumer stopped, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)
				return
			}
			logx.Errorf("fetch kafka message failed, topic=%s group=%s err=%v", s.cfg.Topic, s.cfg.GroupID, err)
			s.sleepBackoff()
			continue
		}

		if err := s.handleFetchedMessage(message); err != nil {
			if errors.Is(err, context.Canceled) {
				logx.Infof("kafka consumer canceled while handling message, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)
				return
			}
			logx.Errorf("handle kafka message aborted, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
			s.sleepBackoff()
			continue
		}
	}
}

// Stop 取消消费循环并关闭底层客户端。
func (s *consumerService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.reader != nil {
		_ = s.reader.Close()
	}
	if s.deadLetter != nil {
		_ = s.deadLetter.Close()
	}
}

func (s *consumerService) handleFetchedMessage(message rawkafka.Message) error {
	payload := toMessage(message)

	for {
		err := s.handler(s.ctx, payload)
		switch s.classifier(err) {
		case ConsumeAck:
			return s.commit(message)
		case ConsumeDeadLetter:
			if err := s.publishDeadLetter(message, err); err != nil {
				logx.Errorf("publish dead letter failed, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
				s.sleepBackoff()
				continue
			}
			return s.commit(message)
		default:
			if err != nil {
				logx.Errorf("retry kafka message, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
			}
			s.sleepBackoff()
			if s.ctx.Err() != nil {
				return s.ctx.Err()
			}
		}
	}
}

func (s *consumerService) publishDeadLetter(message rawkafka.Message, cause error) error {
	if s.deadLetter == nil {
		logx.Errorf("dead letter topic not configured, acking poison message topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, cause)
		return nil
	}

	body, err := json.Marshal(deadLetterPayload{
		SourceTopic: message.Topic,
		Key:         message.Key,
		Value:       message.Value,
		Headers:     toHeaders(message.Headers),
		Partition:   message.Partition,
		Offset:      message.Offset,
		Error:       cause.Error(),
		FailedAt:    time.Now(),
	})
	if err != nil {
		return fmt.Errorf("marshal dead letter payload: %w", err)
	}

	return s.deadLetter.PushWithKey(s.ctx, string(message.Key), string(body))
}

func (s *consumerService) commit(message rawkafka.Message) error {
	commitCtx, cancel := context.WithTimeout(context.Background(), s.commitTimeout)
	defer cancel()

	if err := s.reader.CommitMessages(commitCtx, message); err != nil {
		return fmt.Errorf("commit kafka message: %w", err)
	}
	return nil
}

func (s *consumerService) sleepBackoff() {
	backoff := time.Duration(firstPositive(s.cfg.RetryBackoffMs, int(defaultConsumerRetryBackoff/time.Millisecond))) * time.Millisecond
	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-s.ctx.Done():
	case <-timer.C:
	}
}

func defaultClassifier(err error) ConsumeAction {
	if err == nil {
		return ConsumeAck
	}
	return ConsumeRetry
}

func firstPositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func toMessage(message rawkafka.Message) Message {
	return Message{
		Topic:     message.Topic,
		Key:       append([]byte(nil), message.Key...),
		Value:     append([]byte(nil), message.Value...),
		Headers:   toHeaders(message.Headers),
		Partition: message.Partition,
		Offset:    message.Offset,
		Time:      message.Time,
	}
}

func toHeaders(headers []rawkafka.Header) map[string][]byte {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string][]byte, len(headers))
	for _, header := range headers {
		result[header.Key] = append([]byte(nil), header.Value...)
	}
	return result
}
