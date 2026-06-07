// Package kafka centralizes Kafka producer and consumer construction for
// refactored go-zero services.
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

// ConsumeAction describes how the consumer loop should handle one processing
// result.
//
// Why the consumer loop uses explicit actions instead of only checking whether
// an error exists:
//   - async workflows must distinguish retryable failures from poison messages
//   - dead-lettering should be a framework concern rather than duplicated in
//     each business consumer
//   - handlers stay focused on domain behavior while the queue layer owns
//     offset commit policy
type ConsumeAction int

const (
	// ConsumeAck commits the current offset immediately.
	ConsumeAck ConsumeAction = iota
	// ConsumeRetry keeps the offset uncommitted and retries the same message
	// locally after a backoff delay.
	ConsumeRetry
	// ConsumeDeadLetter publishes the original message into the configured
	// dead-letter topic and then commits the offset.
	ConsumeDeadLetter
)

// Message is the queue-agnostic payload exposed to business handlers.
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string][]byte
	Partition int
	Offset    int64
	Time      time.Time
}

// Handler processes one Kafka message.
type Handler func(ctx context.Context, message Message) error

// ErrorClassifier converts a handler result into one queue action.
type ErrorClassifier func(err error) ConsumeAction

// ConsumerConfig contains the shared Kafka consumer settings.
//
// The configuration intentionally keeps only the fields the migration is using
// now. More Kafka tuning knobs can be added later from one place without
// rewriting individual consumers.
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

// NewConsumerService creates a go-zero-compatible background consumer service.
//
// The returned value implements `service.Service`, so callers can register it in
// one `service.ServiceGroup` together with other workers and scheduled tasks.
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

// Start blocks and keeps consuming messages until Stop is called.
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

// Stop cancels the consume loop and closes underlying clients.
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
