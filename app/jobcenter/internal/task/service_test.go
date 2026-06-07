package task

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"mscoin_go/app/jobcenter/internal/config"
)

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
