// Package logic 提供 market-api 服务的业务逻辑实现。
// 本文件实现交易对信息查询的业务逻辑。
package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// SymbolInfoLogic 实现交易对信息查询的业务逻辑。
// 该逻辑通过 RPC 调用 market-rpc 服务获取交易对的详细配置信息。
//
// 业务流程：
//  1. 接收交易对代码参数（如 "BTCUSDT"）
//  2. 调用 MarketClient.FindSymbolInfo RPC 方法
//  3. 将 RPC 响应转换为 API 响应格式
//  4. 返回交易对配置信息
//
// 返回信息包括：
//   - 交易对基本信息（ID, Symbol, 基础货币, 计价货币）
//   - 精度配置（价格精度, 数量精度）
//   - 交易限制（最小/最大成交量, 最小成交额）
//   - 状态信息（启用状态, 可见性, 引擎状态）
type SymbolInfoLogic struct {
	// marketLogicBase 嵌入基类，提供 ctx 和 svcCtx 字段
	marketLogicBase
}

// NewSymbolInfoLogic 创建 SymbolInfoLogic 实例的工厂函数。
// 该函数遵循 go-zero 的 Logic 创建模式，便于依赖注入和测试。
//
// 参数：
//   - ctx: 请求上下文，用于超时控制和请求追踪
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回：
//   - *SymbolInfoLogic: 初始化完成的业务逻辑实例
func NewSymbolInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolInfoLogic {
	return &SymbolInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

// SymbolInfo 执行交易对信息查询的业务逻辑。
// 该方法通过 RPC 调用获取指定交易对的详细配置信息。
//
// 参数：
//   - req: 市场请求参数，包含以下字段：
//   - IP: 客户端 IP 地址，用于地理位置相关的处理
//   - Symbol: 交易对代码，如 "BTCUSDT"
//
// 返回：
//   - *types.ExchangeCoinResp: 交易对配置信息
//   - error: 错误信息，RPC 调用失败或数据转换失败时返回
//
// 处理步骤：
//  1. 创建带超时的子上下文（10秒超时）
//  2. 调用 MarketClient.FindSymbolInfo RPC 方法
//  3. 使用 copier 将 RPC 响应转换为 API 响应格式
//  4. 返回转换后的结果
//
// 数据转换说明：
//   - RPC 响应和 API 响应字段名和类型基本一致
//   - copier 自动进行字段映射和类型转换
//   - 字段名不一致时需要手动映射
//
// 超时说明：
//   - 设置 10 秒超时，交易对信息查询可能涉及数据库操作
//   - 超时后上下文取消，RPC 调用会被中断
func (l *SymbolInfoLogic) SymbolInfo(req types.MarketReq) (*types.ExchangeCoinResp, error) {
	// 创建带超时的子上下文，超时时间 10 秒
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel() // 确保资源释放

	// 调用 RPC 服务获取交易对信息
	// 传递 IP 和 Symbol 参数
	payload, err := l.svcCtx.MarketClient.FindSymbolInfo(ctx, &marketpb.MarketReq{
		Ip:     req.IP,
		Symbol: req.Symbol,
	})
	if err != nil {
		// RPC 调用失败，返回错误
		return nil, err
	}

	// 将 RPC 响应转换为 API 响应格式
	// 使用 copier 进行结构体字段映射
	resp := &types.ExchangeCoinResp{}
	if err := copier.Copy(resp, payload); err != nil {
		// 数据转换失败，返回错误
		return nil, err
	}

	return resp, nil
}