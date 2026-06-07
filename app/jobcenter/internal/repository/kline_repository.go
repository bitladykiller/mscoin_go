package repository

import (
	"context"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// KlineRepository owns MongoDB writes for asynchronously synchronized candles.
type KlineRepository struct {
	db *mongo.Database
}

func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// ReplaceBatch deletes the overlapping tail and inserts the refreshed batch.
//
// Why the repository rewrites the tail instead of trying to upsert row by row:
//   - OKX returns the most recent window on each request, so tail overlap is
//     expected
//   - batch delete + insert keeps the logic compact and close to the old
//     service behavior
//   - the collection is append-heavy and not part of the core transactional
//     state, so this rewrite approach is acceptable here
func (r *KlineRepository) ReplaceBatch(ctx context.Context, symbol string, period string, list []*model.Kline) error {
	if r == nil || r.db == nil {
		return errors.New("mongo database is not initialized")
	}
	if len(list) == 0 {
		return nil
	}

	collection := r.db.Collection((&model.Kline{}).TableName(symbol, period))
	cutoff := list[len(list)-1].Time
	if _, err := collection.DeleteMany(ctx, bson.D{{Key: "time", Value: bson.D{{Key: "$gte", Value: cutoff}}}}); err != nil {
		return fmt.Errorf("delete overlapping klines for %s %s: %w", symbol, period, err)
	}

	documents := make([]interface{}, 0, len(list))
	for _, item := range list {
		if item != nil {
			documents = append(documents, item)
		}
	}
	if len(documents) == 0 {
		return nil
	}
	if _, err := collection.InsertMany(ctx, documents); err != nil {
		return fmt.Errorf("insert klines for %s %s: %w", symbol, period, err)
	}
	return nil
}
