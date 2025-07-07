#!/bin/bash

# Set environment variables
KEY="dev0"
CHAINID="hetu_560000-1"
MONIKER="localtestnet"
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
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

# Increase block time
jq ".consensus_params[\"block\"][\"time_iota_ms\"]=\"30000\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Set the maximum gas limit for blocks
jq ".consensus_params[\"block\"][\"max_gas\"]=\"10000000\"" "$GENESIS" > "$TMPGENESIS" && mv "$TMPGENESIS" "$GENESIS"

# Modify the configuration file
sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' "$ETHCONFIG"

# Allocate genesis accounts
hetud add-genesis-account "$KEY" 100000000000000000000000000ahetu --keyring-backend "$KEYRING" --home "$HETUD_HOME"

# Sign genesis transaction
hetud gentx "$KEY" 1000000000000000000000ahetu --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$HETUD_HOME"

# Collect genesis transactions
hetud collect-gentxs

# Validate genesis file
hetud validate-genesis

# Start the node
hetud start --pruning=nothing $TRACE --log_level "$LOGLEVEL" --minimum-gas-prices=0.0001ahetu --home "$HETUD_HOME" --chain-id "$CHAINID"