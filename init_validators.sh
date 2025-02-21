#!/bin/bash

# Exit on error
set -e

# Check arguments
if [ "$#" -lt 4 ]; then
    echo "Usage: $0 <validator0_ip> <validator1_ip> <validator2_ip> <validator3_ip>"
    echo "Example: $0 1.2.3.4 5.6.7.8 9.10.11.12 13.14.15.16"
    exit 1
fi

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
    echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
    exit 1
}

# Set number of validators
NUM_VALIDATORS=4

# Store validator IPs in array
declare -a VALIDATOR_IPS=($1 $2 $3 $4)
echo "All validator IPs: ${VALIDATOR_IPS[@]}"
echo "Number of validators: $NUM_VALIDATORS"

# Configuration
CHAIN_ID="hetu_560000-1"
KEYRING="test"
KEYALGO="eth_secp256k1"
DENOM="ahetu"
HOME_PREFIX="$HOME/.hetud"
# Set balance and stake amounts (matching local_node.sh exactly)
GENESIS_BALANCE="1000000000000000000000000000" # 1 million hetu
GENTX_STAKE="1000000000000000000000000"        # 1 million hetu (1000000000000000000000000 = 10^24)
BASEFEE=1000000000

# Port configuration
P2P_PORT=26656
RPC_PORT=26657 # Must be different from P2P_PORT
API_PORT=1317
GRPC_PORT=9090
GRPC_WEB_PORT=9092
JSON_RPC_PORT=8545
WS_PORT=8546

# Clean up all existing data locally and remotely
echo "Cleaning up all existing data..."

# Clean local node data
rm -rf "${HOME_PREFIX}"/*
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    echo "Cleaning up local validator $i data..."
    rm -rf "${HOME}/.hetud$i"
done

# Clean remote node data (skip any IP matching primary)
PRIMARY_IP=${VALIDATOR_IPS[0]}
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    TARGET_IP=${VALIDATOR_IPS[$i]}
    if [ "$TARGET_IP" = "$PRIMARY_IP" ]; then
        echo "Skipping IP $TARGET_IP since it matches primary node"
        continue
    fi
    echo "Cleaning up data on $TARGET_IP..."
    ssh root@${TARGET_IP} 'rm -rf ~/.hetud ~/.hetud* 2>/dev/null || true'
done

# Initialize primary node
echo "Initializing primary node..."
hetud init "node0" -o --chain-id="${CHAIN_ID}" --home "${HOME_PREFIX}"

# Path variables
GENESIS="${HOME_PREFIX}/config/genesis.json"
TMP_GENESIS="${HOME_PREFIX}/config/tmp_genesis.json"

# Change parameter token denominations to ahetu
jq --arg denom "$DENOM" '.app_state["staking"]["params"]["bond_denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg denom "$DENOM" '.app_state["gov"]["params"]["min_deposit"][0]["denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg denom "$DENOM" '.app_state["gov"]["params"]["min_deposit"][0]["denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg denom "$DENOM" '.app_state["inflation"]["params"]["mint_denom"]=$denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set gas limit in genesis
jq '.consensus_params["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set base fee in genesis
jq --arg fee "$BASEFEE" '.app_state["feemarket"]["params"]["base_fee"]=$fee' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Change proposal periods to pass within a reasonable time
sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$GENESIS"
sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$GENESIS"
sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "15s"/g' "$GENESIS"

# Create gentx directory in primary node
mkdir -p "${HOME_PREFIX}/config/gentx"

# Create validator keys and add genesis accounts
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    echo "Creating validator $i key..."
    hetud keys add "validator$i" \
        --keyring-backend="${KEYRING}" \
        --algo="${KEYALGO}" \
        --home "${HOME_PREFIX}"

    echo "Adding genesis account for validator $i..."
    hetud add-genesis-account "validator$i" "${GENESIS_BALANCE}${DENOM}" \
        --keyring-backend="${KEYRING}" \
        --home "${HOME_PREFIX}"
done

# Create clone directories, gentx, and get node IDs
declare -a NODE_IDS
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    CLONE_HOME="${HOME}/.hetud$i"
    echo "Creating gentx for validator $i in ${CLONE_HOME}..."

    # Initialize fresh node
    rm -rf "${CLONE_HOME}"
    hetud init "node$i" --chain-id="${CHAIN_ID}" --home "${CLONE_HOME}" >/dev/null 2>&1

    # Get and store node ID early
    NODE_IDS[$i]=$(hetud tendermint show-node-id --home "${CLONE_HOME}")
    echo "Node $i ID: ${NODE_IDS[$i]}"

    # Copy necessary files from primary node
    cp "${HOME_PREFIX}/config/genesis.json" "${CLONE_HOME}/config/"
    cp -r "${HOME_PREFIX}/keyring-test" "${CLONE_HOME}/" 2>/dev/null || true
    mkdir -p "${CLONE_HOME}/config/gentx"

    # Set pruning to nothing for archive mode and configure settings
    APP_TOML="${CLONE_HOME}/config/app.toml"
    CONFIG_TOML="${CLONE_HOME}/config/config.toml"

    # Archive mode settings
    sed -i.bak 's/^pruning = "default"/pruning = "nothing"/' "$APP_TOML"

    # Configure external access in config.toml
    # Update RPC and P2P ports
    sed -i.bak -e '/^\[rpc\]/,/^\[/s|^laddr *= *.*|laddr = "tcp://0.0.0.0:26657"|' "$CONFIG_TOML"
    sed -i.bak -e '/^\[p2p\]/,/^\[/s|^laddr *= *.*|laddr = "tcp://0.0.0.0:26656"|' "$CONFIG_TOML"

    # Update other settings
    sed -i.bak \
        -e "s/^moniker *=.*/moniker = \"node${i}\"/" \
        -e "s/^proxy_app *=.*/proxy_app = \"tcp:\/\/127.0.0.1:26658\"/" \
        -e "s/^allow_duplicate_ip *=.*/allow_duplicate_ip = true/" \
        -e "s/^#allow_duplicate_ip *=.*/allow_duplicate_ip = true/" \
        "$CONFIG_TOML"

    # Set minimum gas price to 0
    sed -i.bak "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0${DENOM}\"/g" "$APP_TOML"

    # Configure API and EVM settings in app.toml
    sed -i.bak \
        -e "/^\[api\]/,/^\[/s|^address *= *.*|address = \"tcp://0.0.0.0:${API_PORT}\"|" \
        -e "/^\[grpc\]/,/^\[/s|^address *= *.*|address = \"0.0.0.0:${GRPC_PORT}\"|" \
        -e "/^\[grpc-web\]/,/^\[/s|^address *= *.*|address = \"0.0.0.0:${GRPC_WEB_PORT}\"|" \
        -e "/^\[json-rpc\]/,/^\[/s|^address *= *.*|address = \"0.0.0.0:${JSON_RPC_PORT}\"|" \
        -e "/^\[json-rpc\]/,/^\[/s|^ws-address *= *.*|ws-address = \"0.0.0.0:${WS_PORT}\"|" \
        -e "/^\[json-rpc\]/,/^\[/s|^enable *= *.*|enable = true|" \
        -e "/^\[json-rpc\]/,/^\[/s|^api *= *.*|api = \"eth,net,web3,txpool,debug\"|" \
        -e 's/^json-rpc.enable-indexer = .*$/json-rpc.enable-indexer = true/' \
        -e 's/^evm.tracer = .*$/evm.tracer = ""/' \
        "$APP_TOML"

    # Set timeouts for better network stability
    sed -i.bak 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG_TOML"
    sed -i.bak 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG_TOML"

    # Use the corresponding validator IP
    PUBLIC_IP=${VALIDATOR_IPS[$i]}

    # Create gentx
    hetud gentx "validator$i" \
        "${GENTX_STAKE}${DENOM}" \
        --chain-id="${CHAIN_ID}" \
        --moniker="node$i" \
        --commission-rate="0.05" \
        --commission-max-rate="0.20" \
        --commission-max-change-rate="0.01" \
        --min-self-delegation="1" \
        --ip="${PUBLIC_IP}" \
        --home "${CLONE_HOME}" \
        --keyring-backend="${KEYRING}"

    # Copy gentx back to primary node
    if [ -d "${CLONE_HOME}/config/gentx" ] && [ "$(ls -A "${CLONE_HOME}/config/gentx")" ]; then
        cp "${CLONE_HOME}/config/gentx/"* "${HOME_PREFIX}/config/gentx/"
    else
        echo "Warning: No gentx files found in ${CLONE_HOME}/config/gentx"
    fi

    echo "Gentx created for node $i"
done

# Collect gentxs
echo "Collecting gentxs..."
hetud collect-gentxs --home "${HOME_PREFIX}"

# Validate genesis
echo "Validating genesis..."
hetud validate-genesis --home "${HOME_PREFIX}"

# Configure peers for each validator
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    CLONE_HOME="${HOME}/.hetud$i"
    PEERS=""

    # Build peers string excluding self
    for j in $(seq 0 $((NUM_VALIDATORS - 1))); do
        if [ $i -ne $j ]; then
            if [ ! -z "$PEERS" ]; then
                PEERS="${PEERS},"
            fi
            PEERS="${PEERS}${NODE_IDS[$j]}@${VALIDATOR_IPS[$j]}:${P2P_PORT}"
        fi
    done

    # Configure peers
    echo "Configuring peers for node $i..."
    sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" "${CLONE_HOME}/config/config.toml"
done

# Copy genesis to all validators
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    CLONE_HOME="${HOME}/.hetud$i"
    cp "${HOME_PREFIX}/config/genesis.json" "${CLONE_HOME}/config/"
done

# Copy validator data to target machines (skip any IP matching primary)
PRIMARY_IP=${VALIDATOR_IPS[0]}
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    TARGET_IP=${VALIDATOR_IPS[$i]}
    if [ "$TARGET_IP" = "$PRIMARY_IP" ]; then
        echo "Skipping IP $TARGET_IP since it matches primary node"
        continue
    fi
    echo "Copying validator $i data to $TARGET_IP..."
    # First remove the old directory on remote
    ssh root@${TARGET_IP} "rm -rf ${HOME}/.hetud${i}"
    # Then copy the new data
    rsync -av "${HOME}/.hetud${i}/" "root@${TARGET_IP}:${HOME}/.hetud${i}/"
done

echo "All validators initialized successfully!"
echo "Genesis file location: ${HOME_PREFIX}/config/genesis.json"
echo "Validator information:"
for i in $(seq 0 $((NUM_VALIDATORS - 1))); do
    CLONE_HOME="${HOME}/.hetud$i"
    echo "Validator $i:"
    echo "  Directory: ${CLONE_HOME}"
    echo "  Node ID: ${NODE_IDS[$i]}"
    echo "  IP: ${VALIDATOR_IPS[$i]}"
done
