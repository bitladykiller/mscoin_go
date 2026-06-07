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

	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/app/jobcenter/internal/model"
	"mscoin_go/app/jobcenter/internal/svc"
	"mscoin_go/pkg/mq/kafka"

	coreservice "github.com/zeromicro/go-zero/core/service"
)

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
			return svcCtx.WithdrawService.ProcessApplied(ctx, &event)
		},
		classifyWithdrawError,
	)
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
