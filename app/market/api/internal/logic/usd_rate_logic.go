// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现法币汇率查询的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

// UsdRateLogic 实现法币汇率查询的业务逻辑。
// 该逻辑通过 RPC 调用 rate-rpc 服务获取指定法币对 USD 的汇率。
//
// 业务流程：
//  1. 接收法币单位参数（如 "CNY", "EUR", "JPY"）
//  2. 调用 RateClient.UsdRate RPC 方法
//  3. 返回汇率值
//
// 支持的法币单位：
//   - CNY: 人民币
//   - EUR: 欧元
//   - JPY: 日元
//   - KRW: 韩元
//   - 其他主流法币
//
// 使用场景：
//   - 法币资产价值换算
//   - 多币种价格展示
//   - 汇率实时更新
type UsdRateLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewUsdRateLogic 创建 UsdRateLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *UsdRateLogic: 初始化完成的业务逻辑实例
func NewUsdRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UsdRateLogic {
	return &UsdRateLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// UsdRate 执行法币汇率查询的业务逻辑。
// 该方法通过 RPC 调用获取指定法币对 USD 的实时汇率。
//
// 参数：
//   - req: 汇率请求参数，包含以下字段：
//   - Unit: 法币单位代码，如 "CNY", "EUR", "JPY"
//   - IP: 客户端 IP 地址，用于基于地理位置的汇率选择
//
// 返回：
//   - *types.RateResponse: 汇率响应，包含 Rate 字段
//   - error: 错误信息，RPC 调用失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（5秒超时）
//  2. 调用 RateClient.UsdRate RPC 方法
//  3. 将 RPC 响应转换为 API 响应格式
//  4. 返回汇率值
//
// 汇率说明：
//   - Rate 表示 1 USD 可兑换的法币数量
//   - 例如 CNY 的 Rate 为 7.24，表示 1 USD = 7.24 CNY
//   - 汇率数据由后端 rate-rpc 服务提供
//
// 超时说明：
//   - 设置 5 秒超时，汇率查询是轻量级操作
//   - 超时后上下文取消，RPC 调用会被中断
//
// IP 参数用途：
//   - 可用于基于用户地理位置自动选择法币
//   - 例如中国用户默认显示 CNY 汇率
//   - 具体逻辑由后端 rate-rpc 服务实现
func (l *UsdRateLogic) UsdRate(req *types.RateRequest) (*types.RateResponse, error) {
	// 创建带超时的子上下文，超时时间 5 秒
	// 汇率查询是轻量级操作，超时时间较短
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取汇率数据
	// RateClient.UsdRate 根据法币单位返回对 USD 的汇率
	payload, err := l.svcCtx.RateClient.UsdRate(ctx, &ratepb.RateReq{
		Unit: req.Unit, // 法币单位代码
		Ip:   req.IP,   // 客户端 IP，用于地理位置优化
	})
	if err != nil {
		// RPC 调用失败，返回错误
		return nil, err
	}

	// 构造并返回汇率响应
	// Rate 表示 1 USD 可兑换的法币数量
	return &types.RateResponse{Rate: payload.Rate}, nil
}