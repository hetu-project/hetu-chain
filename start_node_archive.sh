#!/bin/bash

# Exit on error
set -e

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <node_number>"
    echo "Example: $0 0"
    exit 1
fi

NODE_NUMBER=$1
HOME_PREFIX="$HOME/.hetud${NODE_NUMBER}"

# Start the node
hetud start \
    --home "${HOME_PREFIX}" \
    --chain-id hetu_560000-1 \
    --log_level info
