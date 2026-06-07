package server

import (
	"context"

	"mscoin_go/app/market/rpc/internal/logic"
	"mscoin_go/app/market/rpc/internal/svc"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

// ExchangeRateServer 是法币汇率查询的 RPC 门面。
//
// 实现了 ratepb.ExchangeRateServer 接口，提供以下方法：
//   - UsdRate：获取 USD 对目标法币的汇率
//
// 汇率数据来源：
//   - 优先从 Redis 缓存读取（由外部同步任务更新）
//   - 缓存未命中或读取失败时使用硬编码的回退值
//
// 这种设计确保：
//   - 汇率查询始终可用，即使 Redis 暂时不可用
//   - 服务启动不依赖外部汇率数据
type ExchangeRateServer struct {
	svcCtx *svc.ServiceContext
	ratepb.UnimplementedExchangeRateServer
}

// NewExchangeRateServer 创建 ExchangeRateServer 实例。
//
// 参数：
//   - svcCtx：服务上下文，包含 RateService 等依赖
func NewExchangeRateServer(svcCtx *svc.ServiceContext) *ExchangeRateServer {
	return &ExchangeRateServer{svcCtx: svcCtx}
}

// UsdRate 获取 USD 对目标法币的汇率。
//
// 请求参数：
//   - Unit：目标法币代码，如 "CNY"、"JPY"
//
// 返回：
//   - Rate：USD 对目标法币的汇率
//
// 例如 Unit="CNY" 返回约 7.0，表示 1 USD ≈ 7.0 CNY。
func (s *ExchangeRateServer) UsdRate(ctx context.Context, req *ratepb.RateReq) (*ratepb.RateRes, error) {
	return logic.NewUsdRateLogic(ctx, s.svcCtx).UsdRate(req)
}
