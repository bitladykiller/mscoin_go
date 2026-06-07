package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mscoin_go/app/jobcenter/internal/config"
	"mscoin_go/app/jobcenter/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	coreservice "github.com/zeromicro/go-zero/core/service"
)

const (
	defaultIntervalSeconds = 60
)

type jobRunner func(ctx context.Context) error

type intervalJob struct {
	name     string
	schedule config.ScheduleConfig
	run      jobRunner
}

// Service manages all goroutine-based periodic tasks in `jobcenter`.
//
// This service intentionally uses plain goroutines and `time.Ticker` because
// the project only needs long-lived interval jobs, and the user explicitly
// prefers a native Go concurrency model over an extra scheduler framework.
type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
	waiter sync.WaitGroup
	jobs   []intervalJob
}

func NewService(svcCtx *svc.ServiceContext) coreservice.Service {
	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		ctx:    ctx,
		cancel: cancel,
	}

	if svcCtx != nil {
		service.registerJobs(svcCtx)
	}
	return service
}

func (s *Service) Start() {
	for _, job := range s.jobs {
		if !job.schedule.Enabled {
			logx.Infof("jobcenter task %s is disabled", job.name)
			continue
		}

		s.waiter.Add(1)
		go s.runLoop(job)
	}

	<-s.ctx.Done()
	s.waiter.Wait()
}

func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Service) registerJobs(svcCtx *svc.ServiceContext) {
	s.jobs = append(s.jobs, intervalJob{
		name:     "rate-sync",
		schedule: svcCtx.Config.Tasks.RateSync,
		run:      svcCtx.ExchangeRateSyncService.SyncUSDCNY,
	})

	for _, item := range svcCtx.Config.Tasks.Klines {
		cfg := item
		period := cfg.Period
		name := fmt.Sprintf("kline-sync-%s", period)
		s.jobs = append(s.jobs, intervalJob{
			name:     name,
			schedule: cfg.ScheduleConfig,
			run: func(ctx context.Context) error {
				return svcCtx.KlineSyncService.SyncPeriod(ctx, period, cfg.PublishLatest, cfg.PublishTopic)
			},
		})
	}
}

func (s *Service) runLoop(job intervalJob) {
	defer s.waiter.Done()

	interval := time.Duration(defaultPositive(job.schedule.IntervalSeconds, defaultIntervalSeconds)) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logx.Infof("starting jobcenter task %s with interval %s", job.name, interval)

	if job.schedule.RunOnStart {
		s.execute(job)
	}

	for {
		select {
		case <-s.ctx.Done():
			logx.Infof("jobcenter task %s stopped", job.name)
			return
		case <-ticker.C:
			s.execute(job)
		}
	}
}

func (s *Service) execute(job intervalJob) {
	if job.run == nil {
		return
	}
	if err := job.run(s.ctx); err != nil {
		logx.Errorf("jobcenter task %s failed: %v", job.name, err)
		return
	}
	logx.Infof("jobcenter task %s completed successfully", job.name)
}

func defaultPositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
