#!/bin/bash

KEY="dev0"
CHAINID="hhub_9000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t hhub-datadir.XXXXX)

echo "create and add new keys"
./hhubd keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init Hhub with moniker=$MONIKER and chain-id=$CHAINID"
./hhubd init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./hhubd add-genesis-account \
"$(./hhubd keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000ahhub,1000000000000000000stake \
--home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./hhubd gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./hhubd collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./hhubd validate-genesis --home $DATA_DIR

echo "starting hhub node $i in background ..."
./hhubd start --pruning=nothing --rpc.unsafe \
--keyring-backend test --home $DATA_DIR \
>$DATA_DIR/node.log 2>&1 & disown

echo "started hhub node"
tail -f /dev/null