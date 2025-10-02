#!/bin/bash

# Hetu Chain Information query script
# Used to query block height, coin issuance, community pool balance, and various proportion configurations

echo "=== Hetu Chain Information inquiry ==="
echo ""

# Set the node URL
NODE_URL=${NODE_URL:-"http://localhost:26657"}
NODE_STATUS_URL="${NODE_URL%/}/status"
NODE_FLAG="--node $NODE_URL"

# Check if the node is running
if ! curl -s "$NODE_STATUS_URL" > /dev/null; then
    echo "❌ The node is not running, please start the node first:"
    echo "   ./local_node.sh"
    echo ""
    echo "Or connect to a remote node:"
    echo "   export NODE_URL=https://your-rpc-endpoint"
    exit 1
fi

echo "🔗 Connect to the node: $NODE_URL"
echo ""

# 1. Query the current block height
echo "📊 1. Current block height"
echo "----------------------------------------"
BLOCK_HEIGHT=$(hetud status $NODE_FLAG --output json 2>/dev/null | jq -r '.sync_info.latest_block_height // "Cannot get"')
echo "Current block height: $BLOCK_HEIGHT"
echo ""

# 2. Query the total coin issuance (inflation related)
echo "💰 2. Coin issuance information"
echo "----------------------------------------"

# Query the current circulating supply
echo "🔍 Query the circulating supply..."
CIRCULATING_SUPPLY=$(hetud query inflation circulating-supply $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$CIRCULATING_SUPPLY" != "" ]; then
    # The circulating supply is returned in string format, such as: "279688800000000000010000.000000000000000000ahetu"
    echo "Circulating supply: $CIRCULATING_SUPPLY"
else
    echo "Circulating supply: query failed (node may not be running or module may not be available)"
fi

# Query the current epoch coin issuance
echo "🔍 Query the current epoch coin issuance..."
EPOCH_PROVISION=$(hetud query inflation epoch-mint-provision $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$EPOCH_PROVISION" != "" ]; then
    echo "$EPOCH_PROVISION" | jq -r 'if type == "object" then "Current epoch coin issuance: " + (.epoch_mint_provision.amount // "Unknown") + " " + (.epoch_mint_provision.denom // "ahetu") else "Current epoch coin issuance: query failed" end'
else
    echo "Current epoch coin issuance: query failed (node may not be running or module may not be available)"
fi

# Query the inflation rate
echo "🔍 Query the inflation rate..."
INFLATION_RATE=$(hetud query inflation inflation-rate $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$INFLATION_RATE" != "" ]; then
    echo "$INFLATION_RATE" | jq -r 'if type == "object" then "Current inflation rate: " + (.inflation_rate // "Unknown") + "%" else "Current inflation rate: query failed" end'
else
    echo "Current inflation rate: query failed (node may not be running or module may not be available)"
fi

# Query the current period
echo "🔍 Query the current period..."
PERIOD=$(hetud query inflation period $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$PERIOD" != "" ]; then
    echo "$PERIOD" | jq -r 'if type == "object" then "Current period: " + (.period // "Unknown" | tostring) elif type == "number" then "Current period: " + (. | tostring) else "Current period: query failed" end'
else
    echo "Current period: query failed (node may not be running or module may not be available)"
fi

# Query the block inflation parameters
echo "🔍 Query the block inflation parameters..."
hetud query blockinflation params $NODE_FLAG --output json 2>/dev/null | jq '.params'

echo ""

# 3. Query the community pool balance
echo "🏛️ 3. Community pool information"
echo "----------------------------------------"
echo "🔍 Query the community pool balance..."
COMMUNITY_POOL=$(hetud query distribution community-pool $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$COMMUNITY_POOL" != "" ]; then
    echo "$COMMUNITY_POOL" | jq -r '.pool[] | "Community pool: " + . + " (original format)"'
else
    echo "Community pool: query failed (node may not be running)"
fi
echo ""

# 4. Query the inflation distribution ratio
echo "⚖️ 4. Inflation distribution ratio"
echo "----------------------------------------"
echo "🔍 Query the inflation parameters..."
INFLATION_PARAMS=$(hetud query inflation params $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$INFLATION_PARAMS" != "" ]; then
    echo "Inflation distribution ratio:"        
    echo "$INFLATION_PARAMS" | jq -r '
        "  - Staking rewards: " + (.inflation_distribution.staking_rewards // "0"),
        "  - Usage incentives: " + (.inflation_distribution.usage_incentives // "0"), 
        "  - Community pool: " + (.inflation_distribution.community_pool // "0")'
    echo ""
    echo "Exponential calculation parameters:"    
    echo "$INFLATION_PARAMS" | jq -r '
        "  - Initial value(a): " + (.exponential_calculation.a // "0"),
        "  - Decay rate(r): " + (.exponential_calculation.r // "0"),
        "  - Long-term value(c): " + (.exponential_calculation.c // "0"),
        "  - Bonding target: " + (.exponential_calculation.bonding_target // "0"),
        "  - Maximum variance: " + (.exponential_calculation.max_variance // "0")'
else
    echo "Failed to get inflation parameters (node may not be running)"
fi
echo ""

# 5. Query the staking parameters (including commission related)
echo "🔗 5. Staking and Commission information"
echo "----------------------------------------"
echo "🔍 Query the staking parameters..."
STAKING_PARAMS=$(hetud query staking params $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$STAKING_PARAMS" != "" ]; then
    echo "Staking parameters:"
    echo "$STAKING_PARAMS" | jq -r '
    if .params then
        "  - Maximum number of validators: " + (.params.max_validators // 0 | tostring),
        "  - Maximum number of delegations: " + (.params.max_entries // 0 | tostring),
        "  - Unbonding time: " + (.params.unbonding_time // "0"),
        "  - Minimum commission rate: " + (.params.min_commission_rate // "0")
    else
        "  Failed to parse staking parameters"
    end'
else
    echo "Failed to get staking parameters (node may not be running)"
fi

# Query all validator information (including commission)
echo ""
echo "🔍 Query the validator commission information..."
hetud query staking validators $NODE_FLAG --output json 2>/dev/null | jq -r '
.validators[] | 
"Validator: " + .description.moniker + 
" | Commission rate: " + .commission.commission_rates.rate + 
" | Maximum commission rate: " + .commission.commission_rates.max_rate'

echo ""

# 6. Query the distribution parameters
echo "📈 6. Distribution parameters"
echo "----------------------------------------"
echo "🔍 Query the distribution parameters..."
DIST_PARAMS=$(hetud query distribution params $NODE_FLAG --output json 2>/dev/null)
if [ $? -eq 0 ] && [ "$DIST_PARAMS" != "" ]; then
    echo "Distribution parameters:"
    echo "$DIST_PARAMS" | jq -r '
        "  - Community pool tax: " + (.params.community_tax // "0"),
        "  - Base proposer reward: " + (.params.base_proposer_reward // "0"),
        "  - Withdraw address change enabled: " + (.params.withdraw_addr_enabled // false | tostring)'
else
    echo "Failed to get distribution parameters (node may not be running)"
fi

echo ""
echo "=== Query completed ==="
echo ""
echo "💡 Tips:"
echo "   - If some queries fail, please ensure that the node is running and synchronized"
echo "   - You can use hetud status to check the node status"
echo "   - Coin issuance will be calculated dynamically based on block height and algorithm"
echo "   - Community pool balance will change with inflation and governance proposals"
