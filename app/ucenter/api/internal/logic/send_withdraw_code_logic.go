// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含发送提现验证码相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/rpc/pb/member"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// SendWithdrawCodeLogic 是发送提现验证码业务逻辑处理器。
//
// 该结构体负责发送提现操作所需的短信验证码。
type SendWithdrawCodeLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewSendWithdrawCodeLogic 创建发送提现验证码业务逻辑处理器实例。
func NewSendWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawCodeLogic {
	return &SendWithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 执行发送提现验证码业务逻辑。
//
// 发送流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 调用 ucenter-rpc MemberClient 查询会员手机号
//  3. 调用 ucenter-rpc WithdrawClient 发送验证码
//
// 为什么验证码发送到会员手机而非请求中指定的号码：
//   - 安全考虑：防止攻击者将验证码发送到其他手机
//   - 确保验证码只能被账号持有者接收
//
// RPC 调用链：
//   1. MemberClient.FindMemberById -> ucenter-rpc (member pb)
//      - 获取会员已认证的手机号
//   2. WithdrawClient.SendCode -> ucenter-rpc (withdraw pb)
//      - 发送提现验证码到会员手机
//
// 安全考虑：
//   - 验证码只发送到用户已认证的手机号
//   - 应有发送频率限制，防止短信轰炸
//   - 验证码有效期通常较短（如 5 分钟）
//
// 返回：
//   - string："success" 表示发送成功
//   - error：发送失败时的错误信息
func (l *SendWithdrawCodeLogic) SendCode() (string, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 MemberClient 获取会员信息
	// 需要获取会员已认证的手机号用于发送验证码
	memberInfo, err := l.svcCtx.MemberClient.FindMemberById(ctx, &member.MemberReq{MemberId: userID})
	if err != nil {
		return "", err
	}

	// 调用 WithdrawClient 发送提现验证码
	// 验证码将发送到会员已认证的手机号
	_, err = l.svcCtx.WithdrawClient.SendCode(ctx, &withdrawpb.WithdrawReq{Phone: memberInfo.MobilePhone})
	if err != nil {
		return "", err
	}

	return "success", nil
}
