package logic

import (
	"context"
	"fmt"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type QueryWithdrawCoinLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryWithdrawCoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryWithdrawCoinLogic {
	return &QueryWithdrawCoinLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryWithdrawCoinLogic) QueryWithdrawCoin() ([]*types.WithdrawWalletInfo, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	coinList, err := l.svcCtx.MarketClient.FindAllCoin(ctx, &marketpb.MarketReq{})
	if err != nil {
		return nil, err
	}

	coinMap := make(map[string]*marketpb.Coin, len(coinList.GetList()))
	for _, coin := range coinList.GetList() {
		coinMap[coin.Unit] = coin
	}

	walletList, err := l.svcCtx.AssetClient.FindWallet(ctx, &assetpb.AssetReq{UserId: userID})
	if err != nil {
		return nil, err
	}

	resp := make([]*types.WithdrawWalletInfo, 0, len(walletList.GetList()))
	for _, wallet := range walletList.GetList() {
		if wallet.GetCoin() == nil {
			return nil, fmt.Errorf("wallet coin info missing")
		}
		coin, ok := coinMap[wallet.GetCoin().GetUnit()]
		if !ok {
			return nil, fmt.Errorf("coin %s not found", wallet.GetCoin().GetUnit())
		}

		addressList, err := l.svcCtx.WithdrawClient.FindAddressByCoinId(ctx, &withdrawpb.WithdrawReq{
			UserId: userID,
			CoinId: int64(coin.Id),
		})
		if err != nil {
			return nil, err
		}

		item := &types.WithdrawWalletInfo{
			Unit:            coin.Unit,
			Threshold:       coin.WithdrawThreshold,
			MinAmount:       coin.MinWithdrawAmount,
			MaxAmount:       coin.MaxWithdrawAmount,
			MinTxFee:        coin.MinTxFee,
			MaxTxFee:        coin.MaxTxFee,
			NameCn:          coin.NameCn,
			Name:            coin.Name,
			Balance:         wallet.Balance,
			CanAutoWithdraw: autoWithdrawString(coin.CanAutoWithdraw),
			WithdrawScale:   coin.WithdrawScale,
			AccountType:     coin.AccountType,
			Addresses:       make([]types.AddressSimple, 0, len(addressList.GetList())),
		}
		for _, address := range addressList.GetList() {
			item.Addresses = append(item.Addresses, types.AddressSimple{
				Remark:  address.Remark,
				Address: address.Address,
			})
		}
		resp = append(resp, item)
	}

	return resp, nil
}

func autoWithdrawString(value int32) string {
	if value == 0 {
		return "true"
	}
	return "false"
}
