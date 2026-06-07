// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现币种信息查询的业务逻辑。
package logic

import (
	"context"
	"errors"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// CoinInfoLogic 实现币种信息查询的业务逻辑。
// 该逻辑通过 RPC 调用 market-rpc 服务获取币种详细信息。
//
// 业务流程：
//  1. 接收币种单位参数（如 "BTC", "ETH"）
//  2. 调用 MarketClient.FindCoinInfo RPC 方法
//  3. 将 RPC 响应转换为 API 响应格式
//  4. 返回币种详细信息
type CoinInfoLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewCoinInfoLogic 创建 CoinInfoLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *CoinInfoLogic: 初始化完成的业务逻辑实例
func NewCoinInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoinInfoLogic {
	return &CoinInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// CoinInfo 执行币种信息查询的业务逻辑。
// 该方法通过 RPC 调用获取指定币种的详细信息。
//
// 参数：
//   - req: 市场请求参数，包含 unit（币种单位）
//
// 返回：
//   - *types.Coin: 币种详细信息，包括名称、状态、充提配置等
//   - error: 错误信息，RPC 调用失败或数据转换失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（5秒超时）
//  2. 调用 MarketClient.FindCoinInfo RPC 方法
//  3. 使用 copier 将 RPC 响应转换为 API 响应格式
//  4. 返回转换后的结果
//
// 错误处理：
//   - RPC 调用失败：返回原始错误
//   - 数据转换失败：返回 "market coin payload copy failed" 错误
//
// 超时说明：
//   - 设置 5 秒超时，防止长时间阻塞
//   - 超时后上下文取消，RPC 调用会被中断
func (l *CoinInfoLogic) CoinInfo(req *types.MarketReq) (*types.Coin, error) {
	// 创建带超时的子上下文，超时时间 5 秒
	// 币种信息查询是轻量级操作，超时时间较短
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取币种信息
	// MarketClient.FindCoinInfo 会根据 unit 参数查询数据库
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: req.Unit})
	if err != nil {
		// RPC 调用失败，返回错误
		return nil, err
	}

	// 将 RPC 响应转换为 API 响应格式
	// 使用 copier 进行结构体字段映射
	resp := &types.Coin{}
	if err := copier.Copy(resp, coin); err != nil {
		// 数据转换失败，返回明确的错误信息
		return nil, errors.New("market coin payload copy failed")
	}

	return resp, nil
}