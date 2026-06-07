// Package service 包含 market 模块面向业务的领域服务。
//
// Domain Service 层职责：
//   - 封装核心业务规则和业务逻辑
//   - 协调一个或多个 Repository 完成业务用例
//   - 不依赖传输层（gRPC/HTTP），保持业务纯净
//   - 可被多个 Logic 复用
//
// 调用链路：
//
//	Logic -> Domain Service -> Repository
//
// 本包包含四个领域服务：
//   - CoinService：币种相关业务规则
//   - ExchangeCoinService：交易对相关业务规则
//   - MarketService：市场数据（K 线、缩略图）业务规则
//   - RateService：汇率查询业务规则
package service

import (
	"context"
	"errors"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
)

// CoinService 持有币种相关的业务规则。
//
// 业务职责：
//   - 验证币种是否存在
//   - 提供币种查询服务
//   - 封装币种相关的业务判断
//
// 依赖：
//   - repo：CoinRepository，用于数据访问
type CoinService struct {
	repo *repository.CoinRepository
}

// NewCoinService 创建 CoinService 实例。
//
// 参数：
//   - repo：币种数据仓库
func NewCoinService(repo *repository.CoinRepository) *CoinService {
	return &CoinService{repo: repo}
}

// FindCoinInfo 根据 unit 查询币种详情。
//
// 业务规则：
//   - unit 为币种单位标识，如 "BTC"、"ETH"、"USDT"
//   - 如果币种不存在，返回明确的业务错误
//
// 参数：
//   - ctx：请求上下文
//   - unit：币种单位标识
//
// 返回：
//   - *model.Coin：币种完整信息
//   - error：币种不存在时返回 "not support this coin: {unit}"
func (s *CoinService) FindCoinInfo(ctx context.Context, unit string) (*model.Coin, error) {
	coin, err := s.repo.FindByUnit(ctx, unit)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, errors.New("not support this coin: " + unit)
	}
	return coin, nil
}

// FindCoinByID 根据 ID 查询币种详情。
//
// 业务规则：
//   - ID 为数据库主键
//   - 如果币种不存在，返回明确的业务错误
//
// 参数：
//   - ctx：请求上下文
//   - id：币种 ID
//
// 返回：
//   - *model.Coin：币种完整信息
//   - error：币种不存在时返回 "not support this coin"
func (s *CoinService) FindCoinByID(ctx context.Context, id int64) (*model.Coin, error) {
	coin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if coin == nil {
		return nil, errors.New("not support this coin")
	}
	return coin, nil
}

// FindAll 获取所有币种列表。
//
// 业务规则：
//   - 返回所有币种，不区分状态
//   - 调用方如需过滤，自行处理
//
// 参数：
//   - ctx：请求上下文
//
// 返回：
//   - []*model.Coin：币种列表
func (s *CoinService) FindAll(ctx context.Context) ([]*model.Coin, error) {
	return s.repo.FindAll(ctx)
}
