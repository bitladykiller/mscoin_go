// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现 K 线历史数据查询的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// HistoryLogic 实现 K 线历史数据查询的业务逻辑。
// 该逻辑通过 RPC 调用 market-rpc 服务获取指定交易对的 K 线数据。
//
// 业务流程：
//  1. 接收交易对代码、时间范围、K 线周期参数
//  2. 调用 MarketClient.HistoryKline RPC 方法
//  3. 将 RPC 响应转换为前端可用的格式
//  4. 返回 K 线数据列表
//
// K 线周期说明：
//   - "1": 1分钟
//   - "5": 5分钟
//   - "15": 15分钟
//   - "30": 30分钟
//   - "60": 1小时
//   - "1D": 1天
//   - "1W": 1周
//   - "1M": 1月
type HistoryLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewHistoryLogic 创建 HistoryLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *HistoryLogic: 初始化完成的业务逻辑实例
func NewHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryLogic {
	return &HistoryLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// History 执行 K 线历史数据查询的业务逻辑。
// 该方法通过 RPC 调用获取指定交易对的历史 K 线数据。
//
// 参数：
//   - req: 市场请求参数，包含以下字段：
//   - Symbol: 交易对代码，如 "BTCUSDT"
//   - From: 起始时间戳（毫秒）
//   - To: 结束时间戳（毫秒）
//   - Resolution: K 线周期，如 "1", "5", "15", "60", "1D"
//
// 返回：
//   - *types.HistoryKline: K 线数据响应，包含 List 数组
//   - error: 错误信息，RPC 调用失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（10秒超时）
//  2. 调用 MarketClient.HistoryKline RPC 方法
//  3. 将 RPC 响应中的每个 K 线点转换为数组格式
//  4. 返回转换后的结果
//
// 数据格式说明：
//   - RPC 响应：每个 K 线点是结构体，包含 Time, Open, High, Low, Close, Volume 字段
//   - API 响应：每个 K 线点是数组 [time, open, high, low, close, volume]
//   - 这种转换是为了保持与前端图表库的兼容性
//
// 超时说明：
//   - 设置 10 秒超时，K 线数据量可能较大
//   - 超时后上下文取消，RPC 调用会被中断
func (l *HistoryLogic) History(req *types.MarketReq) (*types.HistoryKline, error) {
	// 创建带超时的子上下文，超时时间 10 秒
	// K 线数据可能较多，超时时间相对较长
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取 K 线历史数据
	// 传递交易对代码、时间范围和周期参数
	payload, err := l.svcCtx.MarketClient.HistoryKline(ctx, &marketpb.MarketReq{
		Symbol:     req.Symbol,
		From:       req.From,
		To:         req.To,
		Resolution: req.Resolution,
	})
	if err != nil {
		// RPC 调用失败，返回错误
		return nil, err
	}

	// 转换 K 线数据格式
	// RPC 响应中每个 K 线点是结构体，需要转换为数组格式
	// 数组格式: [time, open, high, low, close, volume]
	list := make([][]any, len(payload.List))
	for i, item := range payload.List {
		list[i] = []any{item.Time, item.Open, item.High, item.Low, item.Close, item.Volume}
	}

	return &types.HistoryKline{List: list}, nil
}