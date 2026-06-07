// Package logic 定义验证码发送业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// SendCodeLogic 发送验证码逻辑处理器
// 处理发送注册验证码 RPC 请求
type SendCodeLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewSendCodeLogic 创建逻辑处理器实例
func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 发送注册验证码
// 调用 MemberService.SendRegisterCode 处理验证码发送逻辑
//
// 验证码规则：
//   - 长度：4 位数字
//   - 有效期：15 分钟
//   - 缓存键：REGISTER::{phone}
//
// 注意：实际发送短信由外部服务完成，本方法只负责生成和缓存验证码
//
// 参数：
//   - req: 验证码请求，包含手机号
//
// 返回：
//   - NoRes: 空响应
//   - error: 错误信息
func (l *SendCodeLogic) SendCode(req *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return l.svcCtx.MemberService.SendRegisterCode(l.ctx, req)
}
