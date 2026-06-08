// Package consumer 提供 jobcenter 的 Kafka 消费者实现。
//
// 当前实现：
//   - WithdrawConsumer: 提现事件消费者，处理用户提现申请并执行链上转账
//
// 设计原则：
//   - 消费者仅负责消息适配和错误分类
//   - 业务逻辑全部在领域服务中实现
//   - 便于不依赖 Kafka 进行业务逻辑测试
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/app/jobcenter/internal/model"
	"mscoin_go/app/jobcenter/internal/svc"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/lock"
	"mscoin_go/pkg/mq/kafka"

	coreservice "github.com/zeromicro/go-zero/core/service"
)

const (
	// withdrawEventLockKeyPrefix 是提现消费分布式锁的键前缀。
	// 完整键格式：jobcenter:withdraw:{recordID}
	withdrawEventLockKeyPrefix = "jobcenter:withdraw:"

	// withdrawEventLockTTL 是提现消费分布式锁的 TTL。
	// 链上提现可能比普通数据库操作更慢，因此保守使用 2 分钟并交给看门狗自动续期。
	withdrawEventLockTTL = 2 * time.Minute
)

// withdrawProcessor 抽象提现事件的处理能力，便于单元测试锁包装逻辑。
type withdrawProcessor interface {
	ProcessApplied(ctx context.Context, event *model.WithdrawRecordEvent) error
}

// NewWithdrawConsumer 创建提现事件消费者。
//
// 该消费者负责处理由 ucenter-rpc 发出的提现申请事件。
//
// 消息处理流程：
//  1. 从 Kafka 接收提现事件消息
//  2. 反序列化为 WithdrawRecordEvent 结构
//  3. 调用 WithdrawService.ProcessApplied 执行业务逻辑
//  4. 根据返回的错误类型决定后续处理：
//     - nil: 确认消息（Ack）
//     - NonRetryableError: 发送到死信队列（DeadLetter）
//     - 其他错误: 重试（Retry）
//
// 错误处理策略：
//   - 反序列化失败：NonRetryableError，避免无限重试毒消息
//   - 业务逻辑错误：由领域服务决定是否可重试
//   - 临时性错误（如数据库连接失败）：自动重试
//
// 参数：
//   - svcCtx: 服务上下文，提供 WithdrawService 和配置
//
// 返回：
//   - coreservice.Service: 可加入 ServiceGroup 的服务实例
//   - error: 创建失败时的错误
func NewWithdrawConsumer(svcCtx *svc.ServiceContext) (coreservice.Service, error) {
	return kafka.NewConsumerService(
		svcCtx.Config.Kafka,
		func(ctx context.Context, message kafka.Message) error {
			var event model.WithdrawRecordEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				return domainservice.NewNonRetryableError(fmt.Errorf("unmarshal withdraw event: %w", err))
			}
			return processWithdrawEvent(ctx, svcCtx.Cache, svcCtx.WithdrawService, &event)
		},
		classifyWithdrawError,
	)
}

// processWithdrawEvent 在处理提现事件前先获取按记录 ID 粒度的分布式锁。
//
// 为什么这里需要锁：
//   - Kafka 是至少一次投递语义，重复消费在工程上必须视为常态
//   - 提现处理会触发链上转账，属于不可逆的外部副作用
//   - 仅依赖最终状态更新不足以防止“两个实例同时广播转账”
//
// 为什么锁范围放在消费者适配层：
//   - 当前只有 Kafka 消费者会驱动 `ProcessApplied`
//   - 领域服务保持业务聚合职责，不直接感知消息队列和锁键策略
func processWithdrawEvent(ctx context.Context, cache *redisx.Client, processor withdrawProcessor, event *model.WithdrawRecordEvent) error {
	if processor == nil {
		return fmt.Errorf("withdraw processor is required")
	}
	if cache == nil {
		return processor.ProcessApplied(ctx, event)
	}

	eventLock, err := lock.NewRedisLock(
		cache.Raw(),
		fmt.Sprintf("%s%d", withdrawEventLockKeyPrefix, event.Id),
		lock.WithTTL(withdrawEventLockTTL),
		lock.WithRetry(0, 0),
		lock.WithWatchdog(true),
	)
	if err != nil {
		return fmt.Errorf("create withdraw event lock: %w", err)
	}
	defer eventLock.Close()

	if err := eventLock.Lock(ctx); err != nil {
		return fmt.Errorf("acquire withdraw event lock: %w", err)
	}

	return processor.ProcessApplied(ctx, event)
}

// classifyWithdrawError 根据错误类型决定 Kafka 消息的处理动作。
//
// 错误分类策略：
//   - nil: ConsumeAck，确认消息已成功处理
//   - NonRetryableError: ConsumeDeadLetter，发送到死信队列
//   - 其他错误: ConsumeRetry，触发消息重试
//
// 设计原因：
//   - 毒消息（如格式错误）无限重试会阻塞消费者
//   - 业务上不可恢复的错误（如不支持币种）无需重试
//   - 临时性错误（如网络问题）通过重试机制自动恢复
//
// 参数：
//   - err: 消息处理返回的错误
//
// 返回：
//   - kafka.ConsumeAction: 消息后续处理动作
func classifyWithdrawError(err error) kafka.ConsumeAction {
	if err == nil {
		return kafka.ConsumeAck
	}
	if domainservice.IsNonRetryable(err) {
		return kafka.ConsumeDeadLetter
	}
	return kafka.ConsumeRetry
}
