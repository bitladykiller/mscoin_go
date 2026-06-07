package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/svc"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

// UsdRateLogic 处理 UsdRate 请求。
//
// 业务用例：获取 USD 对目标法币的汇率
//   - 查询 USD 对指定法币的实时汇率
//   - 用于法币计价显示、资产估值等场景
//
// 调用链路：
//   ExchangeRateServer -> UsdRateLogic -> RateService.USDRate
type UsdRateLogic struct {
	marketLogicBase
}

// NewUsdRateLogic 创建 UsdRateLogic 实例。
func NewUsdRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UsdRateLogic {
	return &UsdRateLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// UsdRate 执行汇率查询的业务逻辑。
//
// 请求参数：
//   - req.Unit：目标法币代码，如 "CNY"、"JPY"
//
// 返回数据：
//   - Rate：USD 对目标法币的汇率
//
// 数据来源优先级：
//  1. Redis 缓存（由外部同步任务更新）
//  2. 硬编码的回退值（服务可用性保障）
//
// 例如：
//   - Unit="CNY" 返回约 7.0，表示 1 USD ≈ 7.0 CNY
//   - Unit="JPY" 返回约 136.0，表示 1 USD ≈ 136.0 JPY
//
// 注意：本方法不会返回错误，即使缓存不可用也会返回回退值，
// 确保汇率查询始终可用。
func (l *UsdRateLogic) UsdRate(req *ratepb.RateReq) (*ratepb.RateRes, error) {
	return &ratepb.RateRes{Rate: l.rateService.USDRate(l.ctx, req.Unit)}, nil
}
