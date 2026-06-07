// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现行情趋势数据查询的业务逻辑。
package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// SymbolThumbTrendLogic 实现行情趋势数据查询的业务逻辑。
// 该逻辑通过 RPC 调用 market-rpc 服务获取带有趋势数据的行情快照。
//
// 业务流程：
//  1. 接收客户端 IP 参数（可选）
//  2. 调用 MarketClient.FindSymbolThumbTrend RPC 方法
//  3. 将 RPC 响应转换为 API 响应格式
//  4. 返回包含趋势数据的行情快照列表
//
// 与 SymbolThumbLogic 的区别：
//   - SymbolThumbLogic: 只返回基础行情数据，Trend 字段为空
//   - SymbolThumbTrendLogic: 返回基础行情数据 + Trend 数组
//   - Trend 数组包含一段时间内的价格走势点，用于绘制迷你趋势图
//
// 返回信息包括：
//   - 所有交易对的开盘价、收盘价、最高价、最低价
//   - 涨跌幅、涨跌额
//   - 成交量、成交额
//   - USDT 汇率
//   - Trend 数组（价格走势点）
type SymbolThumbTrendLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewSymbolThumbTrendLogic 创建 SymbolThumbTrendLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *SymbolThumbTrendLogic: 初始化完成的业务逻辑实例
func NewSymbolThumbTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolThumbTrendLogic {
	return &SymbolThumbTrendLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// SymbolThumbTrend 执行行情趋势数据查询的业务逻辑。
// 该方法通过 RPC 调用获取带有趋势数据的行情快照。
//
// 参数：
//   - req: 市场请求参数，主要使用 IP 字段进行地理位置相关处理
//
// 返回：
//   - []*types.CoinThumbResp: 交易对行情快照列表（包含 Trend 数据）
//   - error: 错误信息，RPC 调用失败或数据转换失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（10秒超时）
//  2. 调用 MarketClient.FindSymbolThumbTrend RPC 方法
//  3. 使用 copier 将 RPC 响应列表转换为 API 响应格式
//  4. 返回转换后的结果
//
// 数据转换说明：
//   - RPC 返回的是 protobuf 消息数组
//   - 需要转换为 types.CoinThumbResp 数组
//   - Trend 字段会自动被 copier 转换
//
// Trend 数据格式：
//   - Trend 是 float64 数组，包含一段时间内的价格点
//   - 前端可以使用这些数据绘制迷你趋势图（Sparkline）
//   - 通常包含最近 24 小时内的价格采样点
//
// 超时说明：
//   - 设置 10 秒超时，行情数据可能涉及多个交易对的聚合查询
//   - Trend 数据会增加响应体积，但仍使用相同的超时时间
func (l *SymbolThumbTrendLogic) SymbolThumbTrend(req *types.MarketReq) ([]*types.CoinThumbResp, error) {
	// 创建带超时的子上下文，超时时间 10 秒
	// 行情数据可能涉及多个交易对的聚合查询
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取行情趋势数据
	// 传递客户端 IP 用于可能的地理位置优化
	payload, err := l.svcCtx.MarketClient.FindSymbolThumbTrend(ctx, &marketpb.MarketReq{Ip: req.IP})
	if err != nil {
		// RPC 调用失败，返回错误
		return nil, err
	}

	// 将 RPC 响应列表转换为 API 响应格式
	// 预分配切片容量以提高性能
	list := make([]*types.CoinThumbResp, len(payload.List))

	// 使用 copier 进行批量结构体转换
	// 注意：需要传递切片指针以支持切片类型的转换
	if err := copier.Copy(&list, payload.List); err != nil {
		return nil, err
	}

	return list, nil
}