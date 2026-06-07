// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含查询可提现币种信息相关的业务逻辑。
package logic

import (
	"context"
	"fmt"
	"time"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// QueryWithdrawCoinLogic 是查询可提现币种信息业务逻辑处理器。
//
// 该结构体负责聚合多个 RPC 服务的数据，为提现页面提供完整的展示信息。
type QueryWithdrawCoinLogic struct {
	// ctx 是请求上下文，包含已认证的用户 ID。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewQueryWithdrawCoinLogic 创建查询可提现币种信息业务逻辑处理器实例。
func NewQueryWithdrawCoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryWithdrawCoinLogic {
	return &QueryWithdrawCoinLogic{ctx: ctx, svcCtx: svcCtx}
}

// QueryWithdrawCoin 执行查询可提现币种信息业务逻辑。
//
// 该方法是提现页面初始化的核心接口，聚合以下数据：
//   - 币种列表和配置（来自 market-rpc）
//   - 用户各币种钱包余额（来自 ucenter-rpc AssetClient）
//   - 用户已保存的提现地址列表（来自 ucenter-rpc WithdrawClient）
//
// 数据聚合流程：
//  1. 从 market-rpc 获取所有币种配置（限额、手续费等）
//  2. 从 ucenter-rpc 获取用户钱包列表和余额
//  3. 遍历钱包列表，为每个币种获取已保存的提现地址
//  4. 按币种聚合上述数据，生成 WithdrawWalletInfo 列表
//
// RPC 调用链：
//   1. MarketClient.FindAllCoin -> market-rpc
//      - 获取所有币种的提现配置
//   2. AssetClient.FindWallet -> ucenter-rpc
//      - 获取用户各币种钱包余额
//   3. WithdrawClient.FindAddressByCoinId -> ucenter-rpc (每个币种调用一次)
//      - 获取用户该币种已保存的提现地址列表
//
// 返回的币种提现信息：
//   - 币种基本信息（名称、单位等）
//   - 提现限额（最小/最大金额）
//   - 矿工费范围（最小/最大手续费）
//   - 用户余额
//   - 自动提现支持情况
//   - 用户已保存的提现地址列表
//
// 返回：
//   - []*types.WithdrawWalletInfo：币种提现信息列表
//   - error：查询失败时的错误信息
func (l *QueryWithdrawCoinLogic) QueryWithdrawCoin() ([]*types.WithdrawWalletInfo, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 从 context 获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)

	// 1. 调用 market-rpc 获取所有币种配置
	// 用于获取提现限额、手续费等信息
	coinList, err := l.svcCtx.MarketClient.FindAllCoin(ctx, &marketpb.MarketReq{})
	if err != nil {
		return nil, err
	}

	// 构建币种映射，便于后续按单位查找
	coinMap := make(map[string]*marketpb.Coin, len(coinList.GetList()))
	for _, coin := range coinList.GetList() {
		coinMap[coin.Unit] = coin
	}

	// 2. 调用 ucenter-rpc AssetClient 获取用户钱包列表
	walletList, err := l.svcCtx.AssetClient.FindWallet(ctx, &assetpb.AssetReq{UserId: userID})
	if err != nil {
		return nil, err
	}

	// 3. 遍历钱包，聚合币种信息和提现地址
	resp := make([]*types.WithdrawWalletInfo, 0, len(walletList.GetList()))
	for _, wallet := range walletList.GetList() {
		// 检查钱包关联的币种信息是否存在
		if wallet.GetCoin() == nil {
			return nil, fmt.Errorf("wallet coin info missing")
		}

		// 从币种映射中获取配置信息
		coin, ok := coinMap[wallet.GetCoin().GetUnit()]
		if !ok {
			return nil, fmt.Errorf("coin %s not found", wallet.GetCoin().GetUnit())
		}

		// 3.1 调用 WithdrawClient 获取该币种已保存的提现地址
		addressList, err := l.svcCtx.WithdrawClient.FindAddressByCoinId(ctx, &withdrawpb.WithdrawReq{
			UserId: userID,
			CoinId: int64(coin.Id),
		})
		if err != nil {
			return nil, err
		}

		// 4. 构建聚合响应结构
		item := &types.WithdrawWalletInfo{
			Unit:            coin.Unit,                      // 币种单位
			Threshold:       coin.WithdrawThreshold,         // 提现阈值
			MinAmount:       coin.MinWithdrawAmount,         // 最小提现金额
			MaxAmount:       coin.MaxWithdrawAmount,         // 最大提现金额
			MinTxFee:        coin.MinTxFee,                  // 最小矿工费
			MaxTxFee:        coin.MaxTxFee,                  // 最大矿工费
			NameCn:          coin.NameCn,                    // 中文名称
			Name:            coin.Name,                      // 英文名称
			Balance:         wallet.Balance,                 // 用户余额
			CanAutoWithdraw: autoWithdrawString(coin.CanAutoWithdraw), // 是否支持自动提现
			WithdrawScale:   coin.WithdrawScale,             // 提现精度
			AccountType:     coin.AccountType,               // 账户类型
			Addresses:       make([]types.AddressSimple, 0, len(addressList.GetList())), // 已保存地址列表
		}

		// 填充已保存的提现地址
		for _, address := range addressList.GetList() {
			item.Addresses = append(item.Addresses, types.AddressSimple{
				Remark:  address.Remark,  // 地址备注
				Address: address.Address, // 提现地址
			})
		}
		resp = append(resp, item)
	}

	return resp, nil
}

// autoWithdrawString 将自动提现标识转换为字符串格式。
//
// 参数：
//   - value：自动提现标识（0: 支持，其他: 不支持）
//
// 返回：
//   - string："true" 表示支持自动提现，"false" 表示不支持
func autoWithdrawString(value int32) string {
	if value == 0 {
		return "true"
	}
	return "false"
}
