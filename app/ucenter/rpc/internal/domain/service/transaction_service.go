// Package service 定义交易记录领域服务。
//
// TransactionService 是交易记录查询的领域服务，负责：
//   - 交易记录查询：支持按币种、时间、类型筛选
//   - 分页查询：支持分页参数
//
// 设计原则：
//   - 只读服务：只负责查询，不负责创建交易记录
//   - 交易记录由充值、提现、转账等服务创建
//   - 职责单一：只处理交易记录查询的业务逻辑
//
// 与其他服务的关系：
//   - 充值：由 jobcenter 处理链上充值后创建交易记录
//   - 提现：由 WithdrawService 处理提现申请后创建交易记录
//   - 转账：由 TransferService 处理转账后创建交易记录
package service

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/repository"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// TransactionService 交易服务
// 负责会员交易记录查询等业务逻辑
type TransactionService struct {
	repo *repository.TransactionRepository // 交易仓储
}

// NewTransactionService 创建交易服务实例
// 参数 repo 为交易仓储
func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

// FindTransaction 查询会员交易记录
// 支持按币种、时间范围、交易类型筛选，支持分页
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - pageNo: 页码，从 1 开始
//   - pageSize: 每页条数
//   - symbol: 币种符号筛选（可选）
//   - startTime: 开始时间筛选（可选）
//   - endTime: 结束时间筛选（可选）
//   - transactionType: 交易类型筛选（可选）
//
// 返回：
//   - list: 交易记录列表
//   - total: 总记录数（用于分页计算）
//   - error: 错误信息
//
// 交易类型：
//   - RECHARGE: 充值
//   - WITHDRAW: 提现
//   - TRANSFER_ACCOUNTS: 转账
//   - EXCHANGE: 兑换
func (s *TransactionService) FindTransaction(
	ctx context.Context,
	memberID int64,
	pageNo int64,
	pageSize int64,
	symbol string,
	startTime string,
	endTime string,
	transactionType string,
) ([]*assetpb.MemberTransaction, int64, error) {
	// 从仓储查询交易记录
	// 支持多条件筛选和分页
	list, total, err := s.repo.FindTransaction(ctx, memberID, pageNo, pageSize, symbol, startTime, endTime, transactionType)
	if err != nil {
		return nil, 0, err
	}

	// 转换为 protobuf 响应
	// 每条记录转换为前端可用的格式
	resp := make([]*assetpb.MemberTransaction, 0, len(list))
	for _, transaction := range list {
		resp = append(resp, transaction.ToProto())
	}
	return resp, total, nil
}
