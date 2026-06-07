package consumer

import (
	"errors"
	"testing"

	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/pkg/mq/kafka"
)

func TestClassifyWithdrawError(t *testing.T) {
	t.Parallel()

	if action := classifyWithdrawError(nil); action != kafka.ConsumeAck {
		t.Fatalf("classifyWithdrawError(nil) = %v, want %v", action, kafka.ConsumeAck)
	}
	if action := classifyWithdrawError(errors.New("temporary")); action != kafka.ConsumeRetry {
		t.Fatalf("classifyWithdrawError(retryable) = %v, want %v", action, kafka.ConsumeRetry)
	}
	if action := classifyWithdrawError(domainservice.NewNonRetryableError(errors.New("poison"))); action != kafka.ConsumeDeadLetter {
		t.Fatalf("classifyWithdrawError(non-retryable) = %v, want %v", action, kafka.ConsumeDeadLetter)
	}
}
