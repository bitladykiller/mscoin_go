package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/market/rpc/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// KlineRepository encapsulates MongoDB access for historical candle data.
type KlineRepository struct {
	db *mongo.Database
}

func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// FindBySymbolTime loads candles in a concrete time range. The `sortOrder`
// contract intentionally matches the old service because market endpoints depend
// on both ascending and descending reads for different calculations.
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
