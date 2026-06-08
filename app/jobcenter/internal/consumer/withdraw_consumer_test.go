// Package consumer 提供消费者模块的单元测试。
//
// 测试覆盖：
//   - 错误分类：验证 classifyWithdrawError 的三种处理路径
//   - 边界条件：验证 nil、可重试、不可重试错误的正确分类
package consumer

import (
	"context"
	"errors"
	"testing"

	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/app/jobcenter/internal/model"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/lock"
	"mscoin_go/pkg/mq/kafka"

	"github.com/alicebob/miniredis/v2"
)

type fakeWithdrawProcessor struct {
	processAppliedFn func(ctx context.Context, event *model.WithdrawRecordEvent) error
}

func (f *fakeWithdrawProcessor) ProcessApplied(ctx context.Context, event *model.WithdrawRecordEvent) error {
	return f.processAppliedFn(ctx, event)
}

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

func TestProcessWithdrawEventCallsProcessorWithoutRedis(t *testing.T) {
	t.Parallel()

	called := false
	err := processWithdrawEvent(context.Background(), nil, &fakeWithdrawProcessor{
		processAppliedFn: func(ctx context.Context, event *model.WithdrawRecordEvent) error {
			called = true
			if event.Id != 9 {
				t.Fatalf("event.Id = %d, want 9", event.Id)
			}
			return nil
		},
	}, &model.WithdrawRecordEvent{Id: 9})
	if err != nil {
		t.Fatalf("processWithdrawEvent() error = %v", err)
	}
	if !called {
		t.Fatal("processWithdrawEvent() did not call processor")
	}
}

func TestProcessWithdrawEventFailsWhenCompetingLockExists(t *testing.T) {
	t.Parallel()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer miniRedis.Close()

	cache := redisx.New(redisx.Config{Addrs: []string{miniRedis.Addr()}})

	competingLock, err := lock.NewRedisLock(
		cache.Raw(),
		withdrawEventLockKeyPrefix+"12",
		lock.WithTTL(withdrawEventLockTTL),
		lock.WithRetry(0, 0),
		lock.WithWatchdog(true),
	)
	if err != nil {
		t.Fatalf("NewRedisLock() error = %v", err)
	}
	defer competingLock.Close()

	if err := competingLock.Lock(context.Background()); err != nil {
		t.Fatalf("competingLock.Lock() error = %v", err)
	}

	called := false
	err = processWithdrawEvent(context.Background(), cache, &fakeWithdrawProcessor{
		processAppliedFn: func(ctx context.Context, event *model.WithdrawRecordEvent) error {
			called = true
			return nil
		},
	}, &model.WithdrawRecordEvent{Id: 12})
	if err == nil {
		t.Fatal("processWithdrawEvent() should fail when lock is held by another worker")
	}
	if !errors.Is(err, lock.ErrLockNotAcquired) {
		t.Fatalf("processWithdrawEvent() error = %v, want ErrLockNotAcquired", err)
	}
	if called {
		t.Fatal("processWithdrawEvent() should not call processor when lock acquisition fails")
	}
}
