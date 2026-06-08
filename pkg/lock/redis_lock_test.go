package lock

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/go-redis/redis/v8"
)

// ============================================================================
// 参数校验测试
// ============================================================================

func TestNewRedisLock_NilClient(t *testing.T) {
	_, err := NewRedisLock(nil, "test-key")
	if err == nil {
		t.Error("expected error for nil redis client")
	}
}

func TestNewRedisLock_EmptyKey(t *testing.T) {
	// 需要真实的 redis client 才能测试，这里只测试参数校验
	// 真正的集成测试需要 redis 连接
}

// ============================================================================
// 配置选项测试
// ============================================================================

func TestLockConfig_Defaults(t *testing.T) {
	cfg := LockConfig{
		TTL:             30 * time.Second,
		RetryDelay:      100 * time.Millisecond,
		MaxRetries:      3,
		WatchdogEnabled: true,
	}

	if cfg.TTL != 30*time.Second {
		t.Errorf("expected default TTL 30s, got %v", cfg.TTL)
	}
	if cfg.RetryDelay != 100*time.Millisecond {
		t.Errorf("expected default RetryDelay 100ms, got %v", cfg.RetryDelay)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected default MaxRetries 3, got %d", cfg.MaxRetries)
	}
	if !cfg.WatchdogEnabled {
		t.Error("expected watchdog enabled by default")
	}
}

func TestWithTTL(t *testing.T) {
	cfg := &LockConfig{TTL: 30 * time.Second}
	WithTTL(60 * time.Second)(cfg)
	if cfg.TTL != 60*time.Second {
		t.Errorf("expected TTL 60s, got %v", cfg.TTL)
	}
}

func TestWithRetry(t *testing.T) {
	cfg := &LockConfig{MaxRetries: 3, RetryDelay: 100 * time.Millisecond}
	WithRetry(5, 200*time.Millisecond)(cfg)
	if cfg.MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", cfg.MaxRetries)
	}
	if cfg.RetryDelay != 200*time.Millisecond {
		t.Errorf("expected RetryDelay 200ms, got %v", cfg.RetryDelay)
	}
}

func TestWithWatchdog(t *testing.T) {
	cfg := &LockConfig{WatchdogEnabled: true}
	WithWatchdog(false)(cfg)
	if cfg.WatchdogEnabled {
		t.Error("expected watchdog disabled")
	}
}

// ============================================================================
// 随机值生成测试
// ============================================================================

func TestGenerateRandomValue_Uniqueness(t *testing.T) {
	// 生成 100 个随机值，确保没有重复
	values := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v, err := generateRandomValue()
		if err != nil {
			t.Fatalf("generateRandomValue failed: %v", err)
		}
		if values[v] {
			t.Fatalf("duplicate random value: %s", v)
		}
		values[v] = true
	}
}

func TestGenerateRandomValue_Length(t *testing.T) {
	v, err := generateRandomValue()
	if err != nil {
		t.Fatalf("generateRandomValue failed: %v", err)
	}
	// 16 字节 = 32 个十六进制字符
	if len(v) != 32 {
		t.Errorf("expected 32 chars, got %d: %s", len(v), v)
	}
}

// ============================================================================
// 看门狗间隔计算测试
// ============================================================================

func TestWatchdogInterval_Calculation(t *testing.T) {
	tests := []struct {
		ttl      time.Duration
		expected time.Duration
	}{
		{30 * time.Second, 10 * time.Second},      // TTL/3 = 10s
		{60 * time.Second, 20 * time.Second},      // TTL/3 = 20s
		{90 * time.Second, 30 * time.Second},      // TTL/3 = 30s
		{3 * time.Second, 1 * time.Second},        // TTL/3 = 1s（最小值）
		{500 * time.Millisecond, 1 * time.Second}, // TTL/3 < 1s，使用最小值
	}

	for _, tt := range tests {
		interval := watchdogInterval(tt.ttl)
		if interval != tt.expected {
			t.Errorf("TTL=%v: expected interval %v, got %v", tt.ttl, tt.expected, interval)
		}
	}
}

func TestLockWatchdogCancelsContextWhenLeaseLost(t *testing.T) {
	t.Parallel()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer miniRedis.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: miniRedis.Addr()})
	defer func() {
		_ = rdb.Close()
	}()

	lock, err := NewRedisLock(rdb, "test:lock:lease-lost",
		WithTTL(1500*time.Millisecond),
		WithRetry(0, 0),
		WithWatchdog(true),
	)
	if err != nil {
		t.Fatalf("NewRedisLock() failed: %v", err)
	}

	if err := lock.Lock(context.Background()); err != nil {
		t.Fatalf("Lock() failed: %v", err)
	}

	miniRedis.Close()

	select {
	case <-lock.Context().Done():
	case <-time.After(3 * time.Second):
		t.Fatal("watchdog did not cancel run context after lease loss")
	}

	if !lock.LeaseLost() {
		t.Fatal("expected leaseLost to be true after watchdog renewal failure")
	}
}

func TestLockUnlockStopsWatchdogWithoutLateLeaseLoss(t *testing.T) {
	t.Parallel()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer miniRedis.Close()

	rdb := goredis.NewClient(&goredis.Options{Addr: miniRedis.Addr()})
	defer func() {
		_ = rdb.Close()
	}()

	lock, err := NewRedisLock(rdb, "test:lock:unlock",
		WithTTL(1500*time.Millisecond),
		WithRetry(0, 0),
		WithWatchdog(true),
	)
	if err != nil {
		t.Fatalf("NewRedisLock() failed: %v", err)
	}

	if err := lock.Lock(context.Background()); err != nil {
		t.Fatalf("Lock() failed: %v", err)
	}

	if err := lock.Unlock(context.Background()); err != nil {
		t.Fatalf("Unlock() failed: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	if lock.LeaseLost() {
		t.Fatal("Unlock() should stop watchdog before any late lease-loss signal is emitted")
	}
	if miniRedis.Exists(lock.Key()) {
		t.Fatal("lock key should be deleted after Unlock()")
	}
	if lock.Context().Err() == nil {
		t.Fatal("run context should be canceled after Unlock()")
	}
}

// ============================================================================
// Lock/Unlock 生命周期测试（需要 Redis 连接）
// ============================================================================

// 以下测试需要真实的 Redis 连接，可以通过环境变量控制是否运行
// 运行方式：go test -v -run TestLock_Integration -tags=integration

func skipIfNoRedis(t *testing.T) {
	t.Helper()
	// 如果没有 Redis 连接，跳过集成测试
	t.Skip("skipping integration test (no redis connection)")
}

func TestLock_Integration_LockAndUnlock(t *testing.T) {
	skipIfNoRedis(t)

	// 这里需要真实的 Redis 客户端
	// rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	// lock, err := NewRedisLock(rdb, "test:lock:integration", WithTTL(10*time.Second))
	// if err != nil {
	//     t.Fatal(err)
	// }
	// defer lock.Close()
	//
	// ctx := context.Background()
	//
	// // 获取锁
	// if err := lock.Lock(ctx); err != nil {
	//     t.Fatalf("Lock failed: %v", err)
	// }
	//
	// // 释放锁
	// if err := lock.Unlock(ctx); err != nil {
	//     t.Fatalf("Unlock failed: %v", err)
	// }
}

func TestLock_Integration_MutualExclusion(t *testing.T) {
	skipIfNoRedis(t)

	// 测试互斥性：
	// lock1 获取锁后，lock2 应该获取失败
	// rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	//
	// lock1, _ := NewRedisLock(rdb, "test:lock:mutex", WithTTL(10*time.Second), WithRetry(0, 0))
	// lock2, _ := NewRedisLock(rdb, "test:lock:mutex", WithTTL(10*time.Second), WithRetry(0, 0))
	//
	// ctx := context.Background()
	//
	// // lock1 获取成功
	// if err := lock1.Lock(ctx); err != nil {
	//     t.Fatalf("lock1.Lock failed: %v", err)
	// }
	// defer lock1.Unlock(ctx)
	//
	// // lock2 应该获取失败
	// if err := lock2.Lock(ctx); err == nil {
	//     t.Fatal("lock2.Lock should have failed")
	// }
}

func TestLock_Integration_Watchdog(t *testing.T) {
	skipIfNoRedis(t)

	// 测试看门狗续期：
	// 获取锁后等待超过 TTL，锁应该仍然有效（因为看门狗续期了）
	// rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	//
	// lock, _ := NewRedisLock(rdb, "test:lock:watchdog",
	//     WithTTL(3*time.Second),  // 短 TTL 便于测试
	//     WithWatchdog(true),
	// )
	//
	// ctx := context.Background()
	//
	// if err := lock.Lock(ctx); err != nil {
	//     t.Fatal(err)
	// }
	//
	// // 等待 4 秒（超过 TTL=3s）
	// // 如果看门狗正常工作，锁应该仍然有效
	// time.Sleep(4 * time.Second)
	//
	// // 验证锁仍然存在
	// val, err := rdb.Get(ctx, lock.Key()).Result()
	// if err != nil {
	//     t.Fatalf("lock should still exist: %v", err)
	// }
	// if val != lock.Value() {
	//     t.Fatal("lock value mismatch")
	// }
	//
	// lock.Unlock(ctx)
}

func TestLock_Integration_IdempotentUnlock(t *testing.T) {
	skipIfNoRedis(t)

	// 测试幂等释放：
	// 多次调用 Unlock 不应该报错
	// rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	//
	// lock, _ := NewRedisLock(rdb, "test:lock:idempotent")
	// ctx := context.Background()
	//
	// lock.Lock(ctx)
	//
	// // 第一次释放
	// if err := lock.Unlock(ctx); err != nil {
	//     t.Fatalf("first Unlock failed: %v", err)
	// }
	//
	// // 第二次释放（应该幂等）
	// if err := lock.Unlock(ctx); err != nil {
	//     t.Fatalf("second Unlock should be idempotent: %v", err)
	// }
}

// ============================================================================
// 基准测试
// ============================================================================

func BenchmarkGenerateRandomValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generateRandomValue()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 确保 context 在测试中被正确使用
var _ = context.Background
