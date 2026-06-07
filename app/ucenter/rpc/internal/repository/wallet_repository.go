package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
)

type WalletRepository struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error) {
	var list []*model.MemberWallet
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_wallet WHERE member_id=?", memberID); err != nil {
		return nil, fmt.Errorf("query member wallets: %w", err)
	}
	return list, nil
}

func (r *WalletRepository) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, r.db, memberID, coinName, false)
}

// FindByMemberIDAndCoinNameForUpdate loads the wallet row inside an existing
// transaction and applies `FOR UPDATE` so concurrent withdraw requests cannot
// freeze the same balance snapshot simultaneously.
func (r *WalletRepository) FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, exec, memberID, coinName, true)
}

func (r *WalletRepository) findByMemberIDAndCoinName(ctx context.Context, exec sqlx.QueryerContext, memberID int64, coinName string, forUpdate bool) (*model.MemberWallet, error) {
	query := "SELECT * FROM member_wallet WHERE member_id=? AND coin_name=? LIMIT 1"
	if forUpdate {
		query += " FOR UPDATE"
	}

	var wallet model.MemberWallet
	err := sqlx.GetContext(ctx, exec, &wallet, query, memberID, coinName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query wallet by symbol: %w", err)
	}
	return &wallet, nil
}

func (r *WalletRepository) FindAllAddress(ctx context.Context, coinName string) ([]string, error) {
	var list []string
	if err := r.db.SelectContext(ctx, &list, "SELECT address FROM member_wallet WHERE coin_name=?", coinName); err != nil {
		return nil, fmt.Errorf("query wallet addresses: %w", err)
	}
	return list, nil
}

func (r *WalletRepository) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	if wallet == nil {
		return errors.New("wallet is nil")
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE member_wallet SET address=?, address_private_key=? WHERE id=?", wallet.Address, wallet.AddressPrivateKey, wallet.Id); err != nil {
		return fmt.Errorf("update wallet address: %w", err)
	}
	return nil
}

// FreezeBalance moves funds from available balance to frozen balance.
//
// Why the update stays in SQL instead of manipulating the struct in Go:
//   - this keeps the mutation atomic inside MySQL
//   - the database remains the single source of truth for monetary values
//   - later compensation workers can rely on the persisted frozen amount
func (r *WalletRepository) FreezeBalance(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string, amount float64) error {
	if exec == nil {
		return errors.New("sql executor is nil")
	}
	if _, err := exec.ExecContext(
		ctx,
		"UPDATE member_wallet SET balance=balance-?, frozen_balance=frozen_balance+? WHERE member_id=? AND coin_name=?",
		amount,
		amount,
		memberID,
		coinName,
	); err != nil {
		return fmt.Errorf("freeze wallet balance: %w", err)
	}
	return nil
}

func (r *WalletRepository) Save(ctx context.Context, wallet *model.MemberWallet) error {
	return r.save(ctx, r.db, wallet)
}

func (r *WalletRepository) save(ctx context.Context, exec mysqlx.ExtContext, wallet *model.MemberWallet) error {
	if exec == nil {
		return errors.New("sql executor is nil")
	}
	query := `INSERT INTO member_wallet (
		address, balance, frozen_balance, release_balance, is_lock, member_id,
		version, coin_id, to_released, coin_name, address_private_key
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := exec.ExecContext(
		ctx,
		query,
		wallet.Address,
		wallet.Balance,
		wallet.FrozenBalance,
		wallet.ReleaseBalance,
		wallet.IsLock,
		wallet.MemberId,
		wallet.Version,
		wallet.CoinId,
		wallet.ToReleased,
		wallet.CoinName,
		wallet.AddressPrivateKey,
	)
	if err != nil {
		return fmt.Errorf("insert member wallet: %w", err)
	}
	if wallet != nil {
		if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
			wallet.Id = insertedID
		}
	}
	return nil
}
