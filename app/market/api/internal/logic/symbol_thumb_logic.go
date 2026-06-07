// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现行情缩略图数据查询的业务逻辑。
package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// SymbolThumbLogic 实现行情缩略图数据查询的业务逻辑。
// 该逻辑通过 RPC 调用 market-rpc 服务获取所有交易对的行情快照。
//
// 业务流程：
//  1. 接收客户端 IP 参数（可选）
//  2. 调用 MarketClient.FindSymbolThumbTrend RPC 方法
//  3. 将 RPC 响应转换为 API 响应格式
//  4. 返回行情快照列表
//
// 返回信息包括：
//   - 所有交易对的开盘价、收盘价、最高价、最低价
//   - 涨跌幅、涨跌额
//   - 成交量、成交额
//   - USDT 汇率
//
// 注意事项：
//   - 该接口返回的 Trend 字段为空，如需趋势数据请使用 SymbolThumbTrendLogic
//   - IP 参数可用于地理位置相关的数据筛选
type SymbolThumbLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewSymbolThumbLogic 创建 SymbolThumbLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *SymbolThumbLogic: 初始化完成的业务逻辑实例
func NewSymbolThumbLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolThumbLogic {
	return &SymbolThumbLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// SymbolThumb 执行行情缩略图数据查询的业务逻辑。
// 该方法通过 RPC 调用获取所有交易对的行情快照。
//
// 参数：
//   - req: 市场请求参数，主要使用 IP 字段进行地理位置相关处理
//
// 返回：
//   - []*types.CoinThumbResp: 交易对行情快照列表
//   - error: 错误信息，RPC 调用失败或数据转换失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（10秒超时）
//  2. 调用 MarketClient.FindSymbolThumbTrend RPC 方法
//     注意：虽然方法名包含 Trend，但此处不使用趋势数据
//  3. 使用 copier 将 RPC 响应列表转换为 API 响应格式
//  4. 返回转换后的结果
//
// 数据转换说明：
//   - RPC 返回的是 protobuf 消息数组
//   - 需要转换为 types.CoinThumbResp 数组
//   - copier 支持切片类型的批量转换
//
// 超时说明：
//   - 设置 10 秒超时，行情数据可能涉及多个交易对
//   - 超时后上下文取消，RPC 调用会被中断
func (l *SymbolThumbLogic) SymbolThumb(req *types.MarketReq) ([]*types.CoinThumbResp, error) {
	// 创建带超时的子上下文，超时时间 10 秒
	// 行情数据可能涉及多个交易对的聚合查询
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取行情缩略图数据
	// 虽然方法名是 FindSymbolThumbTrend，但 Trend 字段在此场景不使用
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