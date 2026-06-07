package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// AssetServer 资产 RPC 服务端
// 处理钱包、交易等资产相关的 RPC 请求
type AssetServer struct {
	svcCtx *svc.ServiceContext
	assetpb.UnimplementedAssetServer
}

// NewAssetServer 创建资产服务端实例
func NewAssetServer(svcCtx *svc.ServiceContext) *AssetServer {
	return &AssetServer{svcCtx: svcCtx}
}

// FindWalletBySymbol 根据币种查询会员钱包
func (s *AssetServer) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	return logic.NewFindWalletBySymbolLogic(ctx, s.svcCtx).FindWalletBySymbol(in)
}

// FindWallet 查询会员所有钱包
func (s *AssetServer) FindWallet(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
	return logic.NewFindWalletLogic(ctx, s.svcCtx).FindWallet(in)
}

// ResetAddress 重置钱包地址
func (s *AssetServer) ResetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AssetResp, error) {
	return logic.NewResetAddressLogic(ctx, s.svcCtx).ResetAddress(in)
}

// FindTransaction 查询会员交易记录
func (s *AssetServer) FindTransaction(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberTransactionList, error) {
	return logic.NewFindTransactionLogic(ctx, s.svcCtx).FindTransaction(in)
}

// GetAddress 获取指定币种的所有钱包地址
func (s *AssetServer) GetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AddressList, error) {
	return logic.NewGetAddressLogic(ctx, s.svcCtx).GetAddress(in)
}
