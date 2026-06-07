package service

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/model"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type walletRepository interface {
	FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error)
	FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	FindAllAddress(ctx context.Context, coinName string) ([]string, error)
	UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error
	Save(ctx context.Context, wallet *model.MemberWallet) error
}

type WalletService struct {
	repo walletRepository
}

func NewWalletService(repo walletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) FindWallet(ctx context.Context, memberID int64, findCoin func(context.Context, string) (*marketpb.Coin, error)) ([]*assetpb.MemberWallet, error) {
	wallets, err := s.repo.FindByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	resp := make([]*assetpb.MemberWallet, 0, len(wallets))
	for _, wallet := range wallets {
		coin, err := findCoin(ctx, wallet.CoinName)
		if err != nil {
			return nil, err
		}
		resp = append(resp, wallet.ToProto(coin))
	}
	return resp, nil
}

func (s *WalletService) FindWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*assetpb.MemberWallet, error) {
	wallet, err := s.EnsureWalletBySymbol(ctx, memberID, coinName, coin)
	if err != nil {
		return nil, err
	}
	return wallet.ToProto(coin), nil
}

func (s *WalletService) EnsureWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*model.MemberWallet, error) {
	wallet, err := s.repo.FindByMemberIDAndCoinName(ctx, memberID, coinName)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	wallet = &model.MemberWallet{
		MemberId: memberID,
		CoinId:   int64(coin.Id),
		CoinName: coin.Unit,
	}
	if err := s.repo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	return s.repo.UpdateAddress(ctx, wallet)
}

func (s *WalletService) GetAllAddress(ctx context.Context, coinName string) ([]string, error) {
	return s.repo.FindAllAddress(ctx, coinName)
}
