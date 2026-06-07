// Package repository 定义会员地址仓储层。
//
// MemberAddressRepository 封装会员地址表（member_address）的数据库操作。
// 会员地址表存储会员配置的提现目标地址，便于会员快速选择常用提现地址。
//
// 与 WalletRepository 的区别：
//   - WalletRepository：管理平台控制的钱包（充值地址）
//   - MemberAddressRepository：管理用户配置的提现地址（提现目标）
package repository

import (
	"context"
	"fmt"

	"mscoin_go/app/ucenter/rpc/internal/model"

	"github.com/jmoiron/sqlx"
)

// MemberAddressRepository 拥有提现地址簿表的 SQL 访问能力。
// 提供会员提现地址的查询功能。
type MemberAddressRepository struct {
	db *sqlx.DB // 数据库连接池
}

// NewMemberAddressRepository 创建会员地址仓储实例
// 参数 db 为数据库连接池，由 ServiceContext 提供
func NewMemberAddressRepository(db *sqlx.DB) *MemberAddressRepository {
	return &MemberAddressRepository{db: db}
}

// FindByMemberIDAndCoinID 根据会员 ID 和币种 ID 查询提现地址列表
// 用于提现页面展示会员保存的提现地址
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//   - coinID: 币种 ID
//
// 返回：该会员该币种下的所有提现地址
// 注意：包含已删除的地址（Status 或 DeleteTime 筛选应在上层处理）
func (r *MemberAddressRepository) FindByMemberIDAndCoinID(ctx context.Context, memberID int64, coinID int64) ([]*model.MemberAddress, error) {
	var list []*model.MemberAddress
	if err := r.db.SelectContext(ctx, &list, "SELECT * FROM member_address WHERE member_id=? AND coin_id=?", memberID, coinID); err != nil {
		return nil, fmt.Errorf("query member addresses: %w", err)
	}
	return list, nil
}
