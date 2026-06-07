// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含查询钱包列表相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletLogic 是查询钱包列表业务逻辑处理器。
//
// 该结构体负责查询用户持有的所有币种钱包信息。
type FindWalletLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewFindWalletLogic 创建查询钱包列表业务逻辑处理器实例。
func NewFindWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletLogic {
	return &FindWalletLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWallet 执行查询钱包列表业务逻辑。
//
// 查询流程：
//  1. 从 context 获取已认证的用户 ID
//  2. 调用 ucenter-rpc AssetClient 查询钱包列表
//  3. 转换 RPC 响应为 API 响应格式
//
// 返回的钱包信息：
//   - 钱包 ID
//   - 充值地址
//   - 可用余额
//   - 冻结余额
//   - 待释放余额
//   - 关联的币种信息
//
// RPC 调用：
//   - AssetClient.FindWallet -> ucenter-rpc
//   - ucenter-rpc 负责：查询会员所有钱包、组装币种信息
//
// 用户身份获取：
//   - 使用 middleware.UserIDFromContext 获取用户 ID
//
// 返回：
//   - []*types.MemberWallet：钱包列表
//   - error：查询失败时的错误信息
func (l *FindWalletLogic) FindWallet() ([]*types.MemberWallet, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 调用 ucenter-rpc AssetClient 查询钱包列表
	payload, err := l.svcCtx.AssetClient.FindWallet(ctx, &assetpb.AssetReq{UserId: userID})
	if err != nil {
		return nil, err
	}

	// 使用 copier 进行类型转换
	// copier 会自动匹配同名字段进行赋值
	resp := make([]*types.MemberWallet, len(payload.List))
	if err := copier.Copy(&resp, payload.List); err != nil {
		return nil, err
	}
	return resp, nil
}
