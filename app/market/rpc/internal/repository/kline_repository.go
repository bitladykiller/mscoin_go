package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/market/rpc/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// KlineRepository 封装历史 K 线数据的 MongoDB 访问。
//
// 数据源：MongoDB
// 集合命名规则：exchange_kline_{symbol}_{period}
//   - symbol：交易对标识，如 "BTCUSDT"
//   - period：K 线周期，如 "1H"、"1D"、"15m"
//
// 为什么使用 MongoDB：
//   - K 线历史是追加密集型数据
//   - 按币种/时间范围查询导向
//   - 不属于必须留在 MySQL 中的核心事务状态
//   - 天然适合时序数据存储
//
// 提供的查询方法：
//   - FindBySymbolTime：根据交易对和时间范围查询 K 线
type KlineRepository struct {
	db *mongo.Database
}

// NewKlineRepository 创建 KlineRepository 实例。
//
// 参数：
//   - db：MongoDB 数据库实例
func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// FindBySymbolTime 加载指定时间范围内的 K 线数据。
//
// 查询规则：
//   - 集合名由 symbol 和 period 动态计算
//   - 时间范围查询：time >= from AND time <= to
//   - 支持升序或降序排序
//
// 参数：
//   - ctx：请求上下文
//   - symbol：交易对标识，如 "BTCUSDT"
//   - period：K 线周期，如 "1H"、"1D"、"15m"
//   - from：开始时间（毫秒时间戳）
//   - to：结束时间（毫秒时间戳）
//   - sortOrder：排序方向
//     - "asc"：按时间升序（用于 K 线图表展示）
//     - 其他：按时间降序（用于计算最新数据）
//
// 返回：
//   - []*model.Kline：K 线数据列表（可能为空切片）
//   - error：数据库错误
//
// 注意：sortOrder 参数契约与旧服务匹配，
// 因为市场接口对不同的计算同时依赖升序和降序读取。
func (r *KlineRepository) FindBySymbolTime(
	ctx context.Context,
	symbol string,
	period string,
	from int64,
	to int64,
	sortOrder string,
) ([]*model.Kline, error) {
	mk := &model.Kline{}
	sortValue := int32(-1)
	if sortOrder == "asc" {
		sortValue = 1
	}

	collection := r.db.Collection(mk.TableName(symbol, period))
	cursor, err := collection.Find(
		ctx,
		bson.D{{Key: "time", Value: bson.D{{Key: "$gte", Value: from}, {Key: "$lte", Value: to}}}},
		&options.FindOptions{
			Sort: bson.D{{Key: "time", Value: sortValue}},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("query klines for %s %s: %w", symbol, period, err)
	}
	defer cursor.Close(ctx)

	var list []*model.Kline
	if err := cursor.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode klines for %s %s: %w", symbol, period, err)
	}

	return list, nil
}
