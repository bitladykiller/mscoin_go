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
type KlineRepository struct {
	db *mongo.Database
}

func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// FindBySymbolTime 加载指定时间范围内的 K 线数据。
// `sortOrder` 契约有意与旧服务匹配，因为市场接口对不同的计算同时依赖升序和降序读取。
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
