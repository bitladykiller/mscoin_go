// Package main 是 exchange-api 服务的入口点。
// 该服务提供交易订单相关的 HTTP 接口，包括下单、查询当前订单和历史订单等功能。
package main

import (
	"flag"
	"fmt"
	"net/http"

	"mscoin_go/app/exchange/api/internal/config"
	"mscoin_go/app/exchange/api/internal/handler"
	"mscoin_go/app/exchange/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// configFile 指定配置文件路径，默认为 "etc/exchange-api.yaml"。
var configFile = flag.String("f", "etc/exchange-api.yaml", "配置文件路径")

// main 是 exchange-api 服务的入口函数。
// 主要流程：
// 1. 解析命令行参数，加载配置文件
// 2. 创建 REST 服务器，配置跨域支持
// 3. 初始化服务上下文，注册路由处理器
// 4. 启动服务器监听请求
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建 REST 服务器，配置自定义跨域规则
	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCustomCors(func(header http.Header) {
			header.Set("Access-Control-Allow-Headers", "DNT,X-Mx-ReqToken,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization,token,x-auth-token")
		}, nil, "*"),
	)
	defer server.Stop()

	// 初始化服务上下文并注册路由处理器
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 启动服务器
	fmt.Printf("Starting exchange api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
