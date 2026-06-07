-- ---------------------------------------------------------------------------
-- market schema
-- ---------------------------------------------------------------------------
USE market;

-- The `coin` table is read through `SELECT *` in multiple services. To keep
-- `sqlx` scans stable, this schema intentionally contains only the columns that
-- the refactored models declare.
CREATE TABLE IF NOT EXISTS coin (
  id INT NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL DEFAULT '',
  can_auto_withdraw INT NOT NULL DEFAULT 0,
  can_recharge INT NOT NULL DEFAULT 1,
  can_transfer INT NOT NULL DEFAULT 1,
  can_withdraw INT NOT NULL DEFAULT 1,
  cny_rate DECIMAL(24,8) NOT NULL DEFAULT 0,
  enable_rpc INT NOT NULL DEFAULT 1,
  is_platform_coin INT NOT NULL DEFAULT 0,
  max_tx_fee DECIMAL(24,8) NOT NULL DEFAULT 0,
  max_withdraw_amount DECIMAL(24,8) NOT NULL DEFAULT 0,
  min_tx_fee DECIMAL(24,8) NOT NULL DEFAULT 0,
  min_withdraw_amount DECIMAL(24,8) NOT NULL DEFAULT 0,
  name_cn VARCHAR(64) NOT NULL DEFAULT '',
  sort INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 1,
  unit VARCHAR(32) NOT NULL DEFAULT '',
  usd_rate DECIMAL(24,8) NOT NULL DEFAULT 0,
  withdraw_threshold DECIMAL(24,8) NOT NULL DEFAULT 0,
  has_legal INT NOT NULL DEFAULT 0,
  cold_wallet_address VARCHAR(255) NOT NULL DEFAULT '',
  miner_fee DECIMAL(24,8) NOT NULL DEFAULT 0,
  withdraw_scale INT NOT NULL DEFAULT 8,
  account_type INT NOT NULL DEFAULT 0,
  deposit_address VARCHAR(255) NOT NULL DEFAULT '',
  infolink VARCHAR(255) NOT NULL DEFAULT '',
  information TEXT NOT NULL,
  min_recharge_amount DECIMAL(24,8) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_coin_unit (unit)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS exchange_coin (
  id BIGINT NOT NULL AUTO_INCREMENT,
  symbol VARCHAR(64) NOT NULL DEFAULT '',
  base_coin_scale BIGINT NOT NULL DEFAULT 8,
  base_symbol VARCHAR(32) NOT NULL DEFAULT '',
  coin_scale BIGINT NOT NULL DEFAULT 8,
  coin_symbol VARCHAR(32) NOT NULL DEFAULT '',
  enable BIGINT NOT NULL DEFAULT 1,
  fee DECIMAL(18,8) NOT NULL DEFAULT 0,
  sort BIGINT NOT NULL DEFAULT 0,
  enable_market_buy BIGINT NOT NULL DEFAULT 1,
  enable_market_sell BIGINT NOT NULL DEFAULT 1,
  min_sell_price DECIMAL(36,18) NOT NULL DEFAULT 0,
  flag BIGINT NOT NULL DEFAULT 0,
  max_trading_order BIGINT NOT NULL DEFAULT 0,
  max_trading_time BIGINT NOT NULL DEFAULT 0,
  min_turnover DECIMAL(36,18) NOT NULL DEFAULT 0,
  clear_time BIGINT NOT NULL DEFAULT 0,
  end_time BIGINT NOT NULL DEFAULT 0,
  exchangeable BIGINT NOT NULL DEFAULT 1,
  max_buy_price DECIMAL(36,18) NOT NULL DEFAULT 0,
  max_volume DECIMAL(36,18) NOT NULL DEFAULT 0,
  min_volume DECIMAL(36,18) NOT NULL DEFAULT 0,
  publish_amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  publish_price DECIMAL(36,18) NOT NULL DEFAULT 0,
  publish_type BIGINT NOT NULL DEFAULT 0,
  robot_type BIGINT NOT NULL DEFAULT 0,
  start_time BIGINT NOT NULL DEFAULT 0,
  visible BIGINT NOT NULL DEFAULT 1,
  zone BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_exchange_coin_symbol (symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- ucenter schema
-- ---------------------------------------------------------------------------
USE ucenter;

CREATE TABLE IF NOT EXISTS member (
  id BIGINT NOT NULL AUTO_INCREMENT,
  ali_no VARCHAR(128) NOT NULL DEFAULT '',
  qr_code_url VARCHAR(255) NOT NULL DEFAULT '',
  appeal_success_times BIGINT NOT NULL DEFAULT 0,
  appeal_times BIGINT NOT NULL DEFAULT 0,
  application_time BIGINT NOT NULL DEFAULT 0,
  avatar VARCHAR(255) NOT NULL DEFAULT '',
  bank VARCHAR(128) NOT NULL DEFAULT '',
  branch VARCHAR(128) NOT NULL DEFAULT '',
  card_no VARCHAR(128) NOT NULL DEFAULT '',
  certified_business_apply_time BIGINT NOT NULL DEFAULT 0,
  certified_business_check_time BIGINT NOT NULL DEFAULT 0,
  certified_business_status BIGINT NOT NULL DEFAULT 0,
  channel_id BIGINT NOT NULL DEFAULT 0,
  email VARCHAR(128) NOT NULL DEFAULT '',
  first_level BIGINT NOT NULL DEFAULT 0,
  google_date BIGINT NOT NULL DEFAULT 0,
  google_key VARCHAR(255) NOT NULL DEFAULT '',
  google_state BIGINT NOT NULL DEFAULT 0,
  id_number VARCHAR(128) NOT NULL DEFAULT '',
  inviter_id BIGINT NOT NULL DEFAULT 0,
  is_channel BIGINT NOT NULL DEFAULT 0,
  jy_password VARCHAR(255) NOT NULL DEFAULT '',
  last_login_time BIGINT NOT NULL DEFAULT 0,
  city VARCHAR(64) NOT NULL DEFAULT '',
  country VARCHAR(64) NOT NULL DEFAULT '',
  district VARCHAR(64) NOT NULL DEFAULT '',
  province VARCHAR(64) NOT NULL DEFAULT '',
  login_count BIGINT NOT NULL DEFAULT 0,
  login_lock BIGINT NOT NULL DEFAULT 0,
  margin VARCHAR(128) NOT NULL DEFAULT '',
  member_level BIGINT NOT NULL DEFAULT 0,
  mobile_phone VARCHAR(32) NOT NULL DEFAULT '',
  password VARCHAR(255) NOT NULL DEFAULT '',
  promotion_code VARCHAR(64) NOT NULL DEFAULT '',
  publish_advertise BIGINT NOT NULL DEFAULT 0,
  real_name VARCHAR(64) NOT NULL DEFAULT '',
  real_name_status BIGINT NOT NULL DEFAULT 0,
  registration_time BIGINT NOT NULL DEFAULT 0,
  salt VARCHAR(64) NOT NULL DEFAULT '',
  second_level BIGINT NOT NULL DEFAULT 0,
  sign_in_ability BIGINT NOT NULL DEFAULT 0,
  status BIGINT NOT NULL DEFAULT 0,
  third_level BIGINT NOT NULL DEFAULT 0,
  token VARCHAR(255) NOT NULL DEFAULT '',
  token_expire_time BIGINT NOT NULL DEFAULT 0,
  transaction_status BIGINT NOT NULL DEFAULT 0,
  transaction_time BIGINT NOT NULL DEFAULT 0,
  transactions BIGINT NOT NULL DEFAULT 0,
  username VARCHAR(64) NOT NULL DEFAULT '',
  qr_we_code_url VARCHAR(255) NOT NULL DEFAULT '',
  wechat VARCHAR(128) NOT NULL DEFAULT '',
  local VARCHAR(64) NOT NULL DEFAULT '',
  integration BIGINT NOT NULL DEFAULT 0,
  member_grade_id BIGINT NOT NULL DEFAULT 0,
  kyc_status BIGINT NOT NULL DEFAULT 0,
  generalize_total BIGINT NOT NULL DEFAULT 0,
  inviter_parent_id BIGINT NOT NULL DEFAULT 0,
  super_partner VARCHAR(16) NOT NULL DEFAULT '0',
  kick_fee DOUBLE NOT NULL DEFAULT 0,
  power DOUBLE NOT NULL DEFAULT 0,
  team_level BIGINT NOT NULL DEFAULT 0,
  team_power DOUBLE NOT NULL DEFAULT 0,
  member_level_id BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_member_mobile_phone (mobile_phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS member_wallet (
  id BIGINT NOT NULL AUTO_INCREMENT,
  address VARCHAR(255) NOT NULL DEFAULT '',
  balance DECIMAL(36,18) NOT NULL DEFAULT 0,
  frozen_balance DECIMAL(36,18) NOT NULL DEFAULT 0,
  release_balance DECIMAL(36,18) NOT NULL DEFAULT 0,
  is_lock INT NOT NULL DEFAULT 0,
  member_id BIGINT NOT NULL DEFAULT 0,
  version INT NOT NULL DEFAULT 0,
  coin_id BIGINT NOT NULL DEFAULT 0,
  to_released DECIMAL(36,18) NOT NULL DEFAULT 0,
  coin_name VARCHAR(32) NOT NULL DEFAULT '',
  address_private_key TEXT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_member_wallet_member_coin (member_id, coin_name),
  KEY idx_member_wallet_address (address)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS member_address (
  id BIGINT NOT NULL AUTO_INCREMENT,
  member_id BIGINT NOT NULL DEFAULT 0,
  coin_id BIGINT NOT NULL DEFAULT 0,
  address VARCHAR(255) NOT NULL DEFAULT '',
  remark VARCHAR(255) NOT NULL DEFAULT '',
  status INT NOT NULL DEFAULT 0,
  create_time BIGINT NOT NULL DEFAULT 0,
  delete_time BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_member_address_member_coin (member_id, coin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS member_transaction (
  id BIGINT NOT NULL AUTO_INCREMENT,
  address VARCHAR(255) NOT NULL DEFAULT '',
  amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  create_time BIGINT NOT NULL DEFAULT 0,
  fee DECIMAL(36,18) NOT NULL DEFAULT 0,
  flag INT NOT NULL DEFAULT 0,
  member_id BIGINT NOT NULL DEFAULT 0,
  symbol VARCHAR(64) NOT NULL DEFAULT '',
  `type` INT NOT NULL DEFAULT 0,
  discount_fee VARCHAR(64) NOT NULL DEFAULT '',
  real_fee VARCHAR(64) NOT NULL DEFAULT '',
  PRIMARY KEY (id),
  KEY idx_member_transaction_member_time (member_id, create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS withdraw_record (
  id BIGINT NOT NULL AUTO_INCREMENT,
  member_id BIGINT NOT NULL DEFAULT 0,
  coin_id BIGINT NOT NULL DEFAULT 0,
  total_amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  fee DECIMAL(36,18) NOT NULL DEFAULT 0,
  arrived_amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  address VARCHAR(255) NOT NULL DEFAULT '',
  remark VARCHAR(255) NOT NULL DEFAULT '',
  transaction_number VARCHAR(255) NOT NULL DEFAULT '',
  can_auto_withdraw INT NOT NULL DEFAULT 0,
  isAuto INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 0,
  create_time BIGINT NOT NULL DEFAULT 0,
  deal_time BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_withdraw_record_member_time (member_id, create_time),
  KEY idx_withdraw_record_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ---------------------------------------------------------------------------
-- exchange schema
-- ---------------------------------------------------------------------------
USE exchange;

CREATE TABLE IF NOT EXISTS exchange_order (
  id BIGINT NOT NULL AUTO_INCREMENT,
  order_id VARCHAR(64) NOT NULL DEFAULT '',
  amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  base_symbol VARCHAR(32) NOT NULL DEFAULT '',
  canceled_time BIGINT NOT NULL DEFAULT 0,
  coin_symbol VARCHAR(32) NOT NULL DEFAULT '',
  completed_time BIGINT NOT NULL DEFAULT 0,
  direction INT NOT NULL DEFAULT 0,
  member_id BIGINT NOT NULL DEFAULT 0,
  price DECIMAL(36,18) NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 0,
  symbol VARCHAR(64) NOT NULL DEFAULT '',
  `time` BIGINT NOT NULL DEFAULT 0,
  traded_amount DECIMAL(36,18) NOT NULL DEFAULT 0,
  turnover DECIMAL(36,18) NOT NULL DEFAULT 0,
  `type` INT NOT NULL DEFAULT 0,
  use_discount VARCHAR(8) NOT NULL DEFAULT '0',
  PRIMARY KEY (id),
  UNIQUE KEY uk_exchange_order_order_id (order_id),
  KEY idx_exchange_order_member_symbol_status (member_id, symbol, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

