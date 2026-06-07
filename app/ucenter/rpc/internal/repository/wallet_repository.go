// Package repository 定义钱包仓储层。
//
// WalletRepository 封装会员钱包表（member_wallet）的数据库操作。
// 钱包是会员资产管理的核心，存储各币种的余额信息。
//
// 仓储职责：
//   - 钱包查询：按会员、币种查询钱包
//   - 钱包创建：为新会员或新币种创建钱包
//   - 钱包更新：更新地址、余额等
//   - 余额冻结：提现申请时冻结余额（使用 FOR UPDATE 行锁）
//
// 事务安全设计：
//   - FindByMemberIDAndCoinNameForUpdate：使用 FOR UPDATE 锁定行
//   - FreezeBalance：在事务中原子性地转移余额
//   - 这确保并发提现请求不会导致余额超扣
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
//
// 设计要点：
//   - 支持普通查询和事务查询
//   - 事务查询使用 FOR UPDATE 锁防止并发冲突
//   - 余额操作在 SQL 层原子执行，避免竞态条件
type WalletRepository struct {
	db *sqlx.DB // 数据库连接池
}

// NewWalletRepository 创建钱包仓储实例
// 参数 db 为数据库连接池，由 ServiceContext 提供
func NewWalletRepository(db *sqlx.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// FindByMemberID 根据会员 ID 查询所有钱包
// 用于资产页面展示会员持有的所有币种钱包
//
// 返回：该会员的所有钱包列表
func (r *WalletRepository) FindByMemberID(ctx context.Context, memberID int64) ([]*model.MemberWallet, error) {
	var list []*model.MemberWallet
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_wallet WHERE member_id=?", memberID); err != nil {
		return nil, fmt.Errorf("query member wallets: %w", err)
	}
	return list, nil
}

// FindByMemberIDAndCoinName 根据会员 ID 和币种名称查询钱包
// 用于查询特定币种的钱包信息
//
// 不加行锁，适用于只读查询场景
func (r *WalletRepository) FindByMemberIDAndCoinName(ctx context.Context, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, r.db, memberID, coinName, false)
}

// FindByMemberIDAndCoinNameForUpdate 在现有事务内加载钱包行，
// 并应用 `FOR UPDATE` 锁，以防止并发提现请求同时冻结相同的余额快照。
//
// 使用场景：提现申请流程中的余额冻结
// 在事务中使用，确保：
//   - 查询和后续更新在同一事务中
//   - 其他并发请求等待当前事务完成
//   - 避免余额超扣（两次提现同时使用同一余额）
//
// 参数：
//   - ctx: 请求上下文
//   - exec: 事务执行器，由 TxManager 提供
//   - memberID: 会员 ID
//   - coinName: 币种名称
func (r *WalletRepository) FindByMemberIDAndCoinNameForUpdate(ctx context.Context, exec mysqlx.ExtContext, memberID int64, coinName string) (*model.MemberWallet, error) {
	return r.findByMemberIDAndCoinName(ctx, exec, memberID, coinName, true)
}

// findByMemberIDAndCoinName 根据会员 ID 和币种名称查询钱包（支持行锁）
// 内部方法，封装查询逻辑，支持加锁或不加锁
//
// 参数：
//   - ctx: 请求上下文
//   - exec: 查询执行器（可以是 DB 或事务）
//   - memberID: 会员 ID
//   - coinName: 币种名称
//   - forUpdate: 是否加 FOR UPDATE 锁
func (r *WalletRepository) findByMemberIDAndCoinName(ctx context.Context, exec sqlx.QueryerContext, memberID int64, coinName string, forUpdate bool) (*model.MemberWallet, error) {
	query := "SELECT * FROM member_wallet WHERE member_id=? AND coin_name=? LIMIT 1"
	if forUpdate {
		// FOR UPDATE 锁定行，防止并发修改
		// 在事务中使用，事务提交或回滚后释放锁
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
// 用于充值监听服务，获取需要监听的充值地址列表
//
// 参数：
//   - ctx: 请求上下文
//   - coinName: 币种名称
//
// 返回：该币种下所有会员的充值地址列表
func (r *WalletRepository) FindAllAddress(ctx context.Context, coinName string) ([]string, error) {
	var list []string
	if err := r.db.SelectContext(ctx, &list, "SELECT address FROM member_wallet WHERE coin_name=?", coinName); err != nil {
		return nil, fmt.Errorf("query wallet addresses: %w", err)
	}
	return list, nil
}

// UpdateAddress 更新钱包地址
// 用于为会员分配充值地址（如 BTC 地址）
//
// 参数 wallet 必须包含：
//   - Id: 钱包 ID
//   - Address: 新的充值地址
//   - AddressPrivateKey: 地址私钥（BTC 已废弃，由 Bitcoin Core 管理）
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
//
// 使用场景：提现申请流程
// 在事务中调用，确保：
//   - 余额减少和冻结余额增加在同一原子操作中
//   - 不会出现余额减少但冻结余额未增加的情况
//
// 参数：
//   - ctx: 请求上下文
//   - exec: 事务执行器
//   - memberID: 会员 ID
//   - coinName: 币种名称
//   - amount: 冻结金额
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
// 用于为新会员或新币种创建钱包
func (r *WalletRepository) Save(ctx context.Context, wallet *model.MemberWallet) error {
	return r.save(ctx, r.db, wallet)
}

// save 保存钱包记录（支持事务执行器）
// 内部方法，支持在事务中创建钱包
//
// 参数：
//   - ctx: 请求上下文
//   - exec: 执行器（可以是 DB 或事务）
//   - wallet: 钱包对象
//
// 创建后会回填自增 ID 到 wallet.Id
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
	// 新创建的钱包需要知道自己的 ID，用于后续更新操作
	if wallet != nil {
		if insertedID, lastInsertErr := result.LastInsertId(); lastInsertErr == nil {
			wallet.Id = insertedID
		}
	}
	return nil
}