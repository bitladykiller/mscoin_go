// Package logic 定义 RPC 业务逻辑处理器。
//
// 本包包含各 RPC 方法的逻辑处理器，负责：
//   - 接收 gRPC 请求
//   - 调用领域服务处理业务逻辑
//   - 返回 gRPC 响应
//
// 设计原则：
//   - 薄层设计：Logic 层只做请求转发，不包含复杂业务逻辑
//   - 单一职责：每个 Logic 结构体只处理一个 RPC 方法
//   - 依赖注入：通过 ServiceContext 获取领域服务
//
// 分层架构：
//   - Server 层：gRPC 服务端，接收请求
//   - Logic 层：业务逻辑处理器，转发请求
//   - Service 层：领域服务，处理业务逻辑
//   - Repository 层：数据仓储，访问数据库
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

// LoginLogic 登录逻辑处理器
// 处理会员登录 RPC 请求
type LoginLogic struct {
	ctx    context.Context    // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewLoginLogic 创建逻辑处理器实例
// 参数：
//   - ctx: 请求上下文，用于传递 trace 信息和超时控制
//   - svcCtx: 服务上下文，包含所有领域服务
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

// Login 处理会员登录请求
// 调用 MemberService.Login 处理登录逻辑
//
// 参数：
//   - req: 登录请求，包含用户名、密码、验证码等
//
// 返回：
//   - LoginRes: 登录响应，包含 Token、会员信息等
//   - error: 错误信息
func (l *LoginLogic) Login(req *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	return l.svcCtx.MemberService.Login(l.ctx, req)
}
