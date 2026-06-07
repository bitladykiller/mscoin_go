// Package logic 定义地址列表查询业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// GetAddressLogic 获取地址列表的逻辑处理器
// 处理获取指定币种的所有钱包地址 RPC 请求
type GetAddressLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewGetAddressLogic 创建逻辑处理器实例
func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetAddress 获取指定币种的所有钱包地址
// 调用 WalletService.GetAllAddress 处理查询逻辑
//
// 参数：
//   - req: 资产请求，包含币种名称
//
// 返回：
//   - AddressList: 地址列表
//   - error: 错误信息
//
// 使用场景：充值监听服务，获取需要监听的充值地址列表
func (l *GetAddressLogic) GetAddress(req *assetpb.AssetReq) (*assetpb.AddressList, error) {
	// 调用 WalletService 获取所有充值地址
	addresses, err := l.svcCtx.WalletService.GetAllAddress(l.ctx, req.CoinName)
	if err != nil {
		return nil, err
	}
	return &assetpb.AddressList{List: addresses}, nil
}
