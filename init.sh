#!/bin/bash

# Set environment variables
KEY="dev0"
CHAINID="hetu_560000-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
LOGLEVEL2="debug"
TRACE=""
HETUD_HOME="$HOME/.hetud"
ETHCONFIG="$HETUD_HOME/config/config.toml"
GENESIS="$HETUD_HOME/config/genesis.json"
TMPGENESIS="$HETUD_HOME/config/tmp_genesis.json"

echo "Home directory: $HETUD_HOME"

# Build the binary
# go build ./cmd/hetud

# Clear the home folder
rm -rf "$HETUD_HOME"

# Initialize configuration
# hetud config keyring-backend "$KEYRING"
# hetud config chain-id "$CHAINID"

# Add a key
hetud keys add "$KEY" --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$HETUD_HOME"

# Initialize the node
hetud init "$MONIKER" --chain-id "$CHAINID" --home "$HETUD_HOME"

# Modify parameters in the genesis file
jq ".app_state[\"staking\"][\"params\"][\"bond_denom\"]=\"ahetu\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq ".app_state[\"crisis\"][\"constant_fee\"][\"denom\"]=\"ahetu\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq ".app_state[\"gov\"][\"params\"][\"min_deposit\"][0][\"denom\"]=\"ahetu\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq ".app_state[\"mint\"][\"params\"][\"mint_denom\"]=\"ahetu\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# 注意：以下模块现在会自动使用 DefaultGenesisState() 中的默认值
# - blockinflation 模块：自动使用 DefaultParams() 和零值状态
# - event 模块：自动使用空数组作为默认状态 (subnets: [], validator_stakes: [], delegations: [], validator_weights: [])
# Set blockinflation parameters directly in app_state
jq '.app_state.blockinflation.params = {"enable_block_inflation": true, "mint_denom": "ahetu", "total_supply": "21000000000000000000000000", "default_block_emission": "1000000000000000000", "subnet_reward_base": "0.100000000000000000", "subnet_reward_k": "0.100000000000000000", "subnet_reward_max_ratio": "0.900000000000000000", "subnet_moving_alpha": "0.000003000000000000", "subnet_owner_cut": "0.180000000000000000"}' "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Set params module for blockinflation parameters
# jq '.app_state.params = {"subspaces": {"blockinflation": {"key_table": {"params": [{"key": "EnableBlockInflation", "value": true}, {"key": "MintDenom", "value": "ahetu"}, {"key": "TotalSupply", "value": "21000000000000000000000000"}, {"key": "DefaultBlockEmission", "value": "1000000000000000000"}, {"key": "SubnetRewardBase", "value": "0.100000000000000000"}, {"key": "SubnetRewardK", "value": "0.100000000000000000"}, {"key": "SubnetRewardMaxRatio", "value": "0.500000000000000000"}, {"key": "SubnetMovingAlpha", "value": "0.000003000000000000"}, {"key": "SubnetOwnerCut", "value": "0.180000000000000000"}]}}}}' "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Set blockinflation genesis state
jq ".app_state[\"blockinflation\"][\"total_issuance\"]={\"denom\":\"ahetu\",\"amount\":\"0\"}" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq ".app_state[\"blockinflation\"][\"total_burned\"]={\"denom\":\"ahetu\",\"amount\":\"0\"}" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq ".app_state[\"blockinflation\"][\"pending_subnet_rewards\"]={\"denom\":\"ahetu\",\"amount\":\"0\"}" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

jq '.app_state.event = {"subnets": [], "validator_stakes": [], "delegations": [], "validator_weights": []}' "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Set feemarket parameters to match startup gas prices
jq '.app_state.feemarket.params = {"no_base_fee": false, "base_fee_change_denominator": 8, "elasticity_multiplier": 2, "enable_height": "0", "base_fee": "1000000000", "min_gas_price": "0.000100000000000000", "min_gas_multiplier": "0.500000000000000000"}' "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"
jq '.app_state.feemarket.block_gas = "0"' "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Increase block time
jq ".consensus_params[\"block\"][\"time_iota_ms\"]=\"30000\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Set the maximum gas limit for blocks
jq ".consensus_params[\"block\"][\"max_gas\"]=\"10000000\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Modify the configuration file
sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' "$ETHCONFIG"

# Allocate genesis accounts
hetud add-genesis-account "$KEY" 100000000000000000000000000ahetu --keyring-backend "$KEYRING" --home "$HETUD_HOME"

# Sign genesis transaction
hetud gentx "$KEY" 2000000000000000000000ahetu --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$HETUD_HOME" --fees 1000000ahetu --gas 200000

# Collect genesis transactions
hetud collect-gentxs

# Validate genesis file
hetud validate-genesis

# Configure P2P for external connections
# Get the node ID
NODE_ID=$(hetud tendermint show-node-id --home "$HETUD_HOME")
echo "Node ID: $NODE_ID"

# Configure external address (replace YOUR_IP with your actual IP)
# sed -i '' "s/external_address = \"\"/external_address = \"YOUR_IP:26656\"/g" "$ETHCONFIG"

# Start the node
hetud start --pruning=nothing $TRACE --log_level "$LOGLEVEL2" --minimum-gas-prices=0.0001ahetu --home "$HETUD_HOME" --chain-id "$CHAINID"