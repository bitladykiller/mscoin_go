// Package main 是 ucenter RPC 服务的入口点。
//
// ucenter（用户中心）服务是 MSCoin 平台的核心微服务之一，负责处理：
//   - 会员注册与登录认证
//   - 会员信息管理
//   - 钱包资产管理
//   - 交易记录查询
//   - 提现申请处理
//
// 该服务采用 go-zero 框架构建，通过 gRPC 协议对外提供服务。
// 服务依赖包括：
//   - MySQL：持久化会员、钱包、交易等数据
//   - Redis：缓存验证码、会话等临时数据
//   - Kafka：发布提现等异步事件
//   - Market RPC：获取币种市场信息
//   - Bitcoin Node：BTC 地址分配
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
// 默认值为 "etc/ucenter.yaml"，可通过命令行参数 -f 指定
var configFile = flag.String("f", "etc/ucenter.yaml", "配置文件路径")

// main 是服务的入口函数
// 执行流程：
//  1. 解析命令行参数
//  2. 加载配置文件
//  3. 初始化服务上下文（数据库、缓存、消息队列等）
//  4. 创建 gRPC 服务端并注册各业务服务
//  5. 启动服务监听
func main() {
	flag.Parse()

	// 加载配置文件
	// 使用 go-zero 的 conf.MustLoad 加载 YAML 配置
	// 配置文件不存在或格式错误时会 panic
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	// 初始化数据库连接、Redis 客户端、Kafka 生产者等依赖
	// defer 确保服务退出时正确释放资源
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	// 创建 gRPC 服务端
	// zrpc.MustNewServer 封装了 gRPC 服务端的创建和配置
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册各 RPC 服务
		// 每个服务对应一个独立的 protobuf 定义和业务逻辑
		registerpb.RegisterRegisterServer(grpcServer, server.NewRegisterServer(ctx)) // 注册服务：处理会员注册
		loginpb.RegisterLoginServer(grpcServer, server.NewLoginServer(ctx))          // 登录服务：处理会员登录认证
		memberpb.RegisterMemberServer(grpcServer, server.NewMemberServer(ctx))       // 会员服务：处理会员信息查询
		assetpb.RegisterAssetServer(grpcServer, server.NewAssetServer(ctx))          // 资产服务：处理钱包和交易查询
		withdrawpb.RegisterWithdrawServer(grpcServer, server.NewWithdrawServer(ctx)) // 提现服务：处理提现申请

		// 在开发和测试模式下启用 gRPC 反射服务
		// 反射服务允许 grpcurl、grpc_cli 等工具动态发现服务定义
		// 生产环境禁用以减少攻击面
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动服务
	// 打印启动日志并开始监听配置中指定的端口
	fmt.Printf("Starting ucenter rpc server at %s...\n", c.ListenOn)
	s.Start()
}