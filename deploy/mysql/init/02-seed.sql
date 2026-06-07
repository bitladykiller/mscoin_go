-- This seed file only inserts the minimum metadata required for the migrated
-- go-zero services to boot, answer market queries, and let `jobcenter` sync
-- OKX K-lines for visible pairs.

USE market;

INSERT INTO coin (
  id, name, can_auto_withdraw, can_recharge, can_transfer, can_withdraw,
  cny_rate, enable_rpc, is_platform_coin, max_tx_fee, max_withdraw_amount,
  min_tx_fee, min_withdraw_amount, name_cn, sort, status, unit, usd_rate,
  withdraw_threshold, has_legal, cold_wallet_address, miner_fee,
  withdraw_scale, account_type, deposit_address, infolink, information,
  min_recharge_amount
) VALUES
  (
    1, 'Bitcoin', 1, 1, 1, 1,
    500000.00000000, 1, 0, 0.01000000, 100.00000000,
    0.00050000, 0.00100000, '比特币', 1, 1, 'BTC', 68000.00000000,
    0.00010000, 0, '', 0.00050000,
    8, 0, '', 'https://bitcoin.org', 'Bitcoin testnet coin metadata used by the refactored demo environment.',
    0.00010000
  ),
  (
    2, 'Tether', 1, 1, 1, 1,
    7.00000000, 1, 1, 10.00000000, 1000000.00000000,
    0.00000000, 1.00000000, '泰达币', 2, 1, 'USDT', 1.00000000,
    1.00000000, 1, '', 0.00000000,
    8, 0, '', 'https://tether.to', 'USDT acts as the settlement coin for the migrated market pairs.',
    1.00000000
  ),
  (
    3, 'Ethereum', 1, 1, 1, 1,
    25000.00000000, 1, 0, 0.10000000, 1000.00000000,
    0.00100000, 0.01000000, '以太坊', 3, 1, 'ETH', 3500.00000000,
    0.00100000, 0, '', 0.00100000,
    8, 0, '', 'https://ethereum.org', 'ETH metadata keeps the visible market list large enough for K-line sync verification.',
    0.00100000
  )
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  cny_rate = VALUES(cny_rate),
  usd_rate = VALUES(usd_rate),
  status = VALUES(status),
  information = VALUES(information);

INSERT INTO exchange_coin (
  id, symbol, base_coin_scale, base_symbol, coin_scale, coin_symbol, enable,
  fee, sort, enable_market_buy, enable_market_sell, min_sell_price, flag,
  max_trading_order, max_trading_time, min_turnover, clear_time, end_time,
  exchangeable, max_buy_price, max_volume, min_volume, publish_amount,
  publish_price, publish_type, robot_type, start_time, visible, zone
) VALUES
  (
    1, 'BTC/USDT', 8, 'USDT', 8, 'BTC', 1,
    0.00100000, 1, 1, 1, 0.00000000, 0,
    100, 0, 10.00000000, 0, 0,
    1, 1000000.000000000000, 1000.000000000000000000, 0.000100000000000000,
    0.000000000000000000, 0.000000000000000000, 0, 0, 0, 1, 0
  ),
  (
    2, 'ETH/USDT', 8, 'USDT', 8, 'ETH', 1,
    0.00100000, 2, 1, 1, 0.00000000, 0,
    100, 0, 10.00000000, 0, 0,
    1, 1000000.000000000000, 10000.000000000000000000, 0.000100000000000000,
    0.000000000000000000, 0.000000000000000000, 0, 0, 0, 1, 0
  )
ON DUPLICATE KEY UPDATE
  base_symbol = VALUES(base_symbol),
  coin_symbol = VALUES(coin_symbol),
  enable = VALUES(enable),
  visible = VALUES(visible);

