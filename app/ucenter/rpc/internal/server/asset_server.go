package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type AssetServer struct {
	svcCtx *svc.ServiceContext
	assetpb.UnimplementedAssetServer
}

func NewAssetServer(svcCtx *svc.ServiceContext) *AssetServer {
	return &AssetServer{svcCtx: svcCtx}
}

func (s *AssetServer) FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	return logic.NewFindWalletBySymbolLogic(ctx, s.svcCtx).FindWalletBySymbol(in)
}

func (s *AssetServer) FindWallet(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberWalletList, error) {
	return logic.NewFindWalletLogic(ctx, s.svcCtx).FindWallet(in)
}

func (s *AssetServer) ResetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AssetResp, error) {
	return logic.NewResetAddressLogic(ctx, s.svcCtx).ResetAddress(in)
}

func (s *AssetServer) FindTransaction(ctx context.Context, in *assetpb.AssetReq) (*assetpb.MemberTransactionList, error) {
	return logic.NewFindTransactionLogic(ctx, s.svcCtx).FindTransaction(in)
}

func (s *AssetServer) GetAddress(ctx context.Context, in *assetpb.AssetReq) (*assetpb.AddressList, error) {
	return logic.NewGetAddressLogic(ctx, s.svcCtx).GetAddress(in)
}
