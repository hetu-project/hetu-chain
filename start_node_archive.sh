#!/bin/bash

# Exit on error
set -e

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <node_number>"
    echo "Example: $0 0"
    exit 1
fi

NODE_NUMBER=$1
HOME_PREFIX="$HOME/.hhubd${NODE_NUMBER}"

# Start the node
hhubd start \
    --home "${HOME_PREFIX}" \
    --chain-id hhub_9000-1 \
    --log_level info
