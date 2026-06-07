// Package mysqlx centralizes MySQL transaction orchestration helpers.
package mysqlx

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ExtContext aliases the `sqlx` execution context union so repository methods
// can accept either `*sqlx.DB` or `*sqlx.Tx` without duplicating signatures.
type ExtContext = sqlx.ExtContext

// TxManager standardizes transaction boundaries for the refactored services.
//
// Why this interface exists:
//   - write-side workflows in MSCoin often need to coordinate multiple
//     repositories in one atomic unit
//   - unit tests should be able to replace the real database transaction runner
//     with a deterministic fake
//   - keeping transaction entry points in shared infrastructure helps every
//     service follow the same `go-zero` layering style
type TxManager interface {
	WithinTx(ctx context.Context, fn func(exec ExtContext) error) error
}

type txManager struct {
	db *sqlx.DB
}

// NewTxManager builds the default SQL transaction runner used by RPC services.
func NewTxManager(db *sqlx.DB) TxManager {
	return &txManager{db: db}
}

// WithinTx executes the callback in one SQL transaction.
//
// The callback receives the transactional executor abstraction instead of the
// concrete `*sqlx.Tx` so repositories can stay focused on SQL statements rather
// than transaction lifecycle management.
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
