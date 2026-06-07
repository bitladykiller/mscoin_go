// Package main 是 market-api 服务的入口包。
// market-api 是一个基于 go-zero 框架构建的 HTTP REST API 服务，
// 提供市场数据查询功能，包括币种信息、交易对行情、K线历史数据、汇率查询等。
//
// 该服务作为 API 网关层，接收前端 HTTP 请求并通过 RPC 调用后端的 market-rpc 服务，
// 遵循 go-zero 的 API 网关 + RPC 微服务架构模式。
//
// 主要功能:
//   - 币种信息查询 (Coin Info)
//   - 交易对信息查询 (Symbol Info)
//   - 行情缩略图数据 (Symbol Thumb)
//   - 行情趋势数据 (Symbol Thumb Trend)
//   - K线历史数据 (History Kline)
//   - 法币汇率查询 (USD Rate)
//
// 启动方式:
//
//	go run main.go -f etc/market-api.yaml
//
// 配置文件路径通过 -f 参数指定，默认为 etc/market-api.yaml。
package main

import (
	"flag"
	"fmt"
	"net/http"

	"mscoin_go/app/market/api/internal/config"
	"mscoin_go/app/market/api/internal/handler"
	"mscoin_go/app/market/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// configFile 定义配置文件路径的命令行参数。
// 通过 -f 标志传入，默认值为 "etc/market-api.yaml"。
// 示例: ./market-api -f /path/to/config.yaml
var configFile = flag.String("f", "etc/market-api.yaml", "配置文件路径")

// main 是 market-api 服务的入口函数。
//
// 启动流程:
//  1. 解析命令行参数，获取配置文件路径
//  2. 加载配置文件到 Config 结构体
//  3. 创建 HTTP REST 服务器，配置 CORS 跨域策略
//  4. 初始化服务上下文（ServiceContext），注入 RPC 客户端依赖
//  5. 注册所有 HTTP 路由处理器
//  6. 启动 HTTP 服务器，开始监听请求
//
// 注意事项:
//   - 使用 defer 确保服务器在退出时正确关闭
//   - CORS 配置允许任意来源 (*) 访问，适用于开发环境
//   - 生产环境应考虑收紧 CORS 策略
func main() {
	// 解析命令行参数
	flag.Parse()

	// 加载配置文件
	// MustLoad 会在配置文件不存在或格式错误时 panic
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建 HTTP REST 服务器
	// 配置自定义 CORS 策略，允许前端跨域访问
	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCustomCors(func(header http.Header) {
			// 设置允许的请求头
			// 包含常用的认证头（Authorization, token, x-auth-token）和请求头
			header.Set("Access-Control-Allow-Headers", "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization,token,x-auth-token")
		}, nil, "*"), // "*" 允许任意来源访问
	)
	// 确保服务器在函数退出时关闭
	defer server.Stop()

	// 初始化服务上下文
	// 创建 RPC 客户端连接，注入配置和依赖
	ctx := svc.NewServiceContext(c)

	// 注册所有 HTTP 路由处理器
	// 将路由绑定到对应的 handler 函数
	handler.RegisterHandlers(server, ctx)

	// 打印启动信息，显示服务监听地址
	fmt.Printf("Starting market api server at %s:%d...\n", c.Host, c.Port)

	// 启动 HTTP 服务器（阻塞调用）
	server.Start()
}