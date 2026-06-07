// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含申请提现相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawCodeLogic 是申请提现业务逻辑处理器。
//
// 该结构体负责处理用户的提现申请请求。
type WithdrawCodeLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewWithdrawCodeLogic 创建申请提现业务逻辑处理器实例。
func NewWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawCodeLogic {
	return &WithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// WithdrawCode 执行申请提现业务逻辑。
//
// 提现流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 调用 ucenter-rpc WithdrawClient 提交提现申请
//  3. RPC 服务进行多重验证后创建提现工单
//
// 提现验证项（RPC 服务处理）：
//   - 资金密码验证：确保操作者是账号持有者
//   - 短信验证码验证：二次验证，防止密码泄露后资产被盗
//   - 余额检查：确保账户余额充足
//   - 提现限额检查：不超过单次最大提现金额
//   - 最小提现金额检查：不低于单次最小提现金额
//   - 钱包状态检查：钱包未被锁定
//   - 地址有效性检查：提现地址格式正确
//
// RPC 调用：
//   - WithdrawClient.WithdrawCode -> ucenter-rpc
//   - ucenter-rpc 负责：
//       - 验证资金密码
//       - 验证短信验证码
//       - 检查余额和限额
//       - 创建提现工单
//       - 冻冻提现金额
//       - 发送提现通知
//
// 安全设计：
//   - 双重验证：资金密码 + 短信验证码
//   - 提现金额冻结：防止重复提现
//   - 操作日志记录：便于审计追踪
//
// 参数：
//   - req：提现请求，包含币种、金额、地址、密码、验证码等
//
// 返回：
//   - string："success" 表示提现申请成功
//   - error：提现失败时的错误信息
func (l *WithdrawCodeLogic) WithdrawCode(req *types.WithdrawReq) (string, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc WithdrawClient 提交提现申请
	// RPC 服务会执行完整的提现验证流程
	_, err := l.svcCtx.WithdrawClient.WithdrawCode(ctx, &withdrawpb.WithdrawReq{
		UserId:     userID,           // 用户 ID，从 JWT Token 解析
		Unit:       req.Unit,         // 币种单位（如 USDT、BTC）
		JyPassword: req.JyPassword,   // 资金密码，用于验证操作权限
		Code:       req.Code,         // 短信验证码，二次验证
		Address:    req.Address,      // 提现目标地址
		Amount:     req.Amount,       // 提现金额
		Fee:        req.Fee,          // 矿工费（手续费）
	})
	if err != nil {
		return "fail", err
	}
	return "success", nil
}
