// Package model 存储 market 领域的持久化模型。
//
// Model 层职责：
//   - 定义数据实体结构
//   - 映射数据库表/集合结构
//   - 提供模型相关的工具方法
//   - 不包含业务逻辑（业务逻辑在 Domain Service 层）
//
// 本包包含三个 Model：
//   - Coin：币种模型（MySQL 表）
//   - ExchangeCoin：交易对模型（MySQL 表）
//   - Kline：K 线模型（MongoDB 集合）
package model

// Coin 映射 `coin` 表，存储币种配置信息。
//
// 字段名有意保持与现有模式接近，因为此次重构以行为兼容性为首要目标。
// 详细的标签对 `sqlx` 映射和当业务逻辑依赖特定列时使源模式可读都很有用。
//
// 业务用途：
//   - 定义系统支持的币种
//   - 控制充值/提现/转账开关
//   - 设置费率和限额
//   - 存储币种基本信息（名称、单位等）
//
// 字段说明：
//   - ID：币种唯一标识（数据库主键）
//   - Name/NameCN：币种英文名和中文名，用于前端展示
//   - Unit：币种单位标识，如 "BTC"、"ETH"、"USDT"，用于 API 查询
//   - CanWithdraw/CanRecharge/CanTransfer：提现、充值、转账功能开关（1=启用）
//   - CanAutoWithdraw：是否支持自动提现（1=启用）
//   - MaxWithdrawAmount/MinWithdrawAmount：单次提现金额上限和下限
//   - MaxTxFee/MinTxFee：手续费上限和下限
//   - WithdrawThreshold：提现阈值（触发审核等）
//   - MinerFee：矿工费
//   - WithdrawScale：提现精度（小数位数）
//   - CNYRate/USDTRate：对 CNY 和 USDT 的汇率（用于资产估值）
//   - Status：币种状态（启用/禁用）
//   - Sort：排序权重（用于币种列表展示顺序）
//   - EnableRPC：是否启用 RPC 调用（1=启用）
//   - IsPlatformCoin：是否为平台币（1=是）
//   - HasLegal：是否有法律合规要求（1=有）
//   - AccountType：账户类型
//   - ColdWalletAddress：冷钱包地址（用于大额资金存储）
//   - DepositAddress：充值地址（用于充值监控）
//   - InfoLink/Information：币种介绍链接和描述
//   - MinRechargeAmount：最小充值金额（低于此金额可能不入账）
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
