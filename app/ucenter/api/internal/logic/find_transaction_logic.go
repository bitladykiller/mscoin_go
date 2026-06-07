// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含查询交易记录相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	"mscoin_go/pkg/page"
)

// FindTransactionLogic 是查询交易记录业务逻辑处理器。
//
// 该结构体负责查询用户的资产交易历史记录。
type FindTransactionLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewFindTransactionLogic 创建查询交易记录业务逻辑处理器实例。
func NewFindTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTransactionLogic {
	return &FindTransactionLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindTransaction 执行查询交易记录业务逻辑。
//
// 查询流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 处理分页参数（默认 pageNo=1, pageSize=10）
//  3. 调用 ucenter-rpc AssetClient 查询交易记录
//  4. 转换为分页响应格式
//
// 支持的筛选条件：
//   - 时间范围：StartTime ~ EndTime
//   - 币种筛选：Symbol
//   - 交易类型：Type（充值、提现、转账等）
//
// RPC 调用：
//   - AssetClient.FindTransaction -> ucenter-rpc
//   - ucenter-rpc 负责：查询交易记录、按条件筛选、分页处理
//
// 参数：
//   - req：资产请求，包含分页和筛选参数
//
// 返回：
//   - *page.Result：分页结果，包含交易记录列表和总数
//   - error：查询失败时的错误信息
func (l *FindTransactionLogic) FindTransaction(req *types.AssetReq) (*page.Result, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 处理分页参数，设置默认值
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc AssetClient 查询交易记录
	payload, err := l.svcCtx.AssetClient.FindTransaction(ctx, &assetpb.AssetReq{
		UserId:    userID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		PageNo:    int64(pageNo),
		PageSize:  int64(pageSize),
		Symbol:    req.Symbol,
		Type:      req.Type,
	})
	if err != nil {
		return nil, err
	}

	// 转换 RPC 响应为分页结果
	// 将 protobuf 列表转换为通用切片，供 page.New 处理
	items := make([]any, len(payload.List))
	for index, item := range payload.List {
		items[index] = item
	}

	// 构建分页响应
	// page.New 会计算总页数等分页信息
	return page.New(items, int64(pageNo), int64(pageSize), payload.Total), nil
}
