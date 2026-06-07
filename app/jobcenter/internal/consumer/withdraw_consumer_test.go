// Package consumer 提供消费者模块的单元测试。
//
// 测试覆盖：
//   - 错误分类：验证 classifyWithdrawError 的三种处理路径
//   - 边界条件：验证 nil、可重试、不可重试错误的正确分类
package consumer

import (
	"errors"
	"testing"

	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/pkg/mq/kafka"
)

// TestClassifyWithdrawError 验证错误分类函数的三种处理路径。
//
// 测试用例：
//   - nil 错误：返回 ConsumeAck，确认消息处理成功
//   - 普通错误：返回 ConsumeRetry，触发消息重试
//   - NonRetryableError：返回 ConsumeDeadLetter，发送到死信队列
//
// 该测试确保消费者能正确处理各种错误类型，避免：
//   - 毒消息无限重试阻塞消费者
//   - 可恢复错误被错误发送到死信队列
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
