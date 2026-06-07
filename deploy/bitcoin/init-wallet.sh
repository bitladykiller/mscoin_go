#!/bin/sh
set -eu

# This script creates or loads the dedicated `mscoin` Bitcoin Core wallet used
# by both `ucenter-rpc` and `jobcenter`.
#
# Why this extra step exists:
# - `ucenter-rpc` now allocates BTC deposit addresses through Bitcoin Core
# - `jobcenter` signs withdraw transactions with `signrawtransactionwithwallet`
# - both operations must target the same node-managed wallet, otherwise the
#   withdraw flow would fail at runtime even though all containers appear healthy

RPC_ARGS="-testnet -rpcconnect=bitcoin -rpcport=18332 -rpcuser=bitcoin -rpcpassword=123456"

echo "waiting for bitcoin rpc ..."
until bitcoin-cli ${RPC_ARGS} getblockchaininfo >/dev/null 2>&1; do
  sleep 2
done

if bitcoin-cli ${RPC_ARGS} listwallets | grep -q '"mscoin"'; then
  echo "wallet mscoin already loaded"
  exit 0
fi

if bitcoin-cli ${RPC_ARGS} loadwallet "mscoin" >/dev/null 2>&1; then
  echo "wallet mscoin loaded"
  exit 0
fi

echo "creating wallet mscoin"
bitcoin-cli -named ${RPC_ARGS} createwallet wallet_name="mscoin" descriptors=true load_on_startup=true >/dev/null

# Bootstrap one legacy address so the wallet is immediately ready for
# reset-address flows after the initialization container completes.
bitcoin-cli ${RPC_ARGS} -rpcwallet=mscoin getnewaddress "bootstrap" legacy >/dev/null

echo "wallet mscoin is ready"
