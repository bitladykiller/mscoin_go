// Package server 定义资产 RPC 服务端。
package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// AssetServer 资产 RPC 服务端
// 处理钱包、交易等资产相关的 RPC 请求
//
// 实现 assetpb.AssetServer 接口
// 提供以下方法：
//   - FindWalletBySymbol: 根据币种查询会员钱包
//   - FindWallet: 查询会员所有钱包
//   - ResetAddress: 重置钱包地址
//   - FindTransaction: 查询会员交易记录
//   - GetAddress: 获取指定币种的所有钱包地址
type AssetServer struct {
	svcCtx *svc.ServiceContext       // 服务上下文
	assetpb.UnimplementedAssetServer // 未实现方法的默认实现
}

// NewAssetServer 创建资产服务端实例
func NewAssetServer(svcCtx *svc.ServiceContext) *AssetServer {
	return &AssetServer{svcCtx: svcCtx}
}

// FindWalletBySymbol 根据币种查询会员钱包
// 接收 gRPC 请求并转发给 FindWalletBySymbolLogic 处理
func (s *AssetServer) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	return logic.NewFindWalletBySymbolLogic(ctx, s.svcCtx).FindWalletBySymbol(in)
}

// FindWallet 查询会员所有钱包
// 接收 gRPC 请求并转发给 FindWalletLogic 处理
func (s *AssetServer) FindWallet(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
	return logic.NewFindWalletLogic(ctx, s.svcCtx).FindWallet(in)
}

// ResetAddress 重置钱包地址
// 接收 gRPC 请求并转发给 ResetAddressLogic 处理
// 用于为会员分配充值地址
func (s *AssetServer) ResetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AssetResp, error) {
	return logic.NewResetAddressLogic(ctx, s.svcCtx).ResetAddress(in)
}

// FindTransaction 查询会员交易记录
// 接收 gRPC 请求并转发给 FindTransactionLogic 处理
// 支持按币种、时间、类型筛选和分页
func (s *AssetServer) FindTransaction(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberTransactionList, error) {
	return logic.NewFindTransactionLogic(ctx, s.svcCtx).FindTransaction(in)
}

// GetAddress 获取指定币种的所有钱包地址
// 接收 gRPC 请求并转发给 GetAddressLogic 处理
// 用于充值监听服务获取需要监听的地址列表
func (s *AssetServer) GetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AddressList, error) {
	return logic.NewGetAddressLogic(ctx, s.svcCtx).GetAddress(in)
}
