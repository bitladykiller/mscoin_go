package svc

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/middleware"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	Auth           rest.Middleware
	RegisterClient registerpb.RegisterClient
	LoginClient    loginpb.LoginClient
	MemberClient   memberpb.MemberClient
	AssetClient    assetpb.AssetClient
	WithdrawClient withdrawpb.WithdrawClient
	MarketClient   marketpb.MarketClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	ucenterClient := zrpc.MustNewClient(c.UcenterRPC)
	ucenterConn := ucenterClient.Conn()
	marketClient := zrpc.MustNewClient(c.MarketRPC)
	marketConn := marketClient.Conn()

	return &ServiceContext{
		Config:         c,
		Auth:           middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
		RegisterClient: registerpb.NewRegisterClient(ucenterConn),
		LoginClient:    loginpb.NewLoginClient(ucenterConn),
		MemberClient:   memberpb.NewMemberClient(ucenterConn),
		AssetClient:    assetpb.NewAssetClient(ucenterConn),
		WithdrawClient: withdrawpb.NewWithdrawClient(ucenterConn),
		MarketClient:   marketpb.NewMarketClient(marketConn),
	}
}
