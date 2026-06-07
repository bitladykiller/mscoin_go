// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含查询提现记录相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
	"mscoin_go/pkg/page"
)

// WithdrawRecordLogic 是查询提现记录业务逻辑处理器。
//
// 该结构体负责查询用户的历史提现工单。
type WithdrawRecordLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewWithdrawRecordLogic 创建查询提现记录业务逻辑处理器实例。
func NewWithdrawRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawRecordLogic {
	return &WithdrawRecordLogic{ctx: ctx, svcCtx: svcCtx}
}

// Record 执行查询提现记录业务逻辑。
//
// 查询流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 处理分页参数（默认 page=1, pageSize=10）
//  3. 调用 ucenter-rpc WithdrawClient 查询提现记录
//  4. 转换为分页响应格式
//
// 返回的提现记录信息：
//   - 提现 ID
//   - 币种信息
//   - 提现金额、手续费、到账金额
//   - 提现地址、备注
//   - 区块链交易哈希
//   - 提现状态（待审核、审核通过、已完成、已拒绝等）
//   - 创建时间、处理时间
//
// RPC 调用：
//   - WithdrawClient.WithdrawRecord -> ucenter-rpc
//   - ucenter-rpc 负责：查询提现工单、分页处理、组装币种信息
//
// 提现状态说明：
//   - 0：待审核
//   - 1：审核通过/处理中
//   - 2：已完成/已打款
//   - 3：已拒绝
//   - 其他状态根据业务定义
//
// 参数：
//   - req：提现请求，包含分页参数
//
// 返回：
//   - *page.Result：分页结果，包含提现记录列表和总数
//   - error：查询失败时的错误信息
func (l *WithdrawRecordLogic) Record(req *types.WithdrawReq) (*page.Result, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 处理分页参数，设置默认值
	pageNo := req.Page
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc WithdrawClient 查询提现记录
	records, err := l.svcCtx.WithdrawClient.WithdrawRecord(ctx, &withdrawpb.WithdrawReq{
		UserId:   userID,
		Page:     int64(pageNo),
		PageSize: int64(pageSize),
	})
	if err != nil {
		return nil, err
	}

	// 转换 RPC 响应为分页结果
	// 将 protobuf 列表转换为通用切片，供 page.New 处理
	items := make([]any, len(records.GetList()))
	for index, item := range records.GetList() {
		items[index] = item
	}

	// 构建分页响应
	return page.New(items, int64(pageNo), int64(pageSize), records.Total), nil
}
