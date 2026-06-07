// Package logic 定义提现验证码发送业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// SendWithdrawCodeLogic 发送提现验证码逻辑处理器
// 处理发送提现验证码 RPC 请求
type SendWithdrawCodeLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewSendWithdrawCodeLogic 创建逻辑处理器实例
func NewSendWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawCodeLogic {
	return &SendWithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 发送提现验证码
// 调用 WithdrawService.SendCode 处理验证码发送逻辑
//
// 验证码规则：
//   - 长度：6 位数字
//   - 有效期：5 分钟
//   - 缓存键：WITHDRAW::{phone}
//
// 注意：实际发送短信由外部服务完成，本方法只负责生成和缓存验证码
//
// 参数：
//   - req: 提现请求，包含手机号
//
// 返回：
//   - NoRes: 空响应
//   - error: 错误信息
func (l *SendWithdrawCodeLogic) SendCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	// 调用 WithdrawService 发送提现验证码
	if err := l.svcCtx.WithdrawService.SendCode(l.ctx, req.Phone); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
