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

// NewWithdrawConsumer 创建 jobcenter 中首个迁移的 Kafka 消费者。
//
// 该消费者仅负责将队列消息适配为领域服务调用。所有真正的业务规则
// 保留在领域层，这样相同的工作流无需 Kafka 即可进行测试。
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

func classifyWithdrawError(err error) kafka.ConsumeAction {
	if err == nil {
		return kafka.ConsumeAck
	}
	if domainservice.IsNonRetryable(err) {
		return kafka.ConsumeDeadLetter
	}
	return kafka.ConsumeRetry
}
