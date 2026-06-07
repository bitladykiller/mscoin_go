// Package main 是 ucenter-api 服务的入口包。
//
// ucenter-api 是用户中心 HTTP API 服务，提供用户注册、登录、钱包管理、提现等功能的 RESTful 接口。
// 该服务基于 go-zero 框架构建，采用分层架构：
//   - handler 层：处理 HTTP 请求解析和响应
//   - logic 层：实现业务逻辑，调用 RPC 服务
//   - middleware 层：提供 JWT 认证等中间件功能
//
// 服务依赖：
//   - ucenter-rpc：用户中心 RPC 服务，提供注册、登录、会员管理、资产管理、提现等功能
//   - market-rpc：市场 RPC 服务，提供币种信息查询功能
//
// 启动方式：go run main.go -f etc/ucenter-api.yaml
package main

import (
	"flag"
	"fmt"
	"net/http"

	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/handler"
	"mscoin_go/app/ucenter/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// configFile 指定配置文件路径，默认为 etc/ucenter-api.yaml。
// 可通过命令行参数 -f 指定其他配置文件。
var configFile = flag.String("f", "etc/ucenter-api.yaml", "the config file")

// main 是 ucenter-api 服务的入口函数。
//
// 启动流程：
//  1. 解析命令行参数，获取配置文件路径
//  2. 加载配置文件，初始化 Config 结构体
//  3. 创建 REST 服务器，配置 CORS 跨域支持
//  4. 初始化 ServiceContext，建立与 RPC 服务的连接
//  5. 注册路由处理器
//  6. 启动 HTTP 服务器
//
// CORS 配置说明：
//   - 允许所有来源 (*)
//   - 允许的请求头包括：Authorization、token、x-auth-token 等认证相关头部
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCustomCors(func(header http.Header) {
			header.Set("Access-Control-Allow-Headers", "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization,token,x-auth-token")
		}, nil, "*"),
	)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting ucenter api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
