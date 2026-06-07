// Package logic 定义注册相关业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// RegisterByPhoneLogic 手机号注册逻辑处理器
// 处理会员手机号注册 RPC 请求
type RegisterByPhoneLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewRegisterByPhoneLogic 创建逻辑处理器实例
func NewRegisterByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterByPhoneLogic {
	return &RegisterByPhoneLogic{ctx: ctx, svcCtx: svcCtx}
}

// RegisterByPhone 处理手机号注册请求
// 调用 MemberService.RegisterByPhone 处理注册逻辑
//
// 注册流程：
//  1. 验证人机验证码
//  2. 验证短信验证码
//  3. 检查手机号是否已注册
//  4. 创建会员记录
//
// 参数：
//   - req: 注册请求，包含手机号、用户名、密码、验证码等
//
// 返回：
//   - RegRes: 注册响应（空响应）
//   - error: 错误信息
func (l *RegisterByPhoneLogic) RegisterByPhone(req *registerpb.RegReq) (*registerpb.RegRes, error) {
	return l.svcCtx.MemberService.RegisterByPhone(l.ctx, req)
}
