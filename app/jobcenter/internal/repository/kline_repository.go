// Package repository 提供数据访问层实现。
package repository

import (
	"context"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// KlineRepository 负责 MongoDB 的 K 线数据写入操作。
//
// 存储策略：
//   - K 线数据存储在 MongoDB 中
//   - 按交易对和周期分集合存储
//   - 采用"删除尾部+批量插入"的更新策略
//
// 为什么选择 MongoDB：
//   - K 线数据是追加/刷新导向的，非事务性状态
//   - 支持灵活的集合命名，便于按交易对分片
//   - 高写入吞吐量，适合定时批量同步场景
type KlineRepository struct {
	// db MongoDB 数据库实例
	db *mongo.Database
}

// NewKlineRepository 创建 K 线数据 Repository 实例。
//
// 参数：
//   - db: MongoDB 数据库实例
//
// 返回：Repository 实例，可直接使用
func NewKlineRepository(db *mongo.Database) *KlineRepository {
	return &KlineRepository{db: db}
}

// ReplaceBatch 删除重叠的尾部数据并插入刷新后的批次数据。
//
// 更新策略（删除尾部+批量插入）：
//   1. 以最新数据的最早时间点为分界
//   2. 删除该时间点之后的所有数据
//   3. 批量插入新获取的数据
//
// 为什么采用重写尾部而非逐行 upsert：
//   - OKX 每次请求返回最近时间窗口的数据，尾部重叠是预期行为
//   - 批量删除 + 插入保持逻辑紧凑，接近旧服务行为
//   - 该集合以追加为主，不属于核心事务状态，重写方式可接受
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//   - symbol: 交易对符号，如 "BTC/USDT"
//   - period: K 线周期，如 "1m"
//   - list: K 线数据列表
//
// 返回：
//   - error: 操作错误，包括 MongoDB 未初始化错误
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
