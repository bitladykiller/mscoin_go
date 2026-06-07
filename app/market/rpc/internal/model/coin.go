// Package model stores the persistence models owned by the market domain.
package model

// Coin mirrors the `coin` table.
//
// The field names intentionally stay close to the existing schema because this
// refactor targets behavior compatibility first. The detailed tags are useful
// both for `sqlx` mapping and for making the source schema readable when
// business logic depends on specific columns.
type Coin struct {
	ID                int     `db:"id" gorm:"column:id"`
	Name              string  `db:"name" gorm:"column:name"`
	CanAutoWithdraw   int     `db:"can_auto_withdraw" gorm:"column:can_auto_withdraw"`
	CanRecharge       int     `db:"can_recharge" gorm:"column:can_recharge"`
	CanTransfer       int     `db:"can_transfer" gorm:"column:can_transfer"`
	CanWithdraw       int     `db:"can_withdraw" gorm:"column:can_withdraw"`
	CNYRate           float64 `db:"cny_rate" gorm:"column:cny_rate"`
	EnableRPC         int     `db:"enable_rpc" gorm:"column:enable_rpc"`
	IsPlatformCoin    int     `db:"is_platform_coin" gorm:"column:is_platform_coin"`
	MaxTxFee          float64 `db:"max_tx_fee" gorm:"column:max_tx_fee"`
	MaxWithdrawAmount float64 `db:"max_withdraw_amount" gorm:"column:max_withdraw_amount"`
	MinTxFee          float64 `db:"min_tx_fee" gorm:"column:min_tx_fee"`
	MinWithdrawAmount float64 `db:"min_withdraw_amount" gorm:"column:min_withdraw_amount"`
	NameCN            string  `db:"name_cn" gorm:"column:name_cn"`
	Sort              int     `db:"sort" gorm:"column:sort"`
	Status            int     `db:"status" gorm:"column:status"`
	Unit              string  `db:"unit" gorm:"column:unit"`
	USDTRate          float64 `db:"usd_rate" gorm:"column:usd_rate"`
	WithdrawThreshold float64 `db:"withdraw_threshold" gorm:"column:withdraw_threshold"`
	HasLegal          int     `db:"has_legal" gorm:"column:has_legal"`
	ColdWalletAddress string  `db:"cold_wallet_address" gorm:"column:cold_wallet_address"`
	MinerFee          float64 `db:"miner_fee" gorm:"column:miner_fee"`
	WithdrawScale     int     `db:"withdraw_scale" gorm:"column:withdraw_scale"`
	AccountType       int     `db:"account_type" gorm:"column:account_type"`
	DepositAddress    string  `db:"deposit_address" gorm:"column:deposit_address"`
	InfoLink          string  `db:"infolink" gorm:"column:infolink"`
	Information       string  `db:"information" gorm:"column:information"`
	MinRechargeAmount float64 `db:"min_recharge_amount" gorm:"column:min_recharge_amount"`
}
