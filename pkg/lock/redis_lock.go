// Package lock 提供基于 Redis 的分布式锁实现，支持看门狗自动续期机制。
//
// 核心设计：
//   - 基于 Redis SETNX 实现互斥锁
//   - 使用 Lua 脚本保证原子性（比较 + 删除/续期）
//   - 看门狗机制：后台 goroutine 定期续期，防止任务未完成时锁过期
//   - 支持优雅停止：任务完成后停止看门狗并释放锁
//
// 看门狗原理：
//
//	获取锁后启动一个后台 goroutine，每隔 TTL/3 的时间续期一次锁。
//	任务完成后停止看门狗并释放锁。
//	这样可以保证：只要任务还在执行，锁就不会过期。
//
// 使用示例：
//
//	lock, err := lock.NewRedisLock(redisClient, "job:rate-sync", lock.WithTTL(30*time.Second))
//	if err != nil {
//	    return err
//	}
//	defer lock.Close()
//
//	if err := lock.Lock(ctx); err != nil {
//	    // 获取锁失败，其他实例正在执行
//	    return err
//	}
//	defer lock.Unlock(ctx)
//
//	// 执行任务（看门狗会自动续期锁）
//	doWork(ctx)
package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

var (
	// ErrLockNotAcquired 获取锁失败（已被其他实例持有）
	ErrLockNotAcquired = errors.New("lock not acquired")

	// ErrLockReleased 锁已被释放，不能再操作
	ErrLockReleased = errors.New("lock already released")

	// ErrLockLeaseLost 表示看门狗续期失败，当前实例已丢失锁租约。
	ErrLockLeaseLost = errors.New("lock lease lost")
)

// ============================================================================
// 锁配置选项
// ============================================================================

// LockConfig 分布式锁的配置。
type LockConfig struct {
	// TTL 锁的过期时间，默认 30 秒。
	// 注意：TTL 不是锁的持有时间，而是防止死锁的安全兜底。
	// 看门狗会在 TTL/3 的间隔续期，所以锁的实际持有时间可以超过 TTL。
	TTL time.Duration

	// RetryDelay 获取锁失败后的重试间隔，默认 100 毫秒。
	RetryDelay time.Duration

	// MaxRetries 最大重试次数，默认 3 次。设为 0 表示不重试。
	MaxRetries int

	// WatchdogEnabled 是否启用看门狗，默认 true。
	// 禁用后锁到期自动释放，不会续期。
	WatchdogEnabled bool
}

// LockOption 函数式选项模式，用于配置锁。
type LockOption func(*LockConfig)

// WithTTL 设置锁的过期时间。
// 推荐值：任务预估执行时间的 3-5 倍。
func WithTTL(ttl time.Duration) LockOption {
	return func(cfg *LockConfig) {
		cfg.TTL = ttl
	}
}

// WithRetry 设置获取锁的重试策略。
func WithRetry(maxRetries int, delay time.Duration) LockOption {
	return func(cfg *LockConfig) {
		cfg.MaxRetries = maxRetries
		cfg.RetryDelay = delay
	}
}

// WithWatchdog 设置是否启用看门狗。
func WithWatchdog(enabled bool) LockOption {
	return func(cfg *LockConfig) {
		cfg.WatchdogEnabled = enabled
	}
}

// ============================================================================
// Redis 分布式锁
// ============================================================================

// RedisLock 基于 Redis 的分布式锁，支持看门狗自动续期。
//
// 锁的安全性保证：
//   - 互斥性：同一时刻只有一个实例能获取到锁
//   - 无死锁：锁有过期时间，即使持有者崩溃也能自动释放
//   - 看门狗：任务执行期间自动续期，防止任务未完成时锁过期
//   - 原子性：使用 Lua 脚本保证「比较 + 删除/续期」的原子性
type RedisLock struct {
	// rdb Redis 客户端（支持单机和集群模式）
	rdb goredis.UniversalClient

	// key 锁在 Redis 中的键名
	key string

	// value 锁的唯一标识（随机生成，用于安全释放）
	// 为什么需要 value：
	//   - 防止误删其他实例的锁
	//   - 只有持有相同 value 的实例才能释放或续期锁
	value string

	// cfg 锁的配置
	cfg LockConfig

	// watchdog 看门狗定时器
	// 为什么不直接共享 ticker：
	//   - 看门狗 goroutine 会捕获自己的 ticker 和停止上下文副本
	//   - Unlock 只负责触发 cancel 并等待退出，避免共享字段被并发读写
	watchdogCancel context.CancelFunc

	// watchdogWG 等待看门狗 goroutine 退出，确保 Unlock 返回前后台续期已停止。
	watchdogWG sync.WaitGroup

	// runCtx 是持锁任务应使用的执行上下文。
	// 它会在以下任一场景被取消：
	//   - 调用方传入的父 ctx 被取消
	//   - 锁被正常释放
	//   - 看门狗检测到续期失败，租约丢失
	runCtx context.Context

	// runCancel 用于取消 runCtx。
	runCancel context.CancelFunc

	// leaseLost 标记是否已检测到租约丢失。
	leaseLost atomic.Bool

	// released 标记锁是否已释放
	released bool

	// mu 保护 released、runCtx 和 cancel 函数的初始化/读取。
	mu sync.Mutex

	// releaseMu 串行化 Unlock/Close，避免重复释放路径交错执行。
	releaseMu sync.Mutex
}

// NewRedisLock 创建一个新的分布式锁实例。
//
// 参数：
//   - rdb: Redis 客户端（支持 *redis.Client 或 *redis.ClusterClient）
//   - key: 锁的键名，建议使用业务前缀，如 "jobcenter:task:rate-sync"
//   - opts: 配置选项
//
// 返回：
//   - *RedisLock: 锁实例
//   - error: 创建失败时返回错误
func NewRedisLock(rdb goredis.UniversalClient, key string, opts ...LockOption) (*RedisLock, error) {
	if rdb == nil {
		return nil, errors.New("redis client is required")
	}
	if key == "" {
		return nil, errors.New("lock key is required")
	}

	// 默认配置
	cfg := LockConfig{
		TTL:             30 * time.Second,
		RetryDelay:      100 * time.Millisecond,
		MaxRetries:      3,
		WatchdogEnabled: true,
	}

	// 应用选项
	for _, opt := range opts {
		opt(&cfg)
	}

	// 生成随机 value（16 字节 = 32 个十六进制字符）
	// 为什么用随机值：
	//   - 唯一标识当前锁的持有者
	//   - 释放锁时验证身份，防止误删
	value, err := generateRandomValue()
	if err != nil {
		return nil, fmt.Errorf("generate lock value: %w", err)
	}

	return &RedisLock{
		rdb:   rdb,
		key:   key,
		value: value,
		cfg:   cfg,
	}, nil
}

// Lock 获取分布式锁。
//
// 获取流程：
//  1. 使用 SETNX 尝试获取锁，同时设置过期时间（原子操作）
//  2. 如果获取失败，重试 MaxRetries 次
//  3. 获取成功后，如果启用了看门狗，启动后台续期
//
// 错误：
//   - ErrLockNotAcquired: 重试耗尽仍无法获取锁
//   - Redis 连接错误
func (l *RedisLock) Lock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.released {
		return ErrLockReleased
	}

	// 重试获取锁
	var lastErr error
	for i := 0; i <= l.cfg.MaxRetries; i++ {
		ok, err := l.rdb.SetNX(ctx, l.key, l.value, l.cfg.TTL).Result()
		if err != nil {
			lastErr = fmt.Errorf("redis setnx: %w", err)
			time.Sleep(l.cfg.RetryDelay)
			continue
		}

		if ok {
			runCtx, runCancel := context.WithCancel(ctx)
			l.runCtx = runCtx
			l.runCancel = runCancel

			// 获取锁成功，启动看门狗
			if l.cfg.WatchdogEnabled {
				l.startWatchdog(runCtx, runCancel)
			}
			return nil
		}

		// 获取锁失败（已被其他实例持有）
		lastErr = ErrLockNotAcquired
		time.Sleep(l.cfg.RetryDelay)
	}

	return fmt.Errorf("lock acquisition failed after %d retries: %w", l.cfg.MaxRetries, lastErr)
}

// Unlock 释放分布式锁。
//
// 释放流程：
//  1. 停止看门狗（如果正在运行）
//  2. 使用 Lua 脚本原子地「比较 value + 删除 key」
//
// 为什么用 Lua 脚本：
//   - 如果先 GET 再 DELETE，两步之间锁可能已过期被其他实例获取
//   - Lua 脚本保证「比较 + 删除」是原子操作
//
// 注意：Unlock 是幂等的，多次调用不会报错。
func (l *RedisLock) Unlock(ctx context.Context) error {
	l.releaseMu.Lock()
	defer l.releaseMu.Unlock()

	l.mu.Lock()
	if l.released {
		l.mu.Unlock()
		return nil // 已经释放，幂等操作
	}
	l.mu.Unlock()

	// 停止看门狗
	l.stopWatchdog()

	// 使用 Lua 脚本原子释放锁
	// 脚本逻辑：
	//   if redis.call("GET", KEYS[1]) == ARGV[1] then
	//       return redis.call("DEL", KEYS[1])
	//   else
	//       return 0
	//   end
	//
	// 这保证了只有持有相同 value 的实例才能释放锁
	script := goredis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	_, err := script.Run(ctx, l.rdb, []string{l.key}, l.value).Int64()
	if err != nil {
		return fmt.Errorf("release lock: %w", err)
	}

	l.mu.Lock()
	l.released = true
	runCancel := l.runCancel
	l.mu.Unlock()

	if runCancel != nil {
		runCancel()
	}
	return nil
}

// Close 释放锁并清理资源。
// 推荐使用 defer lock.Close() 确保资源释放。
func (l *RedisLock) Close() error {
	return l.Unlock(context.Background())
}

// Key 返回锁的键名。
func (l *RedisLock) Key() string {
	return l.key
}

// Value 返回锁的唯一标识。
func (l *RedisLock) Value() string {
	return l.value
}

// Context 返回持锁任务应使用的执行上下文。
//
// 该上下文在以下场景会被取消：
//   - 调用方传入的父 ctx 被取消
//   - 锁被正常释放
//   - 看门狗检测到续期失败，租约丢失
//
// 如果在成功获取锁之前调用，则返回 Background 上下文作为安全兜底。
func (l *RedisLock) Context() context.Context {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.runCtx != nil {
		return l.runCtx
	}
	return context.Background()
}

// LeaseLost 返回当前实例是否已经丢失锁租约。
func (l *RedisLock) LeaseLost() bool {
	return l.leaseLost.Load()
}

// ============================================================================
// 看门狗机制（核心！）
// ============================================================================

// startWatchdog 启动看门狗续期机制。
//
// 看门狗原理：
//
//	启动一个后台 goroutine，每隔 TTL/3 的时间续期一次锁。
//	这样可以保证：
//	- 只要任务还在执行，锁就不会过期
//	- 任务完成后停止看门狗，锁会在 TTL 后自动过期
//
// 为什么是 TTL/3：
//   - 足够频繁：即使一次续期失败，还有 2 次机会
//   - 不过于频繁：减少 Redis 压力
//   - 例如 TTL=30s，每 10s 续期一次
//
// 续期使用 Lua 脚本保证原子性：
//
//	只有持有相同 value 的实例才能续期锁。
func (l *RedisLock) startWatchdog(runCtx context.Context, runCancel context.CancelFunc) {
	interval := watchdogInterval(l.cfg.TTL)
	ticker := time.NewTicker(interval)
	watchdogCtx, watchdogCancel := context.WithCancel(context.Background())

	l.watchdogCancel = watchdogCancel
	l.watchdogWG.Add(1)

	go func() {
		defer l.watchdogWG.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				renewed, err := l.extendLock(runCtx)
				if err != nil {
					// 父上下文已取消或锁正在释放时，不应误报租约丢失。
					if errors.Is(err, context.Canceled) && runCtx.Err() != nil {
						return
					}
					l.handleLeaseLoss(runCancel)
					return
				}
				if !renewed {
					l.handleLeaseLoss(runCancel)
					return
				}
			case <-watchdogCtx.Done():
				return
			case <-runCtx.Done():
				return
			}
		}
	}()
}

// stopWatchdog 停止看门狗。
// 在释放锁之前调用，确保看门狗不会在锁释放后继续续期。
func (l *RedisLock) stopWatchdog() {
	l.mu.Lock()
	cancel := l.watchdogCancel
	l.mu.Unlock()

	if cancel != nil {
		cancel()
		l.watchdogWG.Wait()
	}
}

// extendLock 续期锁的过期时间。
//
// 使用 Lua 脚本保证原子性：
//
//	if redis.call("GET", KEYS[1]) == ARGV[1] then
//	    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
//	else
//	    return 0
//	end
//
// 只有持有相同 value 的实例才能续期。
// 续期失败不会 panic，但会被看门狗视为租约丢失并取消持锁任务。
func (l *RedisLock) extendLock(ctx context.Context) (bool, error) {
	script := goredis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	ttlMs := int64(l.cfg.TTL / time.Millisecond)
	result, err := script.Run(ctx, l.rdb, []string{l.key}, l.value, ttlMs).Int64()
	if err != nil {
		return false, fmt.Errorf("extend lock: %w", err)
	}
	return result == 1, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// handleLeaseLoss 标记租约丢失并取消持锁任务上下文。
//
// 这里使用原子标志保证：
//   - 续期失败只会触发一次租约丢失语义
//   - 重复的失败信号不会重复取消或污染状态
func (l *RedisLock) handleLeaseLoss(runCancel context.CancelFunc) {
	if !l.leaseLost.CompareAndSwap(false, true) {
		return
	}
	if runCancel != nil {
		runCancel()
	}
}

// watchdogInterval 计算看门狗续期间隔。
//
// 规则：
//   - 使用 TTL/3 作为默认续期间隔
//   - 间隔下限固定为 1 秒，避免极短 TTL 导致过于频繁的 Redis 操作
func watchdogInterval(ttl time.Duration) time.Duration {
	interval := ttl / 3
	if interval < time.Second {
		return time.Second
	}
	return interval
}

// generateRandomValue 生成随机的锁标识。
//
// 使用 crypto/rand 生成 16 字节的随机数据，转换为 32 个十六进制字符。
// 这个值作为锁的唯一标识，用于：
//   - 释放锁时验证身份
//   - 续期锁时验证身份
func generateRandomValue() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
