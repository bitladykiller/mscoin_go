// Package mysqlx 集中管理 MySQL 事务编排辅助工具。
//
// 本包提供简化的事务管理功能，包括：
//   - TxManager 接口：抽象事务管理行为，便于测试
//   - ExtContext 类型别名：统一 DB 和 Tx 的执行上下文
//   - WithinTx 方法：自动处理事务的开始、提交和回滚
//
// 设计理念：
//   - 业务代码只关心事务内的操作，不需要处理事务生命周期
//   - 事务失败自动回滚，成功自动提交
//   - 仓库方法可以接收 ExtContext，从而同时支持 DB 和 Tx
//
// 使用场景：
//   - 跨多个仓库的原子操作
//   - 需要事务保护的复杂业务流程
package mysqlx

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ExtContext 别名 `sqlx` 执行上下文联合类型，以便仓库方法可以接受 `*sqlx.DB` 或 `*sqlx.Tx`，
// 而无需重复签名。
//
// 这是一个关键的设计决策，使得仓库方法可以：
//   - 在非事务场景下接收 *sqlx.DB
//   - 在事务场景下接收 *sqlx.Tx
//   - 两种情况下使用相同的查询代码
//
// 使用示例：
//
//	func (r *userRepository) Create(ctx context.Context, exec ExtContext, user *User) error {
//	    _, err := exec.ExecContext(ctx, "INSERT INTO users (...) VALUES (...)", ...)
//	    return err
//	}
type ExtContext = sqlx.ExtContext

// TxManager 为重构后的服务标准化事务边界。
//
// 为什么需要这个接口：
//   - MSCoin 的写端工作流通常需要在一个原子单元中协调多个仓库
//   - 单元测试应该能够用确定性的 fake 替换真实的数据库事务运行器
//   - 将事务入口点保留在共享基础设施中有助于每个服务遵循相同的 `go-zero` 分层风格
//
// 使用示例：
//
//	err := txManager.WithinTx(ctx, func(exec ExtContext) error {
//	    if err := userRepo.Create(ctx, exec, user); err != nil {
//	        return err
//	    }
//	    if err := walletRepo.Create(ctx, exec, wallet); err != nil {
//	        return err
//	    }
//	    return nil
//	})
type TxManager interface {
	// WithinTx 在一个事务中执行回调函数。
	//
	// 行为说明：
	//   - 回调返回 nil 时，事务自动提交
	//   - 回调返回错误时，事务自动回滚
	//   - 事务开始失败时，返回错误而不执行回调
	//
	// 参数：
	//   - ctx: 上下文，用于超时控制
	//   - fn: 事务内执行的业务逻辑函数，接收 ExtContext 作为执行器
	//
	// 返回值：
	//   - error: 事务失败或回调返回的错误
	WithinTx(ctx context.Context, fn func(exec ExtContext) error) error
}

// txManager 是 TxManager 的默认实现。
type txManager struct {
	db *sqlx.DB
}

// NewTxManager 构建 RPC 服务使用的默认 SQL 事务运行器。
//
// 参数：
//   - db: sqlx 数据库连接实例
//
// 返回值：
//   - TxManager: 事务管理器实例
//
// 使用示例：
//
//	db, _ := mysqlx.New(cfg)
//	txManager := mysqlx.NewTxManager(db)
func NewTxManager(db *sqlx.DB) TxManager {
	return &txManager{db: db}
}

// WithinTx 在一个 SQL 事务中执行回调。
//
// 回调接收事务执行器抽象而非具体的 `*sqlx.Tx`，以便仓库可以专注于 SQL 语句，
// 而非事务生命周期管理。
//
// 事务处理流程：
//  1. 开始事务
//  2. 执行回调
//  3. 回调成功 -> 提交事务
//  4. 回调失败 -> 回滚事务
//  5. defer 确保任何情况下都能回滚未提交的事务
//
// 参数：
//   - ctx: 上下文，用于超时控制
//   - fn: 事务内执行的业务逻辑
//
// 返回值：
//   - error: 事务失败或回调返回的错误
func (m *txManager) WithinTx(ctx context.Context, fn func(exec ExtContext) error) error {
	// 验证管理器已初始化
	if m == nil || m.db == nil {
		return fmt.Errorf("transaction manager is not initialized")
	}

	// 开始事务
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// 使用 defer 确保事务在异常情况下回滚
	// committed 标志用于区分正常提交和需要回滚的情况
	committed := false
	defer func() {
		if !committed {
			// 忽略回滚错误，因为事务可能已经失败
			_ = tx.Rollback()
		}
	}()

	// 执行业务逻辑
	if err := fn(tx); err != nil {
		// 业务逻辑返回错误，事务将由 defer 回滚
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// 标记已提交，防止 defer 回滚
	committed = true
	return nil
}