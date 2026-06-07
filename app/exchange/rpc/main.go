// Package main 是 exchange-rpc 服务的入口点。
// 该服务提供交易订单相关的 RPC 接口，包括下单、查询订单、取消订单等功能。
package main

import (
	"flag"
	"fmt"

	"mscoin_go/app/exchange/rpc/internal/config"
	"mscoin_go/app/exchange/rpc/internal/server"
	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// configFile 指定配置文件路径，默认为 "etc/exchange.yaml"。
var configFile = flag.String("f", "etc/exchange.yaml", "配置文件路径")

// main 是 exchange-rpc 服务的入口函数。
// 主要流程：
// 1. 解析命令行参数，加载配置文件
// 2. 初始化服务上下文（数据库连接、Redis 缓存、RPC 客户端等）
// 3. 创建 gRPC 服务器，注册订单服务
// 4. 在开发/测试模式下启用 gRPC 反射服务
// 5. 启动服务器监听请求
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化服务上下文
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	// 创建 gRPC 服务器并注册服务
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		orderpb.RegisterOrderServer(grpcServer, server.NewOrderServer(ctx))
		// 在开发/测试模式下启用 gRPC 反射，便于调试
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动 RPC 服务器
	fmt.Printf("Starting exchange rpc server at %s...\n", c.ListenOn)
	s.Start()
}
