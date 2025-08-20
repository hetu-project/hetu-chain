package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// 定义Querier类型
type Querier func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error)

// NewQuerier creates a new querier for event module
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case "subnets":
			return querySubnets(ctx, k, legacyQuerierCdc)
		case "subnet":
			return querySubnet(ctx, path[1:], k, legacyQuerierCdc)
		default:
			return nil, fmt.Errorf("unknown event query endpoint: %s", path[0])
		}
	}
}

// NewLegacyQuerier creates a legacy querier for the event module
func NewLegacyQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case "subnets":
			return querySubnets(ctx, k, legacyQuerierCdc)
		case "subnet":
			return querySubnet(ctx, path[1:], k, legacyQuerierCdc)
		default:
			return nil, fmt.Errorf("unknown event query endpoint: %s", path[0])
		}
	}
}

// querySubnets returns all subnets
func querySubnets(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	subnets := k.GetAllSubnetInfos(ctx)
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, subnets)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subnets: %w", err)
	}
	return res, nil
}

// querySubnet handles subnet-related queries
func querySubnet(ctx sdk.Context, path []string, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if len(path) < 1 {
		return nil, fmt.Errorf("subnet netuid required")
	}

	netuidStr := path[0]
	netuid64, err := strconv.ParseUint(netuidStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid netuid: %w", err)
	}
	netuid := uint16(netuid64)

	// Handle different subnet-related queries
	if len(path) == 1 {
		// Query subnet info
		return querySubnetInfo(ctx, netuid, k, legacyQuerierCdc)
	} else if len(path) == 2 {
		switch path[1] {
		case "neurons":
			return querySubnetNeurons(ctx, netuid, k, legacyQuerierCdc)
		case "pool":
			return querySubnetPool(ctx, netuid, k, legacyQuerierCdc)
		default:
			return nil, fmt.Errorf("unknown subnet query endpoint: %s", path[1])
		}
	}

	return nil, fmt.Errorf("unknown subnet query endpoint")
}

// querySubnetInfo returns subnet info by ID
func querySubnetInfo(ctx sdk.Context, netuid uint16, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	subnetInfo, found := k.GetSubnetInfo(ctx, netuid)
	if !found {
		return nil, fmt.Errorf("subnet with netuid %d not found", netuid)
	}

	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, subnetInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subnet info: %w", err)
	}
	return res, nil
}

// querySubnetNeurons returns all neurons in a subnet
func querySubnetNeurons(ctx sdk.Context, netuid uint16, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	neurons := k.GetAllNeuronInfosByNetuid(ctx, netuid)
	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, neurons)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal neurons: %w", err)
	}
	return res, nil
}

// querySubnetPool returns pool info for a subnet
func querySubnetPool(ctx sdk.Context, netuid uint16, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	subnet, found := k.GetSubnet(ctx, netuid)
	if !found {
		return nil, fmt.Errorf("subnet with netuid %d not found", netuid)
	}

	// Create a response struct with pool information
	type PoolResponse struct {
		Netuid       uint16 `json:"netuid"`
		AmmPool      string `json:"amm_pool"`
		LockedAmount string `json:"locked_amount"`
		BurnedAmount string `json:"burned_amount"`
	}

	poolResponse := PoolResponse{
		Netuid:       subnet.Netuid,
		AmmPool:      subnet.AmmPool,
		LockedAmount: subnet.LockedAmount,
		BurnedAmount: subnet.BurnedAmount,
	}

	res, err := json.Marshal(poolResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pool response: %w", err)
	}
	return res, nil
}
