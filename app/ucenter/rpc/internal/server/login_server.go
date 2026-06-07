// Package server 定义 gRPC 服务端。
//
// 本包包含各 RPC 服务的服务端实现，负责：
//   - 实现 gRPC 服务接口
//   - 接收 gRPC 请求并转发给 Logic 层
//   - 返回 gRPC 响应
//
// 设计原则：
//   - 薄层设计：Server 层只做请求转发，不包含业务逻辑
//   - 单一职责：每个 Server 结构体只实现一个 RPC 服务
//   - 依赖注入：通过 ServiceContext 获取 Logic 处理器
//
// 分层架构：
//   - Server 层：gRPC 服务端，接收请求
//   - Logic 层：业务逻辑处理器，转发请求
//   - Service 层：领域服务，处理业务逻辑
//   - Repository 层：数据仓储，访问数据库
package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

// LoginServer 登录 RPC 服务端
// 处理会员登录相关的 RPC 请求
//
// 实现 loginpb.LoginServer 接口
// 提供 Login 方法处理登录请求
type LoginServer struct {
	svcCtx *svc.ServiceContext       // 服务上下文
	loginpb.UnimplementedLoginServer // 未实现方法的默认实现
}

// NewLoginServer 创建登录服务端实例
// 参数 svcCtx 为服务上下文，包含所有业务服务
func NewLoginServer(svcCtx *svc.ServiceContext) *LoginServer {
	return &LoginServer{svcCtx: svcCtx}
}

// Login 处理会员登录请求
// 接收 gRPC 请求并转发给 LoginLogic 处理
//
// 参数：
//   - ctx: 请求上下文
//   - in: 登录请求，包含用户名、密码、验证码等
//
// 返回：
//   - LoginRes: 登录响应，包含 Token、会员信息等
//   - error: 错误信息
func (s *LoginServer) Login(ctx context.Context, in *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	// 创建 Logic 处理器并调用 Login 方法
	return logic.NewLoginLogic(ctx, s.svcCtx).Login(in)
}
