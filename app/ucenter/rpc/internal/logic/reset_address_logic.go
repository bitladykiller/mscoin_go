// Package logic 定义地址重置业务逻辑处理器。
package logic

import (
	"context"
	"fmt"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// ResetAddressLogic 重置钱包地址的逻辑处理器
// 处理为会员分配充值地址 RPC 请求
type ResetAddressLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewResetAddressLogic 创建逻辑处理器实例
func NewResetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAddressLogic {
	return &ResetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

// ResetAddress 重置钱包地址
// 为会员分配新的充值地址（主要用于 BTC）
//
// 地址分配流程：
//  1. 从 Market RPC 获取币种信息
//  2. 确保会员拥有该币种的钱包
//  3. 如果是 BTC 且地址为空，从 Bitcoin Core 分配新地址
//  4. 更新钱包地址
//
// 参数：
//   - req: 资产请求，包含用户 ID 和币种名称
//
// 返回：
//   - AssetResp: 空响应
//   - error: 错误信息
//
// 注意：
//   - BTC 地址由 Bitcoin Core 管理，私钥不在 MySQL 中存储
//   - 其他币种的地址分配逻辑待实现
func (l *ResetAddressLogic) ResetAddress(req *assetpb.AssetReq) (*assetpb.AssetResp, error) {
	// 从 Market RPC 获取币种信息
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})
	if err != nil {
		return nil, err
	}

	// 确保会员拥有该币种的钱包
	// 如果钱包不存在会自动创建
	wallet, err := l.svcCtx.WalletService.EnsureWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)
	if err != nil {
		return nil, err
	}

	// 为 BTC 钱包分配充值地址
	// 只有地址为空时才需要分配
	if req.CoinName == "BTC" && wallet.Address == "" {
		// 检查 Bitcoin 地址分配器是否已初始化
		if l.svcCtx.AddressAllocator == nil {
			return nil, fmt.Errorf("bitcoin address allocator is not initialized")
		}

		// 从 Bitcoin Core 分配新地址
		// 使用会员 ID 作为标签，便于追踪
		address, err := l.svcCtx.AddressAllocator.Allocate(l.ctx, fmt.Sprintf("member-%d-btc", req.UserId))
		if err != nil {
			return nil, err
		}

		// 更新钱包地址
		// 该地址现在属于 Bitcoin Core 的钱包，因此没有本地私钥需要持久化到 MySQL
		wallet.Address = address
		wallet.AddressPrivateKey = ""
		if err := l.svcCtx.WalletService.UpdateAddress(l.ctx, wallet); err != nil {
			return nil, err
		}
	}

	return &assetpb.AssetResp{}, nil
}
