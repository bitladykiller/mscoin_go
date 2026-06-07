package main

import (
	"flag"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/config"
	"mscoin_go/app/ucenter/rpc/internal/server"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/ucenter.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		registerpb.RegisterRegisterServer(grpcServer, server.NewRegisterServer(ctx))
		loginpb.RegisterLoginServer(grpcServer, server.NewLoginServer(ctx))
		memberpb.RegisterMemberServer(grpcServer, server.NewMemberServer(ctx))
		assetpb.RegisterAssetServer(grpcServer, server.NewAssetServer(ctx))
		withdrawpb.RegisterWithdrawServer(grpcServer, server.NewWithdrawServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting ucenter rpc server at %s...\n", c.ListenOn)
	s.Start()
}
