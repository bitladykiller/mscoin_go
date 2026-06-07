// Package service 定义钱包领域服务。
//
// WalletService 是钱包管理的领域服务，负责：
//   - 钱包查询：查询会员的所有钱包或指定币种钱包
//   - 钱包创建：为会员创建新币种钱包
//   - 地址管理：更新钱包充值地址
//
// 设计原则：
//   - 接口依赖：通过 walletRepository 接口依赖仓储，便于测试
//   - 延迟创建：钱包在首次访问时创建，不预先创建所有币种钱包
//   - 职责单一：只处理钱包相关的业务逻辑，不涉及余额变动
//
// 与 WithdrawService 的边界：
//   - WalletService：钱包查询、地址管理
//   - WithdrawService：提现申请、余额冻结
package service

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/model"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// walletRepository 钱包仓储接口
// 定义钱包服务所需的仓储操作
//
// 使用接口而非具体实现的好处：
//   - 便于单元测试时 Mock
//   - 解耦服务层与仓储层
//   - 支持不同的仓储实现
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
	repo walletRepository // 钱包仓储
}

// NewWalletService 创建钱包服务实例
// 参数 repo 为钱包仓储接口
func NewWalletService(repo walletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// FindWallet 查询会员所有钱包
// 用于资产页面展示会员持有的所有币种钱包
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - findCoin: 获取币种信息的函数（从 Market RPC 获取）
//
// 返回：
//   - MemberWallet 列表，包含币种信息和余额
//   - error: 错误信息
//
// 注意：findCoin 函数用于丰富钱包的币种信息（汇率、限制等）
func (s *WalletService) FindWallet(ctx context.Context, memberID int64, findCoin func(context.Context, string) (*marketpb.Coin, error)) ([]*assetpb.MemberWallet, error) {
	wallets, err := s.repo.FindByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// 转换为 protobuf 响应
	// 每个钱包需要获取对应的币种信息
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
// 用于特定币种的资产页面或提现页面
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - coinName: 币种名称
//   - coin: 币种信息（从 Market RPC 获取）
//
// 返回：
//   - MemberWallet，包含币种信息和余额
//   - error: 错误信息
func (s *WalletService) FindWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*assetpb.MemberWallet, error) {
	wallet, err := s.EnsureWalletBySymbol(ctx, memberID, coinName, coin)
	if err != nil {
		return nil, err
	}
	return wallet.ToProto(coin), nil
}

// EnsureWalletBySymbol 确保会员拥有指定币种的钱包
// 如果钱包不存在则创建
//
// 使用场景：
//   - 会员首次访问某币种资产页
//   - 会员首次提现某币种
//   - 会员首次充值某币种（需要地址）
//
// 设计原则：
//   - 延迟创建：不预先创建所有币种钱包
//   - 幂等性：多次调用不会创建重复钱包
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - coinName: 币种名称
//   - coin: 币种信息
//
// 返回：
//   - MemberWallet，会员的钱包对象
//   - error: 错误信息
func (s *WalletService) EnsureWalletBySymbol(ctx context.Context, memberID int64, coinName string, coin *marketpb.Coin) (*model.MemberWallet, error) {
	// 先查询是否已存在
	wallet, err := s.repo.FindByMemberIDAndCoinName(ctx, memberID, coinName)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}

	// 创建新钱包
	// 初始余额为 0，地址为空
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
// 用于为会员分配充值地址
//
// 使用场景：
//   - BTC 钱包：从 Bitcoin Core 分配新地址
//   - 其他链：从对应节点分配地址
//
// 参数：
//   - ctx: 请求上下文
//   - wallet: 钱包对象，必须包含新地址
//
// 返回：错误信息
func (s *WalletService) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	return s.repo.UpdateAddress(ctx, wallet)
}

// GetAllAddress 获取指定币种的所有钱包地址
// 用于充值监听服务，获取需要监听的充值地址列表
//
// 参数：
//   - ctx: 请求上下文
//   - coinName: 币种名称
//
// 返回：该币种下所有会员的充值地址列表
func (s *WalletService) GetAllAddress(ctx context.Context, coinName string) ([]string, error) {
	return s.repo.FindAllAddress(ctx, coinName)
}
