package logic

import (
	"context"
	"errors"
	"fmt"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

type AddLogic struct {
	exchangeLogicBase
}

func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

func (l *AddLogic) Add(req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	memberInfo, err := l.svcCtx.MemberClient.FindMemberById(l.ctx, &memberpb.MemberReq{MemberId: req.UserId})
	if err != nil {
		return nil, err
	}

	exchangeCoin, err := l.svcCtx.MarketClient.FindSymbolInfo(l.ctx, &marketpb.MarketReq{Symbol: req.Symbol})
	if err != nil || exchangeCoin == nil {
		return nil, errors.New("nonsupport coin")
	}

	baseSymbol := exchangeCoin.GetBaseSymbol()
	coinSymbol := exchangeCoin.GetCoinSymbol()

	baseWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: baseSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find base wallet: %w", err)
	}
	coinWallet, err := l.svcCtx.AssetClient.FindWalletBySymbol(l.ctx, &assetpb.AssetReq{
		UserId:   req.UserId,
		CoinName: coinSymbol,
	})
	if err != nil {
		return nil, fmt.Errorf("find coin wallet: %w", err)
	}

	if err := l.svcCtx.OrderService.ValidateAddRequest(
		req.Direction,
		req.Type,
		req.Price,
		req.Amount,
		memberInfo,
		exchangeCoin,
		baseWallet,
		coinWallet,
	); err != nil {
		return nil, err
	}

	count, err := l.svcCtx.OrderService.FindCurrentTradingCount(l.ctx, req.UserId, req.Symbol, req.Direction)
	if err != nil {
		return nil, err
	}
	if exchangeCoin.GetMaxTradingOrder() > 0 && count >= exchangeCoin.GetMaxTradingOrder() {
		return nil, fmt.Errorf("too many trading orders, max=%d", exchangeCoin.GetMaxTradingOrder())
	}

	order := l.svcCtx.OrderService.BuildNewOrder(
		req.UserId,
		req.Symbol,
		baseSymbol,
		coinSymbol,
		req.Type,
		req.Direction,
		req.Price,
		req.Amount,
	)
	if err := l.svcCtx.OrderService.AddOrder(l.ctx, order); err != nil {
		return nil, err
	}

	return &orderpb.AddOrderRes{OrderId: order.OrderId}, nil
}
