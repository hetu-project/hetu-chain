#!/bin/bash

KEY="dev0"
CHAINID="hetu_560000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t hetu-datadir.XXXXX)

echo "create and add new keys"
./hetud keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Hetu with moniker=$MONIKER and chain-id=$CHAINID"
./hetud init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./hetud add-genesis-account \
    "$(./hetud keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000ahetu,1000000000000000000stake \
    --home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./hetud gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./hetud collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./hetud validate-genesis --home $DATA_DIR

echo "starting hetu node $i in background ..."
./hetud start --pruning=nothing --rpc.unsafe \
    --keyring-backend test --home $DATA_DIR \
    >$DATA_DIR/node.log 2>&1 &
disown

echo "started hetu node"
tail -f /dev/null
