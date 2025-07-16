package keeper

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// RunCoinbase executes the coinbase logic for distributing rewards to subnets
// This is equivalent to the run_coinbase.rs function
func (k Keeper) RunCoinbase(ctx sdk.Context, blockEmission math.Int) error {
	// --- 0. Get current block
	currentBlock := ctx.BlockHeight()
	k.Logger(ctx).Debug("Current block", "block", currentBlock)

	// --- 1. Get all netuids (filter out root)
	allSubnets := k.eventKeeper.GetAllSubnetNetuids(ctx)
	k.Logger(ctx).Debug("All subnet netuids", "subnets", allSubnets)

	// Filter out subnets with no first emission block number
	subnetsToEmitTo := k.eventKeeper.GetSubnetsToEmitTo(ctx)
	k.Logger(ctx).Debug("Subnets to emit to", "subnets", subnetsToEmitTo)

	// If no subnets to emit to, return early
	if len(subnetsToEmitTo) == 0 {
		k.Logger(ctx).Info("No subnets to emit to, skipping coinbase")
		return nil
	}

	// --- 2. Get sum of moving prices (placeholder for now)
	// TODO: Implement moving price calculation
	totalMovingPrices := math.LegacyNewDec(1) // Placeholder
	k.Logger(ctx).Debug("Total moving prices", "total", totalMovingPrices)

	// --- 3. Calculate subnet terms (tao_in, alpha_in, alpha_out)
	// This is a simplified version - you'll need to implement the full logic
	taoIn := make(map[uint16]math.Int)
	alphaIn := make(map[uint16]math.Int)
	alphaOut := make(map[uint16]math.Int)

	for _, netuid := range subnetsToEmitTo {
		// Placeholder calculations - you'll need to implement the full logic
		// based on your specific requirements

		// For now, distribute equally among subnets
		subnetCount := math.NewInt(int64(len(subnetsToEmitTo)))
		taoIn[netuid] = blockEmission.Quo(subnetCount)
		alphaIn[netuid] = math.ZeroInt()  // Placeholder
		alphaOut[netuid] = math.ZeroInt() // Placeholder

		k.Logger(ctx).Debug("Subnet terms calculated",
			"netuid", netuid,
			"tao_in", taoIn[netuid],
			"alpha_in", alphaIn[netuid],
			"alpha_out", alphaOut[netuid],
		)
	}

	// --- 4. Injection (placeholder)
	// TODO: Implement actual injection logic
	k.Logger(ctx).Debug("Injection phase completed")

	// --- 5. Compute owner cuts (placeholder)
	// TODO: Implement owner cut calculation
	k.Logger(ctx).Debug("Owner cuts phase completed")

	// --- 6. Root dividends (placeholder)
	// TODO: Implement root dividend calculation
	k.Logger(ctx).Debug("Root dividends phase completed")

	// --- 7. Update moving prices (placeholder)
	// TODO: Implement moving price updates
	k.Logger(ctx).Debug("Moving prices updated")

	// --- 8. Drain pending emission (placeholder)
	// TODO: Implement epoch-based emission draining
	k.Logger(ctx).Debug("Pending emission drained")

	// Emit event for coinbase execution
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"coinbase_executed",
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", currentBlock)),
			sdk.NewAttribute("block_emission", blockEmission.String()),
			sdk.NewAttribute("subnets_count", fmt.Sprintf("%d", len(subnetsToEmitTo))),
			sdk.NewAttribute("total_moving_prices", totalMovingPrices.String()),
		),
	)

	k.Logger(ctx).Info("Coinbase executed successfully",
		"block", currentBlock,
		"emission", blockEmission.String(),
		"subnets", len(subnetsToEmitTo),
	)

	return nil
}

// GetSubnetEmissionData returns emission data for a specific subnet
// This is a helper function for testing and debugging
func (k Keeper) GetSubnetEmissionData(ctx sdk.Context, netuid uint16) (types.SubnetEmissionData, error) {
	// Check if subnet exists
	_, exists := k.eventKeeper.GetSubnetFirstEmissionBlock(ctx, netuid)
	if !exists {
		return types.SubnetEmissionData{}, fmt.Errorf("subnet %d not found", netuid)
	}

	// TODO: Implement actual emission data calculation
	// For now, return placeholder data
	return types.SubnetEmissionData{
		Netuid:   netuid,
		TaoIn:    math.ZeroInt(),
		AlphaIn:  math.ZeroInt(),
		AlphaOut: math.ZeroInt(),
		OwnerCut: math.ZeroInt(),
		RootDivs: math.ZeroInt(),
	}, nil
}

// GetAllSubnetEmissionData returns emission data for all subnets
func (k Keeper) GetAllSubnetEmissionData(ctx sdk.Context) []types.SubnetEmissionData {
	subnets := k.eventKeeper.GetSubnetsToEmitTo(ctx)
	var data []types.SubnetEmissionData

	for _, netuid := range subnets {
		if emissionData, err := k.GetSubnetEmissionData(ctx, netuid); err == nil {
			data = append(data, emissionData)
		}
	}

	// Sort by netuid for consistent output
	sort.Slice(data, func(i, j int) bool {
		return data[i].Netuid < data[j].Netuid
	})

	return data
}
