// Package task 提供 jobcenter 的定时任务调度服务。
//
// 该包实现了一个轻量级的定时任务框架，使用 Go 原生的 goroutine 和 time.Ticker，
// 而非引入额外的调度框架（如 cron）。这种设计适合：
//   - 长生命周期的间隔任务
//   - 任务之间相互独立，无复杂依赖关系
//   - 需要精细控制任务执行行为
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
	// defaultIntervalSeconds 是任务执行间隔的默认值（60秒）。
	// 当配置中的 IntervalSeconds <= 0 时使用此默认值。
	defaultIntervalSeconds = 60
)

// jobRunner 定义任务执行函数的签名。
// 返回 error 表示任务执行失败，nil 表示成功。
type jobRunner func(ctx context.Context) error

// intervalJob 封装一个周期性执行的任务。
type intervalJob struct {
	// name 任务名称，用于日志记录和问题排查
	name string

	// schedule 任务调度配置，控制启用状态、启动行为和执行间隔
	schedule config.ScheduleConfig

	// run 任务执行函数，在每次调度周期到达时被调用
	run jobRunner
}

// Service 管理 jobcenter 中所有基于 goroutine 的周期性任务。
//
// 该服务使用 Go 原生 goroutine 和 time.Ticker 实现，因为项目只需要
// 长生命周期的间隔任务，不需要引入额外的调度框架（如 cron）。
//
// 调度逻辑：
//   - 每个任务在独立的 goroutine 中运行
//   - 使用 time.Ticker 控制执行间隔
//   - 支持 RunOnStart 配置，在服务启动时立即执行一次
//   - 任务执行失败仅记录日志，不影响其他任务
//   - 服务停止时等待所有任务优雅退出
//
// 生命周期：
//   - Start() 阻塞运行，直到收到停止信号
//   - Stop() 发送停止信号并等待所有任务完成
type Service struct {
	// ctx 服务上下文，用于通知所有任务停止
	ctx context.Context

	// cancel 取消函数，用于发送停止信号
	cancel context.CancelFunc

	// waiter 等待组，用于等待所有任务完成
	waiter sync.WaitGroup

	// jobs 注册的任务列表
	jobs []intervalJob
}

// NewService 创建定时任务服务实例。
//
// 参数：
//   - svcCtx: 服务上下文，提供配置和领域服务依赖
//
// 返回：go-zero 的 Service 接口，可加入 ServiceGroup 统一管理
//
// 注册的任务：
//   - rate-sync: USD/CNY 汇率同步任务
//   - kline-sync-{period}: K 线数据同步任务（按周期配置多个）
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

// Start 启动所有已注册且已启用的定时任务。
//
// 启动流程：
//  1. 遍历所有任务，跳过未启用的任务
//  2. 为每个启用的任务启动独立 goroutine
//  3. 等待服务上下文的 Done 信号
//  4. 收到停止信号后等待所有任务完成退出
//
// 该方法会阻塞，直到 ctx 被取消。
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

// Stop 发送停止信号并等待所有任务完成。
//
// 调用 cancel() 会触发所有任务的 ctx.Done()，
// 任务会在当前执行周期结束后退出。
func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// registerJobs 从 ServiceContext 注册所有定时任务。
//
// 注册的任务：
//   - rate-sync: 汇率同步，从 OKX 获取 USDT/CNY 实时汇率
//   - kline-sync-{period}: K 线同步，按配置的周期同步市场数据
//
// 注意：K 线任务支持多周期配置，会为每个周期创建独立任务。
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

// runLoop 是单个任务的主循环。
//
// 调度逻辑：
//  1. 计算执行间隔（使用默认值兜底）
//  2. 创建 Ticker 定时触发执行
//  3. 如果配置了 RunOnStart，立即执行一次
//  4. 循环等待 Ticker 或停止信号
//  5. 收到停止信号后退出 goroutine
//
// 参数：
//   - job: 要执行的任务配置
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

// execute 执行单个任务并记录结果。
//
// 执行结果：
//   - 成功：记录 info 日志
//   - 失败：记录 error 日志，但不影响其他任务
//
// 注意：任务执行失败不会导致任务停止，下一个周期会继续尝试执行。
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

// defaultPositive 返回正整数值，如果输入非正则返回默认值。
//
// 参数：
//   - value: 输入值
//   - fallback: 默认值
//
// 返回：如果 value > 0 返回 value，否则返回 fallback
func defaultPositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
