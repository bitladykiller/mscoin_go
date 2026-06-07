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

// WalletRepository 钱包仓储
// 负责钱包数据的持久化操作
type WalletRepository struct {
	db *sqlx.DB
}

// NewWalletRepository 创建钱包仓储实例
func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// FindByMemberID 根据会员 ID 查询所有钱包
func (r *WalletRepository) FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error) {
	var list []*model.MemberWallet
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_wallet WHERE member_id=?", memberID); err != nil {
		return nil, fmt.Errorf("query member wallets: %w", err)
	}
	return list, nil
}

// FindByMemberIDAndCoinName 根据会员 ID 和币种名称查询钱包
func (r *WalletRepository) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, r.db, memberID, coinName, false)
}

// FindByMemberIDAndCoinNameForUpdate 在现有事务内加载钱包行，
// 并应用 `FOR UPDATE` 锁，以防止并发提现请求同时冻结相同的余额快照。
func (r *WalletRepository) FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, exec, memberID, coinName, true)
}

// findByMemberIDAndCoinName 根据会员 ID 和币种名称查询钱包（支持行锁）
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

// FindAllAddress 获取指定币种的所有钱包地址
func (r *WalletRepository) FindAllAddress(ctx context.Context, coinName string) ([]string, error) {
	var list []string
	if err := r.db.SelectContext(ctx, &list, "SELECT address FROM member_wallet WHERE coin_name=?", coinName); err != nil {
		return nil, fmt.Errorf("query wallet addresses: %w", err)
	}
	return list, nil
}

// UpdateAddress 更新钱包地址
func (r *WalletRepository) UpdateAddress(ctx context.Context, wallet *model.MemberWallet) error {
	if wallet == nil {
		return errors.New("wallet is nil")
	}
	if _, err := r.db.ExecContext(ctx, "UPDATE member_wallet SET address=?, address_private_key=? WHERE id=?", wallet.Address, wallet.AddressPrivateKey, wallet.Id); err != nil {
		return fmt.Errorf("update wallet address: %w", err)
	}
	return nil
}

// FreezeBalance 将资金从可用余额转移至冻结余额。
//
// 为何更新操作保留在 SQL 中而非在 Go 中操作结构体：
//   - 这使得变更在 MySQL 中保持原子性
//   - 数据库仍是金额数值的唯一可信源
//   - 后续补偿工作器可依赖已持久化的冻结金额
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

// Save 保存钱包记录
func (r *WalletRepository) Save(ctx context.Context, wallet *model.MemberWallet) error {
	return r.save(ctx, r.db, wallet)
}

// save 保存钱包记录（支持事务执行器）
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
	// 回填自增 ID
	if wallet != nil {
		if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
			wallet.Id = insertedID
		}
	}
	return nil
}