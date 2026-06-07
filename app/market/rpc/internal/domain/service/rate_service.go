package service

import (
	"context"
	"strconv"
	"strings"

	goredis "github.com/go-redis/redis/v8"
)

// usdtCNYRateCacheKey 是 USDT 对 CNY 汇率的 Redis 缓存键。
// 外部同步任务会将最新汇率写入此键。
const (
	usdtCNYRateCacheKey = "USDT::CNY::RATE"
)

// rateCache 定义汇率缓存接口。
//
// 使用接口而非具体类型，便于单元测试时 mock。
// 实际实现为 redisx.Client。
type rateCache interface {
	GetCtx(ctx context.Context, key string, value any) error
}

// RateService 持有汇率查询规则。
//
// 业务职责：
//   - 提供 USD 对各法币的汇率查询
//   - 优先从缓存读取实时汇率
//   - 缓存不可用时使用回退值
//
// 为什么保留回退值：
//   - 公开汇率接口应在异步同步任务暂时滞后时优雅降级
//   - 启动顺序不应导致 market-rpc 不可用
//   - 目前只有动态同步的货币需要读取缓存
//
// 依赖：
//   - cache：Redis 缓存客户端
//   - fallbacks：硬编码的回退汇率表
type RateService struct {
	cache     rateCache
	fallbacks map[string]float64
}

// NewRateService 创建 RateService 实例。
//
// 参数：
//   - cache：Redis 缓存客户端（可为 nil，此时使用回退值）
//
// 初始化的回退汇率：
//   - CNY: 6.95
//   - JPY: 136.23
func NewRateService(cache rateCache) *RateService {
	return &RateService{
		cache: cache,
		fallbacks: map[string]float64{
			"CNY": 6.95,
			"JPY": 136.23,
		},
	}
}

// USDRate 获取 USD 对目标法币的汇率。
//
// 业务规则：
//   - unit 参数会被标准化（去空格、转大写）
//   - 对于 CNY，优先从 Redis 缓存读取实时汇率
//   - 缓存未命中或读取失败时使用回退值
//   - 其他法币直接使用回退值
//
// 为什么缓存错误不返回错误：
//   - 汇率查询应始终返回可用值
//   - Redis 暂时不可用不应导致服务不可用
//   - 前端应始终能显示汇率，即使是稍旧的数据
//
// 参数：
//   - ctx：请求上下文
//   - unit：目标法币代码，如 "CNY"、"JPY"
//
// 返回：
//   - float64：USD 对目标法币的汇率（永不为 0，除非法币不支持）
func (s *RateService) USDRate(ctx context.Context, unit string) float64 {
	normalized := strings.ToUpper(strings.TrimSpace(unit))
	if normalized == "CNY" && s.cache != nil {
		var raw string
		if err := s.cache.GetCtx(ctx, usdtCNYRateCacheKey, &raw); err == nil {
			if value, parseErr := strconv.ParseFloat(strings.TrimSpace(raw), 64); parseErr == nil && value > 0 {
				return value
			}
		} else if err != nil && err != goredis.Nil {
			// 缓存错误有意回退，以便即使 Redis 暂时不可用，
			// market-rpc 仍然可读。
		}
	}
	return s.fallbacks[normalized]
}
