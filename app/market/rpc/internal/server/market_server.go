package server

import (
	"context"

	"mscoin_go/app/market/rpc/internal/logic"
	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// MarketServer is the RPC facade for market-domain requests.
type MarketServer struct {
	svcCtx *svc.ServiceContext
	marketpb.UnimplementedMarketServer
}

func NewMarketServer(svcCtx *svc.ServiceContext) *MarketServer {
	return &MarketServer{svcCtx: svcCtx}
}

func (s *MarketServer) FindSymbolThumbTrend(ctx context.Context, req *marketpb.MarketReq) (*marketpb.SymbolThumbRes, error) {
	return logic.NewFindSymbolThumbTrendLogic(ctx, s.svcCtx).FindSymbolThumbTrend(req)
}

func (s *MarketServer) FindSymbolInfo(ctx context.Context, req *marketpb.MarketReq) (*marketpb.ExchangeCoin, error) {
	return logic.NewFindSymbolInfoLogic(ctx, s.svcCtx).FindSymbolInfo(req)
}

func (s *MarketServer) FindCoinInfo(ctx context.Context, req *marketpb.MarketReq) (*marketpb.Coin, error) {
	return logic.NewFindCoinInfoLogic(ctx, s.svcCtx).FindCoinInfo(req)
}

func (s *MarketServer) FindAllCoin(ctx context.Context, req *marketpb.MarketReq) (*marketpb.CoinList, error) {
	return logic.NewFindAllCoinLogic(ctx, s.svcCtx).FindAllCoin(req)
}

func (s *MarketServer) HistoryKline(ctx context.Context, req *marketpb.MarketReq) (*marketpb.HistoryRes, error) {
	return logic.NewHistoryKlineLogic(ctx, s.svcCtx).HistoryKline(req)
}

func (s *MarketServer) FindExchangeCoinVisible(ctx context.Context, req *marketpb.MarketReq) (*marketpb.ExchangeCoinRes, error) {
	return logic.NewFindExchangeCoinVisibleLogic(ctx, s.svcCtx).FindExchangeCoinVisible(req)
}

func (s *MarketServer) FindCoinById(ctx context.Context, req *marketpb.MarketReq) (*marketpb.Coin, error) {
	return logic.NewFindCoinByIDLogic(ctx, s.svcCtx).FindByID(req)
}
