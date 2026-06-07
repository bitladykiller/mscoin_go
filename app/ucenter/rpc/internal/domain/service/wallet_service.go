package service

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/model"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// walletRepository 钱包仓储接口
// 定义钱包服务所需的仓储操作
type walletRepository interface {
	FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error)
	FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error)
	FindAllAddress(ctx context.Context, coinName string) ([]string, error)
	UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error
	Save(ctx context.Context, wallet *model.MemberWallet) error
}

// WalletService 钱包服务
// 负责会员钱包查询、地址管理等业务逻辑
type WalletService struct {
	repo walletRepository
}

// NewWalletService 创建钱包服务实例
func NewWalletService(repo walletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// FindWallet 查询会员所有钱包
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

// FindWalletBySymbol 根据币种查询会员钱包
func (s *WalletService) FindWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*assetpb.MemberWallet, error) {
	wallet, err := s.EnsureWalletBySymbol(ctx, memberID, coinName, coin)
	if err != nil {
		return nil, err
	}
	return wallet.ToProto(coin), nil
}

// EnsureWalletBySymbol 确保会员拥有指定币种的钱包
// 如果钱包不存在则创建
func (s *WalletService) EnsureWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*model.MemberWallet, error) {
	wallet, err := s.repo.FindByMemberIDAndCoinName(ctx, memberID, coinName)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	// 创建新钱包
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

// UpdateAddress 更新钱包地址
func (s *WalletService) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	return s.repo.UpdateAddress(ctx, wallet)
}

// GetAllAddress 获取指定币种的所有钱包地址
func (s *WalletService) GetAllAddress(ctx context.Context, coinName string) ([]string, error) {
	return s.repo.FindAllAddress(ctx, coinName)
}
