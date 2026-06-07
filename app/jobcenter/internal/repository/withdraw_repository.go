package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/jobcenter/internal/model"

	"github.com/jmoiron/sqlx"
)

// WithdrawRepository encapsulates the direct SQL that jobcenter needs for
// withdraw status finalization.
//
// Why jobcenter touches this table directly in the current migration phase:
//   - protobuf regeneration is intentionally avoided for now because the repo
//     pins an older grpc toolchain
//   - the withdraw table is still the source of truth for async execution state
//   - updating one narrow repository keeps the cross-service coupling explicit
//     and localized until a dedicated RPC contract is introduced later
type WithdrawRepository struct {
	db *sqlx.DB
}

func NewWithdrawRepository(db *sqlx.DB) *WithdrawRepository {
	return &WithdrawRepository{db: db}
}

func (r *WithdrawRepository) FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error) {
	var record model.WithdrawRecord
	err := r.db.GetContext(ctx, &record, "SELECT * FROM withdraw_record WHERE id=? LIMIT 1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query withdraw record: %w", err)
	}
	return &record, nil
}

// MarkSuccess persists the final txid and success state.
//
// The update is guarded by the current status so duplicate Kafka deliveries or
// manual back-office corrections do not overwrite a record that already moved
// out of the processing state.
func (r *WithdrawRepository) MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE withdraw_record SET transaction_number=?, status=?, deal_time=? WHERE id=? AND status=?",
		txID,
		model.WithdrawStatusSuccess,
		dealTime,
		id,
		model.WithdrawStatusProcessing,
	)
	if err != nil {
		return false, fmt.Errorf("update withdraw record success: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read withdraw record update rows: %w", err)
	}
	return rowsAffected > 0, nil
}
