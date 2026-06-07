// Package logic 定义钱包查询业务逻辑处理器。
package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletLogic 查询钱包列表的逻辑处理器
// 处理会员钱包列表查询 RPC 请求
type FindWalletLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewFindWalletLogic 创建逻辑处理器实例
func NewFindWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletLogic {
	return &FindWalletLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWallet 查询会员所有钱包
// 调用 WalletService.FindWallet 处理查询逻辑
//
// 参数：
//   - req: 资产请求，包含用户 ID
//
// 返回：
//   - MemberWalletList: 钱包列表，包含所有币种钱包
//   - error: 错误信息
//
// 注意：每个钱包需要获取对应的币种信息（汇率、限制等），
// 通过 Market RPC 获取
func (l *FindWalletLogic) FindWallet(req *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
	// 调用 WalletService 查询钱包列表
	// 传入获取币种信息的回调函数
	list, err := l.svcCtx.WalletService.FindWallet(l.ctx, req.UserId, func(ctx context.Context, unit string) (*marketpb.Coin, error) {
		// 从 Market RPC 获取币种信息
		return l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: unit})
	})
	if err != nil {
		return nil, err
	}
	return &assetpb.MemberWalletList{List: list}, nil
}
