// Package mysqlx 集中管理 MySQL 事务编排辅助工具。
package mysqlx

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ExtContext 别名 `sqlx` 执行上下文联合类型，以便仓库方法可以接受 `*sqlx.DB` 或 `*sqlx.Tx`，
// 而无需重复签名。
type ExtContext = sqlx.ExtContext

// TxManager 为重构后的服务标准化事务边界。
//
// 为什么需要这个接口：
//   - MSCoin 的写端工作流通常需要在一个原子单元中协调多个仓库
//   - 单元测试应该能够用确定性的 fake 替换真实的数据库事务运行器
//   - 将事务入口点保留在共享基础设施中有助于每个服务遵循相同的 `go-zero` 分层风格
type TxManager interface {
	WithinTx(ctx context.Context, fn func(exec ExtContext) error) error
}

type txManager struct {
	db *sqlx.DB
}

// NewTxManager 构建 RPC 服务使用的默认 SQL 事务运行器。
func NewTxManager(db *sqlx.DB) TxManager {
	return &txManager{db: db}
}

// WithinTx 在一个 SQL 事务中执行回调。
//
// 回调接收事务执行器抽象而非具体的 `*sqlx.Tx`，以便仓库可以专注于 SQL 语句，
// 而非事务生命周期管理。
func (m *txManager) WithinTx(ctx context.Context, fn func(exec ExtContext) error) error {
	if m == nil || m.db == nil {
		return fmt.Errorf("transaction manager is not initialized")
	}

	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}
