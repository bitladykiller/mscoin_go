// Package task 提供定时任务服务的单元测试。
//
// 测试覆盖：
//   - RunOnStart 配置：验证服务启动时立即执行任务
//   - 任务执行：验证任务函数被正确调用
//   - 优雅停止：验证服务停止时等待任务完成
package task

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"mscoin_go/app/jobcenter/internal/config"
)

// TestServiceRunsRunOnStartJobImmediately 验证 RunOnStart 配置的功能。
//
// 测试场景：
//   - 配置一个 RunOnStart=true 的任务
//   - 任务间隔设置为 3600 秒（确保不会因间隔触发）
//   - 启动服务后，任务应立即执行一次
//
// 验证点：
//   - 任务的调用计数在 2 秒内达到 1 次
//   - 服务能正常停止
func TestServiceRunsRunOnStartJobImmediately(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int32
	service := &Service{
		ctx:    ctx,
		cancel: cancel,
		jobs: []intervalJob{
			{
				name: "immediate",
				schedule: config.ScheduleConfig{
					Enabled:         true,
					RunOnStart:      true,
					IntervalSeconds: 3600,
				},
				run: func(ctx context.Context) error {
					atomic.AddInt32(&calls, 1)
					return nil
				},
			},
		},
	}

	done := make(chan struct{})
	go func() {
		service.Start()
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&calls) >= 1 {
			service.Stop()
			<-done
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	service.Stop()
	<-done
	t.Fatal("Start() did not run the task immediately")
}
