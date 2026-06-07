package logic

import (
	"context"
	"fmt"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type ResetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAddressLogic {
	return &ResetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ResetAddressLogic) ResetAddress(req *assetpb.AssetReq) (*assetpb.AssetResp, error) {
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})
	if err != nil {
		return nil, err
	}

	wallet, err := l.svcCtx.WalletService.EnsureWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)
	if err != nil {
		return nil, err
	}

	if req.CoinName == "BTC" && wallet.Address == "" {
		if l.svcCtx.AddressAllocator == nil {
			return nil, fmt.Errorf("bitcoin address allocator is not initialized")
		}

		address, err := l.svcCtx.AddressAllocator.Allocate(l.ctx, fmt.Sprintf("member-%d-btc", req.UserId))
		if err != nil {
			return nil, err
		}

		wallet.Address = address
		// The address now belongs to Bitcoin Core's wallet, so there is no local
		// private key to persist in MySQL.
		wallet.AddressPrivateKey = ""
		if err := l.svcCtx.WalletService.UpdateAddress(l.ctx, wallet); err != nil {
			return nil, err
		}
	}

	return &assetpb.AssetResp{}, nil
}
