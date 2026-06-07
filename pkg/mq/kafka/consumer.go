// Package kafka 为重构后的 go-zero 服务集中管理 Kafka 消费者构建。
//
// 本包提供完整的 Kafka 消费者服务实现，包括：
//   - 消费者配置管理
//   - 消息处理循环
//   - 错误分类和重试策略
//   - 死信队列支持
//   - 优雅停止
//
// 架构设计：
//   - ConsumeAction 枚举：定义消息处理结果的三种动作
//   - Handler 接口：业务处理器，专注于消息处理逻辑
//   - ErrorClassifier 函数：将业务错误转换为队列动作
//   - ConsumerService 结构体：实现 go-zero 的 service.Service 接口
//
// 使用场景：
//   - 处理异步任务（如提现请求）
//   - 事件驱动架构
//   - 日志收集和处理
//
// 消息处理流程：
//  1. 从 Kafka 拉取消息
//  2. 调用 Handler 处理消息
//  3. 根据 ErrorClassifier 决定下一步动作
//  4. Ack -> 提交 offset
//  5. Retry -> 等待退避时间后重试
//  6. DeadLetter -> 发送到死信队列并提交 offset
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

// 默认消费者配置常量
const (
	// defaultConsumerMinBytes 是每次拉取的最小字节数
	// 设置为 10KB 可以减少小批量请求
	defaultConsumerMinBytes = 10 * 1024

	// defaultConsumerMaxBytes 是每次拉取的最大字节数
	// 设置为 10MB 可以提高吞吐量
	defaultConsumerMaxBytes = 10 * 1024 * 1024

	// defaultConsumerMaxWait 是等待消息的最长时间
	// 设置为 3 秒可以平衡延迟和效率
	defaultConsumerMaxWait = 3 * time.Second

	// defaultConsumerRetryBackoff 是重试的退避时间
	// 设置为 2 秒可以避免过于频繁的重试
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
	// 用于消息处理成功的情况。
	ConsumeAck ConsumeAction = iota

	// ConsumeRetry 保持 offset 未提交，在退避延迟后本地重试同一条消息。
	// 用于临时性错误（如数据库连接超时）。
	ConsumeRetry

	// ConsumeDeadLetter 将原始消息发布到配置的死信主题，然后提交 offset。
	// 用于永久性错误（如消息格式错误），避免阻塞正常消息处理。
	ConsumeDeadLetter
)

// Message 是暴露给业务处理器的队列无关负载。
//
// 这种抽象设计使得：
//   - 业务代码不依赖具体的 Kafka 客户端类型
//   - 便于测试时使用 mock 数据
//   - 未来可以切换到其他消息队列而不影响业务代码
type Message struct {
	Topic     string            // 消息所属的主题
	Key       []byte            // 消息键，用于分区路由
	Value     []byte            // 消息值，业务数据的主体
	Headers   map[string][]byte // 消息头，用于传递元数据
	Partition int               // 分区号
	Offset    int64             // 偏移量
	Time      time.Time         // 消息时间戳
}

// Handler 处理一条 Kafka 消息。
//
// 实现者应该：
//   - 专注于业务逻辑处理
//   - 返回 nil 表示成功
//   - 返回错误由 ErrorClassifier 决定后续动作
//
// 参数：
//   - ctx: 上下文，用于取消操作
//   - message: 消息内容
//
// 返回值：
//   - error: 处理失败时返回错误
type Handler func(ctx context.Context, message Message) error

// ErrorClassifier 将处理器结果转换为一种队列动作。
//
// 为什么需要这个函数：
//   - 不同类型的错误需要不同的处理策略
//   - 业务代码决定错误的语义，框架决定队列行为
//   - 例如：数据库临时故障应重试，数据格式错误应进入死信队列
//
// 参数：
//   - err: 处理器返回的错误，nil 表示成功
//
// 返回值：
//   - ConsumeAction: 下一步动作
type ErrorClassifier func(err error) ConsumeAction

// ConsumerConfig 包含共享的 Kafka 消费者设置。
//
// 配置故意只保留迁移当前使用的字段。后续可以从一个地方添加更多 Kafka 调优参数，
// 而无需重写各个消费者。
//
// 字段说明：
//   - Brokers: Kafka 集群地址列表
//   - Topic: 订阅的主题名称
//   - GroupID: 消费者组 ID，同组消费者共享消息
//   - MinBytes: 每次拉取的最小字节数
//   - MaxBytes: 每次拉取的最大字节数
//   - MaxWaitMs: 等待消息的最长时间（毫秒）
//   - RetryBackoffMs: 重试退避时间（毫秒）
//   - StartOffset: 起始偏移量，-1 表示最新，-2 表示最早
//   - DeadLetterTopic: 死信队列主题，留空表示不启用
//   - AllowAutoTopicCreate: 是否允许自动创建主题
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

// reader 是 Kafka 消费者读取接口的抽象。
// 这是一个内部接口，用于解耦消费者服务与具体的 Kafka 客户端实现。
type reader interface {
	FetchMessage(ctx context.Context) (rawkafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...rawkafka.Message) error
	Close() error
}

// consumerService 是消费者服务的实现。
// 它实现了 go-zero 的 service.Service 接口，可以注册到 ServiceGroup 中。
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

// deadLetterPayload 是发送到死信队列的消息格式。
// 它包含原始消息的所有信息和失败原因，便于后续分析和重试。
type deadLetterPayload struct {
	SourceTopic string            `json:"sourceTopic"` // 原始主题
	Key         []byte            `json:"key"`         // 原始消息键
	Value       []byte            `json:"value"`       // 原始消息值
	Headers     map[string][]byte `json:"headers"`     // 原始消息头
	Partition   int               `json:"partition"`   // 原始分区
	Offset      int64             `json:"offset"`      // 原始偏移量
	Error       string            `json:"error"`       // 错误信息
	FailedAt    time.Time         `json:"failedAt"`    // 失败时间
}

// NewConsumerService 创建一个兼容 go-zero 的后台消费者服务。
//
// 返回值实现 `service.Service`，因此调用者可以将它与其他 worker 和定时任务
// 一起注册到一个 `service.ServiceGroup` 中。
//
// 参数：
//   - cfg: 消费者配置
//   - handler: 消息处理函数
//   - classifier: 错误分类函数，nil 时使用默认分类器（成功 Ack，失败 Retry）
//
// 返回值：
//   - *consumerService: 消费者服务实例
//   - error: 配置验证失败时返回错误
//
// 使用示例：
//
//	service, err := NewConsumerService(
//	    ConsumerConfig{
//	        Brokers:  []string{"localhost:9092"},
//	        Topic:    "withdraw",
//	        GroupID:  "withdraw-processor",
//	    },
//	    func(ctx context.Context, msg Message) error {
//	        // 处理消息
//	        return nil
//	    },
//	    nil, // 使用默认错误分类器
//	)
//	if err != nil {
//	    // 处理错误
//	}
//	service.Start() // 阻塞运行
func NewConsumerService(cfg ConsumerConfig, handler Handler, classifier ErrorClassifier) (*consumerService, error) {
	// 配置验证
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

	// 使用默认错误分类器
	if classifier == nil {
		classifier = defaultClassifier
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 配置 Kafka 读取器
	readerConfig := rawkafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    firstPositive(cfg.MinBytes, defaultConsumerMinBytes),
		MaxBytes:    firstPositive(cfg.MaxBytes, defaultConsumerMaxBytes),
		MaxWait:     time.Duration(firstPositive(cfg.MaxWaitMs, int(defaultConsumerMaxWait/time.Millisecond))) * time.Millisecond,
		StartOffset: cfg.StartOffset,
	}

	// 创建死信队列生产者（如果配置了）
	var deadLetter Producer
	var err error
	if cfg.DeadLetterTopic != "" {
		deadLetter, err = NewProducer(Config{
			Brokers:                cfg.Brokers,
			Topic:                  cfg.DeadLetterTopic,
			Sync:                   true, // 死信队列使用同步发送确保可靠
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
//
// 这是 go-zero service.Service 接口的要求。
// 消费循环会持续运行直到：
//   - 调用 Stop()
//   - 上下文被取消
func (s *consumerService) Start() {
	logx.Infof("starting kafka consumer, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)

	for {
		// 从 Kafka 拉取消息
		message, err := s.reader.FetchMessage(s.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logx.Infof("kafka consumer stopped, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)
				return
			}
			// 拉取失败，记录日志并等待退避后重试
			logx.Errorf("fetch kafka message failed, topic=%s group=%s err=%v", s.cfg.Topic, s.cfg.GroupID, err)
			s.sleepBackoff()
			continue
		}

		// 处理拉取到的消息
		if err := s.handleFetchedMessage(message); err != nil {
			if errors.Is(err, context.Canceled) {
				logx.Infof("kafka consumer canceled while handling message, topic=%s group=%s", s.cfg.Topic, s.cfg.GroupID)
				return
			}
			// 处理失败，记录日志并等待退避后重试
			logx.Errorf("handle kafka message aborted, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
			s.sleepBackoff()
			continue
		}
	}
}

// Stop 取消消费循环并关闭底层客户端。
//
// 这是 go-zero service.Service 接口的要求。
// 该方法会：
//   - 取消上下文，停止消费循环
//   - 关闭 Kafka 读取器
//   - 关闭死信队列生产者（如果有）
func (s *consumerService) Stop() {
	// 取消上下文，通知消费循环停止
	if s.cancel != nil {
		s.cancel()
	}
	// 关闭 Kafka 读取器
	if s.reader != nil {
		_ = s.reader.Close()
	}
	// 关闭死信队列生产者
	if s.deadLetter != nil {
		_ = s.deadLetter.Close()
	}
}

// handleFetchedMessage 处理单条消息，包括重试和死信逻辑。
//
// 该方法会持续重试直到：
//   - 消息成功处理（返回 nil）
//   - 消息进入死信队列
//   - 上下文被取消
func (s *consumerService) handleFetchedMessage(message rawkafka.Message) error {
	payload := toMessage(message)

	for {
		// 调用业务处理器
		err := s.handler(s.ctx, payload)

		// 根据错误分类决定下一步动作
		switch s.classifier(err) {
		case ConsumeAck:
			// 成功，提交 offset
			return s.commit(message)
		case ConsumeDeadLetter:
			// 发送到死信队列
			if err := s.publishDeadLetter(message, err); err != nil {
				// 死信发送失败，等待后重试
				logx.Errorf("publish dead letter failed, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
				s.sleepBackoff()
				continue
			}
			// 死信发送成功，提交原始消息的 offset
			return s.commit(message)
		default:
			// 重试
			if err != nil {
				logx.Errorf("retry kafka message, topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, err)
			}
			s.sleepBackoff()
			// 检查是否需要停止
			if s.ctx.Err() != nil {
				return s.ctx.Err()
			}
		}
	}
}

// publishDeadLetter 将失败消息发送到死信队列。
func (s *consumerService) publishDeadLetter(message rawkafka.Message, cause error) error {
	// 如果没有配置死信队列，直接确认消息（丢弃毒消息）
	if s.deadLetter == nil {
		logx.Errorf("dead letter topic not configured, acking poison message topic=%s group=%s offset=%d err=%v", s.cfg.Topic, s.cfg.GroupID, message.Offset, cause)
		return nil
	}

	// 构造死信载荷
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

	// 发送到死信队列
	return s.deadLetter.PushWithKey(s.ctx, string(message.Key), string(body))
}

// commit 提交消息的 offset。
func (s *consumerService) commit(message rawkafka.Message) error {
	commitCtx, cancel := context.WithTimeout(context.Background(), s.commitTimeout)
	defer cancel()

	if err := s.reader.CommitMessages(commitCtx, message); err != nil {
		return fmt.Errorf("commit kafka message: %w", err)
	}
	return nil
}

// sleepBackoff 等待退避时间，可被上下文取消中断。
func (s *consumerService) sleepBackoff() {
	backoff := time.Duration(firstPositive(s.cfg.RetryBackoffMs, int(defaultConsumerRetryBackoff/time.Millisecond))) * time.Millisecond
	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-s.ctx.Done():
	case <-timer.C:
	}
}

// defaultClassifier 是默认的错误分类器。
// 成功返回 Ack，失败返回 Retry。
func defaultClassifier(err error) ConsumeAction {
	if err == nil {
		return ConsumeAck
	}
	return ConsumeRetry
}

// firstPositive 返回第一个正整数，如果都不是正数则返回 fallback。
func firstPositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

// toMessage 将 Kafka 原始消息转换为业务消息结构。
// 使用 append 创建新的切片，避免共享底层数组。
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

// toHeaders 将 Kafka 原始头转换为 map 结构。
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