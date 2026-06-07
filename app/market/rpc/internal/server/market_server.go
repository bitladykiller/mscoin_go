// Package server 包含 gRPC 服务端实现。
//
// Server 层职责：
//   - 实现 gRPC 服务接口（由 protobuf 生成的接口）
//   - 接收 gRPC 请求并路由到对应的 Logic 处理器
//   - 不包含业务逻辑，仅作为传输层适配器
//
// 调用链路：
//
//	gRPC Request -> Server -> Logic -> Domain Service -> Repository -> Database
//
// 本包包含两个 Server：
//   - MarketServer：处理市场数据查询（币种、交易对、K 线）
//   - ExchangeRateServer：处理汇率查询
package server

import (
	"context"

	"mscoin_go/app/market/rpc/internal/logic"
	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// MarketServer 是市场领域请求的 RPC 门面。
//
// 实现了 marketpb.MarketServer 接口，提供以下方法：
//   - FindSymbolThumbTrend：获取所有可见交易对的缩略图和趋势数据
//   - FindSymbolInfo：根据 symbol 获取交易对详情
//   - FindCoinInfo：根据 unit 获取币种详情
//   - FindAllCoin：获取所有币种列表
//   - HistoryKline：获取 K 线历史数据
//   - FindExchangeCoinVisible：获取所有可见的交易对
//   - FindCoinById：根据 ID 获取币种详情
//
// 每个方法都将请求委托给对应的 Logic 处理器，
// Logic 负责编排领域服务完成业务用例。
type MarketServer struct {
	svcCtx *svc.ServiceContext
	marketpb.UnimplementedMarketServer
}

// NewMarketServer 创建 MarketServer 实例。
//
// 参数：
//   - svcCtx：服务上下文，包含所有运行时依赖
func NewMarketServer(svcCtx *svc.ServiceContext) *MarketServer {
	return &MarketServer{svcCtx: svcCtx}
}

// FindSymbolThumbTrend 获取所有可见交易对的缩略图和趋势数据。
//
// 返回每个交易对的当日价格信息、涨跌幅、趋势线等。
// 用于首页市场概览展示。
func (s *MarketServer) FindSymbolThumbTrend(ctx context.Context, req *marketpb.MarketReq) (*marketpb.SymbolThumbRes, error) {
	return logic.NewFindSymbolThumbTrendLogic(ctx, s.svcCtx).FindSymbolThumbTrend(req)
}

// FindSymbolInfo 根据 symbol 获取交易对详情。
//
// symbol 格式为 "BASEQUOTE"，如 "BTCUSDT"。
// 返回交易对的配置信息（精度、费率、限制等）。
func (s *MarketServer) FindSymbolInfo(ctx context.Context, req *marketpb.MarketReq) (*marketpb.ExchangeCoin, error) {
	return logic.NewFindSymbolInfoLogic(ctx, s.svcCtx).FindSymbolInfo(req)
}

// FindCoinInfo 根据 unit 获取币种详情。
//
// unit 为币种单位，如 "BTC"、"ETH"、"USDT"。
// 返回币种的配置信息（充值/提现开关、费率等）。
func (s *MarketServer) FindCoinInfo(ctx context.Context, req *marketpb.MarketReq) (*marketpb.Coin, error) {
	return logic.NewFindCoinInfoLogic(ctx, s.svcCtx).FindCoinInfo(req)
}

// FindAllCoin 获取所有币种列表。
//
// 返回系统中配置的所有币种，用于币种选择器等场景。
func (s *MarketServer) FindAllCoin(ctx context.Context, req *marketpb.MarketReq) (*marketpb.CoinList, error) {
	return logic.NewFindAllCoinLogic(ctx, s.svcCtx).FindAllCoin(req)
}

// HistoryKline 获取 K 线历史数据。
//
// 请求参数：
//   - Symbol：交易对，如 "BTCUSDT"
//   - From/To：时间范围（毫秒时间戳）
//   - Resolution：K 线周期（如 "1H"、"1D"、"15" 等）
//
// 返回指定时间范围内的 K 线数据列表。
func (s *MarketServer) HistoryKline(ctx context.Context, req *marketpb.MarketReq) (*marketpb.HistoryRes, error) {
	return logic.NewHistoryKlineLogic(ctx, s.svcCtx).HistoryKline(req)
}

// FindExchangeCoinVisible 获取所有可见的交易对。
//
// 返回 visible=1 的交易对列表，用于交易对选择器等场景。
func (s *MarketServer) FindExchangeCoinVisible(ctx context.Context, req *marketpb.MarketReq) (*marketpb.ExchangeCoinRes, error) {
	return logic.NewFindExchangeCoinVisibleLogic(ctx, s.svcCtx).FindExchangeCoinVisible(req)
}

// FindCoinById 根据 ID 获取币种详情。
//
// 用于已知币种 ID 的场景，如订单查询后获取币种信息。
func (s *MarketServer) FindCoinById(ctx context.Context, req *marketpb.MarketReq) (*marketpb.Coin, error) {
	return logic.NewFindCoinByIDLogic(ctx, s.svcCtx).FindByID(req)
}
