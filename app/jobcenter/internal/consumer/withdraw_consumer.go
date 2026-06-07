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

// NewWithdrawConsumer creates the first migrated Kafka worker in jobcenter.
//
// The consumer only adapts queue payloads into domain service calls. All real
// business rules stay in the domain layer so the same workflow remains testable
// without Kafka.
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
