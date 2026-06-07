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

// configFile 配置文件路径参数
var configFile = flag.String("f", "etc/ucenter.yaml", "配置文件路径")

func main() {
	flag.Parse()

	// 加载配置文件
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	// 创建 gRPC 服务端
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册各 RPC 服务
		registerpb.RegisterRegisterServer(grpcServer, server.NewRegisterServer(ctx)) // 注册服务
		loginpb.RegisterLoginServer(grpcServer, server.NewLoginServer(ctx))          // 登录服务
		memberpb.RegisterMemberServer(grpcServer, server.NewMemberServer(ctx))       // 会员服务
		assetpb.RegisterAssetServer(grpcServer, server.NewAssetServer(ctx))          // 资产服务
		withdrawpb.RegisterWithdrawServer(grpcServer, server.NewWithdrawServer(ctx)) // 提现服务

		// 在开发和测试模式下启用 gRPC 反射服务
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动服务
	fmt.Printf("Starting ucenter rpc server at %s...\n", c.ListenOn)
	s.Start()
}