package repository

import (
	"context"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// KlineRepository 负责 MongoDB 的异步 K 线写入操作。
type KlineRepository struct {
	db *mongo.Database
}

func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// ReplaceBatch 删除重叠的尾部数据并插入刷新后的批次数据。
//
// 为什么仓库采用重写尾部而非逐行 upsert 的方式：
//   - OKX 每次请求返回最近时间窗口的数据，因此尾部重叠是预期行为
//   - 批量删除 + 插入保持逻辑紧凑，接近旧服务行为
//   - 该集合以追加为主，不属于核心事务状态，因此此重写方式可接受
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
