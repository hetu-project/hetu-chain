package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// SubnetRewards represents the calculated rewards for a subnet
type SubnetRewards struct {
	Netuid   uint16
	TaoIn    math.Int
	AlphaIn  math.Int
	AlphaOut math.Int
	OwnerCut math.Int // Owner cut amount
}

// CalculateSubnetRewards calculates rewards for all subnets participating in emission
func (k Keeper) CalculateSubnetRewards(ctx sdk.Context, subspace paramstypes.Subspace, blockEmission math.Int, subnetsToEmitTo []uint16) (map[uint16]SubnetRewards, error) {
	rewards := make(map[uint16]SubnetRewards)

	// Step 1: Calculate total moving prices
	totalMovingPrices := math.LegacyZeroDec()
	for _, netuid := range subnetsToEmitTo {
		movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)
		totalMovingPrices = totalMovingPrices.Add(movingPrice)
	}

	k.Logger(ctx).Debug("Total moving prices calculated", "total", totalMovingPrices.String())

	// Step 2: Calculate rewards for each subnet
	for _, netuid := range subnetsToEmitTo {
		// Get price information
		price := k.eventKeeper.GetAlphaPrice(ctx, netuid)
		movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)

		// Calculate TAO reward (tao_in)
		var taoIn math.Int
		if totalMovingPrices.IsZero() {
			taoIn = math.ZeroInt()
		} else {
			taoInRatio := movingPrice.Quo(totalMovingPrices)
			taoIn = math.LegacyNewDecFromInt(blockEmission).Mul(taoInRatio).TruncateInt()
		}

		// Calculate Alpha emission
		alphaEmission, err := k.CalculateAlphaEmission(ctx, subspace, netuid)
		if err != nil {
			k.Logger(ctx).Error("failed to calculate Alpha emission", "netuid", netuid, "error", err)
			alphaEmission = math.ZeroInt()
		}

		// Calculate Alpha in (alpha_in)
		var alphaIn math.Int
		if price.IsZero() {
			alphaIn = alphaEmission
		} else {
			idealAlphaIn := math.LegacyNewDecFromInt(taoIn).Quo(price).TruncateInt()
			if idealAlphaIn.GT(alphaEmission) {
				alphaIn = alphaEmission
			} else {
				alphaIn = idealAlphaIn
			}
		}

		// Alpha out equals Alpha emission
		alphaOut := alphaEmission

		rewards[netuid] = SubnetRewards{
			Netuid:   netuid,
			TaoIn:    taoIn,
			AlphaIn:  alphaIn,
			AlphaOut: alphaOut,
		}

		k.Logger(ctx).Debug("Subnet rewards calculated",
			"netuid", netuid,
			"tao_in", taoIn.String(),
			"alpha_in", alphaIn.String(),
			"alpha_out", alphaOut.String(),
			"price", price.String(),
			"moving_price", movingPrice.String(),
		)
	}

	return rewards, nil
}

// ApplySubnetRewards applies the calculated rewards to subnets
// This implements step 4 of the coinbase logic: injection
// Note: This uses the original alpha_out (before owner cut deduction)
func (k Keeper) ApplySubnetRewards(ctx sdk.Context, rewards map[uint16]SubnetRewards) error {
	for netuid, reward := range rewards {
		// Step 4: Injection - Add rewards to subnet pools

		// Add alpha_in to subnet Alpha in pool (for liquidity)
		if reward.AlphaIn.IsPositive() {
			k.eventKeeper.AddSubnetAlphaIn(ctx, netuid, reward.AlphaIn)
			k.eventKeeper.AddSubnetAlphaInEmission(ctx, netuid, reward.AlphaIn)
		}

		// Add alpha_out to subnet Alpha out pool (for distribution)
		// Note: This is the original alpha_out before owner cut deduction
		if reward.AlphaOut.IsPositive() {
			k.eventKeeper.AddSubnetAlphaOut(ctx, netuid, reward.AlphaOut)
			k.eventKeeper.AddSubnetAlphaOutEmission(ctx, netuid, reward.AlphaOut)
		}

		// Add tao_in to subnet TAO pool
		if reward.TaoIn.IsPositive() {
			k.eventKeeper.AddSubnetTAO(ctx, netuid, reward.TaoIn)
			k.eventKeeper.AddSubnetTaoInEmission(ctx, netuid, reward.TaoIn)
		}

		k.Logger(ctx).Info("Injected subnet rewards",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", reward.AlphaIn.String(),
			"alpha_out", reward.AlphaOut.String(),
		)
	}

	return nil
}

// CalculateOwnerCuts calculates owner cuts for all subnets and updates alpha_out
// This implements step 5 of the coinbase logic: owner cuts
func (k Keeper) CalculateOwnerCuts(ctx sdk.Context, subspace paramstypes.Subspace, rewards map[uint16]SubnetRewards) error {
	params := k.GetParams(ctx, subspace)
	cutPercent := params.SubnetOwnerCut

	k.Logger(ctx).Debug("Calculating owner cuts", "cut_percent", cutPercent.String())

	for netuid, reward := range rewards {
		// Step 5.1: Get cut percentage (already done above)

		// Step 5.2: Calculate owner cut
		alphaOut := reward.AlphaOut
		ownerCut := math.LegacyNewDecFromInt(alphaOut).Mul(cutPercent).TruncateInt()

		// Update the reward struct
		reward.OwnerCut = ownerCut
		reward.AlphaOut = alphaOut.Sub(ownerCut) // Subtract owner cut from alpha_out
		rewards[netuid] = reward

		// Step 5.3: Add to pending owner cut pool (for later distribution)
		if ownerCut.IsPositive() {
			k.eventKeeper.AddPendingOwnerCut(ctx, netuid, ownerCut)
		}

		// Step 5.4: Subtract owner cut from subnet Alpha out pool
		// This is equivalent to removing the owner cut from the pool
		if ownerCut.IsPositive() {
			currentAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)
			newAlphaOut := currentAlphaOut.Sub(ownerCut)
			if newAlphaOut.IsNegative() {
				newAlphaOut = math.ZeroInt() // Prevent negative values
			}
			k.eventKeeper.SetSubnetAlphaOut(ctx, netuid, newAlphaOut)
		}

		k.Logger(ctx).Debug("Owner cut calculated",
			"netuid", netuid,
			"alpha_out_original", alphaOut.String(),
			"owner_cut", ownerCut.String(),
			"alpha_out_after_cut", reward.AlphaOut.String(),
			"cut_percent", cutPercent.String(),
		)
	}

	return nil
}

// AddToPendingEmission adds alpha_out to pending emission for each subnet
// This implements step 6 of the coinbase logic: add alpha_out to pending emission
// Since there's no root subnet, we add alpha_out to each subnet's pending emission
func (k Keeper) AddToPendingEmission(ctx sdk.Context, rewards map[uint16]SubnetRewards) error {
	for netuid, reward := range rewards {
		// Step 6: Add alpha_out to pending emission for this subnet
		// pending_alpha = alpha_out_i
		pendingAlpha := reward.AlphaOut

		if pendingAlpha.IsPositive() {
			// Add to pending emission: total = total.saturating_add(pending_alpha)
			k.eventKeeper.AddPendingEmission(ctx, netuid, pendingAlpha)

			k.Logger(ctx).Debug("Added to pending emission",
				"netuid", netuid,
				"alpha_out", pendingAlpha.String(),
			)
		}
	}

	return nil
}
