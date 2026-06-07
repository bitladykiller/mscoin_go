// Package config 提供 market-api 服务的配置结构定义。
// 配置通过 YAML 文件加载，包含 HTTP 服务器配置和 RPC 客户端配置。
//
// 配置文件示例 (etc/market-api.yaml):
//
//	Name: market-api
//	Host: 0.0.0.0
//	Port: 8888
//	MarketRPC:
//	  Etcd:
//	    Hosts:
//	      - 127.0.0.1:2379
//	    Key: market.rpc
//
// 该包被 main 包引用，用于加载和解析配置文件。
package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 定义 market API 服务的完整运行时配置。
// 该结构体与 YAML 配置文件一一对应，由 go-zero 框架自动解析填充。
//
// 配置项说明:
//   - RestConf: HTTP REST 服务器配置（端口、主机、超时等）
//   - MarketRPC: market-rpc 微服务的连接配置
//
// 使用示例:
//
//	var c config.Config
//	conf.MustLoad("etc/market-api.yaml", &c)
type Config struct {
	// RestConf 嵌入 go-zero 的标准 REST 服务器配置。
	// 包含以下常用字段（通过 YAML 配置）:
	//   - Name: 服务名称，用于日志和监控标识
	//   - Host: 监听地址，如 "0.0.0.0" 表示监听所有网卡
	//   - Port: 监听端口号
	//   - Timeout: 请求超时时间（毫秒）
	//   - MaxConns: 最大连接数
	//   - MaxBytes: 请求体最大字节数
	rest.RestConf

	// MarketRPC 配置 market-rpc 微服务的客户端连接。
	// 支持两种服务发现方式:
	//   1. 直连模式: 直接指定 Endpoints 地址列表
	//   2. 服务发现模式: 通过 Etcd 注册中心动态发现服务
	//
	// YAML 配置示例（Etcd 模式）:
	//   MarketRPC:
	//     Etcd:
	//       Hosts:
	//         - 127.0.0.1:2379
	//       Key: market.rpc
	//
	// API 层通过此配置建立与 market-rpc 的 gRPC 连接，
	// 将 HTTP 请求转发到后端的 RPC 服务处理。
	MarketRPC zrpc.RpcClientConf
}