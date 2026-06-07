package service

import (
	"context"
	"errors"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
)

// ExchangeCoinService 持有可见交易对相关的业务规则。
//
// 业务职责：
//   - 管理交易对查询
//   - 验证交易对是否存在
//   - 提供可见交易对列表
//
// 依赖：
//   - repo：ExchangeCoinRepository，用于数据访问
//
// 调用关系：
//   - MarketService 依赖本服务获取可见交易对列表
type ExchangeCoinService struct {
	repo *repository.ExchangeCoinRepository
}

// NewExchangeCoinService 创建 ExchangeCoinService 实例。
//
// 参数：
//   - repo：交易对数据仓库
func NewExchangeCoinService(repo *repository.ExchangeCoinRepository) *ExchangeCoinService {
	return &ExchangeCoinService{repo: repo}
}

// FindVisible 获取所有可见的交易对。
//
// 业务规则：
//   - 只返回 visible=1 的交易对
//   - 用于行情展示、交易对选择器等场景
//
// 参数：
//   - ctx：请求上下文
//
// 返回：
//   - []*model.ExchangeCoin：可见交易对列表
func (s *ExchangeCoinService) FindVisible(ctx context.Context) ([]*model.ExchangeCoin, error) {
	return s.repo.FindVisible(ctx)
}

// FindBySymbol 根据 symbol 查询交易对详情。
//
// 业务规则：
//   - symbol 格式为 "BASEQUOTE"，如 "BTCUSDT"、"ETHUSDT"
//   - 如果交易对不存在，返回明确的业务错误
//
// 参数：
//   - ctx：请求上下文
//   - symbol：交易对标识
//
// 返回：
//   - *model.ExchangeCoin：交易对完整信息
//   - error：交易对不存在时返回 "trading pair not found"
func (s *ExchangeCoinService) FindBySymbol(ctx context.Context, symbol string) (*model.ExchangeCoin, error) {
	item, err := s.repo.FindBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("trading pair not found")
	}
	return item, nil
}
