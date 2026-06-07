// Package logic 定义提现申请业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawCodeLogic 提现申请逻辑处理器
// 处理会员提现申请 RPC 请求
type WithdrawCodeLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewWithdrawCodeLogic 创建逻辑处理器实例
func NewWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawCodeLogic {
	return &WithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// WithdrawCode 处理提现申请
// 调用 WithdrawService.Apply 处理提现申请逻辑
//
// 提现流程：
//  1. 验证请求参数（金额、地址、验证码等）
//  2. 验证 Redis 中的验证码
//  3. 验证交易密码
//  4. 在事务中冻结余额、创建提现记录
//  5. 发布 Kafka 事件通知下游处理
//
// 参数：
//   - req: 提现请求，包含用户 ID、币种、金额、地址、验证码、交易密码等
//
// 返回：
//   - NoRes: 空响应
//   - error: 错误信息
//
// 事务安全：
//   - 使用 FOR UPDATE 行锁防止并发提现超扣
//   - 余额冻结和提现记录创建在同一事务中
func (l *WithdrawCodeLogic) WithdrawCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	// 调用 WithdrawService 处理提现申请
	if err := l.svcCtx.WithdrawService.Apply(l.ctx, req); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
