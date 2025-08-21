package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// CalculateSubnetRewards calculates rewards for all subnets participating in emission
func (k Keeper) CalculateSubnetRewards(ctx sdk.Context, blockEmission math.Int, subnetsToEmitTo []uint16) (map[uint16]types.SubnetRewards, error) {
	rewards := make(map[uint16]types.SubnetRewards)

	k.Logger(ctx).Debug("Starting subnet rewards calculation",
		"block_emission", blockEmission.String(),
		"subnets_count", len(subnetsToEmitTo),
		"subnets", fmt.Sprintf("%v", subnetsToEmitTo))

	// Step 1: Calculate total moving prices
	totalMovingPrices := math.LegacyZeroDec()
	for _, netuid := range subnetsToEmitTo {
		movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)
		k.Logger(ctx).Debug("Individual subnet moving price",
			"netuid", netuid,
			"moving_price", movingPrice.String())
		totalMovingPrices = totalMovingPrices.Add(movingPrice)
	}

	k.Logger(ctx).Debug("Total moving prices calculated", "total", totalMovingPrices.String())

	// Step 2: Calculate rewards for each subnet
	for _, netuid := range subnetsToEmitTo {
		// Get price information
		price := k.eventKeeper.GetAlphaPrice(ctx, netuid)
		movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)

		// 添加更详细的日志
		k.Logger(ctx).Debug("Subnet price details",
			"netuid", netuid,
			"price", price.String(),
			"moving_price", movingPrice.String())

		// 获取并记录子网的alpha_in和alpha_out
		subnetAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
		subnetAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)
		subnetTao := k.eventKeeper.GetSubnetTAO(ctx, netuid)

		k.Logger(ctx).Debug("Subnet current state",
			"netuid", netuid,
			"subnet_alpha_in", subnetAlphaIn.String(),
			"subnet_alpha_out", subnetAlphaOut.String(),
			"subnet_tao", subnetTao.String())

		// Calculate TAO reward (tao_in)
		var taoIn math.Int
		if totalMovingPrices.IsZero() {
			k.Logger(ctx).Debug("Total moving prices is zero, tao_in will be zero",
				"netuid", netuid)
			taoIn = math.ZeroInt()
		} else {
			taoInRatio := movingPrice.Quo(totalMovingPrices)
			taoIn = math.LegacyNewDecFromInt(blockEmission).Mul(taoInRatio).TruncateInt()

			k.Logger(ctx).Debug("Calculated tao_in",
				"netuid", netuid,
				"tao_in_ratio", taoInRatio.String(),
				"block_emission", blockEmission.String(),
				"tao_in", taoIn.String())
		}

		// Calculate Alpha emission
		alphaEmission, err := k.CalculateAlphaEmission(ctx, netuid)
		if err != nil {
			k.Logger(ctx).Error("failed to calculate Alpha emission", "netuid", netuid, "error", err)
			alphaEmission = math.ZeroInt()
		}

		k.Logger(ctx).Debug("Alpha emission calculated",
			"netuid", netuid,
			"alpha_emission", alphaEmission.String())

		// Calculate Alpha in (alpha_in)
		var alphaIn math.Int
		if price.IsZero() {
			k.Logger(ctx).Debug("Price is zero, alpha_in equals alpha_emission",
				"netuid", netuid,
				"alpha_emission", alphaEmission.String())
			alphaIn = alphaEmission
		} else {
			idealAlphaIn := math.LegacyNewDecFromInt(taoIn).Quo(price).TruncateInt()
			k.Logger(ctx).Debug("Ideal alpha_in calculation",
				"netuid", netuid,
				"tao_in", taoIn.String(),
				"price", price.String(),
				"ideal_alpha_in", idealAlphaIn.String(),
				"alpha_emission", alphaEmission.String())

			if idealAlphaIn.GT(alphaEmission) {
				k.Logger(ctx).Debug("Ideal alpha_in > alpha_emission, capping at alpha_emission",
					"netuid", netuid,
					"ideal_alpha_in", idealAlphaIn.String(),
					"alpha_emission", alphaEmission.String())
				alphaIn = alphaEmission
			} else {
				k.Logger(ctx).Debug("Using ideal alpha_in",
					"netuid", netuid,
					"ideal_alpha_in", idealAlphaIn.String())
				alphaIn = idealAlphaIn
			}
		}

		// Alpha out equals Alpha emission
		alphaOut := alphaEmission

		k.Logger(ctx).Debug("Final alpha_out value",
			"netuid", netuid,
			"alpha_out", alphaOut.String())

		rewards[netuid] = types.SubnetRewards{
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
func (k Keeper) ApplySubnetRewards(ctx sdk.Context, rewards map[uint16]types.SubnetRewards) error {
	for netuid, reward := range rewards {
		// Step 4: Injection - Add rewards to subnet pools
		k.Logger(ctx).Debug("Applying subnet rewards",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", reward.AlphaIn.String(),
			"alpha_out", reward.AlphaOut.String())

		// 获取当前值，用于前后对比
		currentAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
		currentAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)
		currentTaoIn := k.eventKeeper.GetSubnetTAO(ctx, netuid)

		k.Logger(ctx).Debug("Current subnet state before applying rewards",
			"netuid", netuid,
			"current_alpha_in", currentAlphaIn.String(),
			"current_alpha_out", currentAlphaOut.String(),
			"current_tao_in", currentTaoIn.String())

		// Add alpha_in to subnet Alpha in pool (for liquidity)
		if reward.AlphaIn.IsPositive() {
			k.Logger(ctx).Debug("Adding to subnet Alpha in",
				"netuid", netuid,
				"adding_amount", reward.AlphaIn.String(),
				"current_amount", currentAlphaIn.String(),
				"new_amount", currentAlphaIn.Add(reward.AlphaIn).String())

			k.eventKeeper.AddSubnetAlphaIn(ctx, netuid, reward.AlphaIn)
			k.eventKeeper.AddSubnetAlphaInEmission(ctx, netuid, reward.AlphaIn)
		} else {
			k.Logger(ctx).Debug("Skipping alpha_in addition, value not positive",
				"netuid", netuid,
				"alpha_in", reward.AlphaIn.String())
		}

		// Add alpha_out to subnet Alpha out pool (for distribution)
		// Note: This is the original alpha_out before owner cut deduction
		if reward.AlphaOut.IsPositive() {
			k.Logger(ctx).Debug("Adding to subnet Alpha out",
				"netuid", netuid,
				"adding_amount", reward.AlphaOut.String(),
				"current_amount", currentAlphaOut.String(),
				"new_amount", currentAlphaOut.Add(reward.AlphaOut).String())

			k.eventKeeper.AddSubnetAlphaOut(ctx, netuid, reward.AlphaOut)
			k.eventKeeper.AddSubnetAlphaOutEmission(ctx, netuid, reward.AlphaOut)
		} else {
			k.Logger(ctx).Debug("Skipping alpha_out addition, value not positive",
				"netuid", netuid,
				"alpha_out", reward.AlphaOut.String())
		}

		// Add tao_in to subnet TAO pool
		if reward.TaoIn.IsPositive() {
			k.Logger(ctx).Debug("Adding to subnet TAO",
				"netuid", netuid,
				"adding_amount", reward.TaoIn.String(),
				"current_amount", currentTaoIn.String(),
				"new_amount", currentTaoIn.Add(reward.TaoIn).String())

			k.eventKeeper.AddSubnetTAO(ctx, netuid, reward.TaoIn)
			k.eventKeeper.AddSubnetTaoInEmission(ctx, netuid, reward.TaoIn)
		} else {
			k.Logger(ctx).Debug("Skipping tao_in addition, value not positive",
				"netuid", netuid,
				"tao_in", reward.TaoIn.String())
		}

		// 获取更新后的值，用于确认
		updatedAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
		updatedAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)
		updatedTaoIn := k.eventKeeper.GetSubnetTAO(ctx, netuid)

		k.Logger(ctx).Debug("Updated subnet state after applying rewards",
			"netuid", netuid,
			"updated_alpha_in", updatedAlphaIn.String(),
			"updated_alpha_out", updatedAlphaOut.String(),
			"updated_tao_in", updatedTaoIn.String(),
			"alpha_in_change", updatedAlphaIn.Sub(currentAlphaIn).String(),
			"alpha_out_change", updatedAlphaOut.Sub(currentAlphaOut).String(),
			"tao_in_change", updatedTaoIn.Sub(currentTaoIn).String())

		k.Logger(ctx).Info("Injected subnet rewards",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", reward.AlphaIn.String(),
			"alpha_out", reward.AlphaOut.String())
	}

	return nil
}

// CalculateOwnerCuts calculates owner cuts for all subnets and updates alpha_out
// This implements step 5 of the coinbase logic: owner cuts
func (k Keeper) CalculateOwnerCuts(ctx sdk.Context, rewards map[uint16]types.SubnetRewards) error {
	params := k.GetParams(ctx)
	cutPercent := params.SubnetOwnerCut

	k.Logger(ctx).Debug("Calculating owner cuts", "cut_percent", cutPercent.String())

	for netuid, reward := range rewards {
		// Step 5.1: Get cut percentage (already done above)

		// Step 5.2: Calculate owner cut
		alphaOut := reward.AlphaOut
		ownerCut := math.LegacyNewDecFromInt(alphaOut).Mul(cutPercent).TruncateInt()

		// 添加日志：记录初始的reward.AlphaOut值和计算的ownerCut
		k.Logger(ctx).Debug("Owner cut calculation - initial values",
			"netuid", netuid,
			"reward_alpha_out_initial", alphaOut.String(),
			"cut_percent", cutPercent.String(),
			"owner_cut", ownerCut.String(),
		)

		// Update the reward struct
		reward.OwnerCut = ownerCut
		reward.AlphaOut = alphaOut.Sub(ownerCut) // Subtract owner cut from alpha_out
		rewards[netuid] = reward

		// 添加日志：记录更新后的reward.AlphaOut值
		k.Logger(ctx).Debug("Owner cut calculation - updated reward",
			"netuid", netuid,
			"reward_alpha_out_after_cut", reward.AlphaOut.String(),
		)

		// Step 5.3: Add to pending owner cut pool (for later distribution)
		if ownerCut.IsPositive() {
			k.eventKeeper.AddPendingOwnerCut(ctx, netuid, ownerCut)
		}

		// Step 5.4: Subtract owner cut from subnet Alpha out pool
		// This is equivalent to removing the owner cut from the pool
		if ownerCut.IsPositive() {
			currentAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)

			// 添加日志：记录当前的subnet_alpha_out值
			k.Logger(ctx).Debug("Owner cut calculation - current alpha out",
				"netuid", netuid,
				"current_alpha_out", currentAlphaOut.String(),
				"owner_cut", ownerCut.String(),
			)

			newAlphaOut := currentAlphaOut.Sub(ownerCut)

			// 添加日志：记录计算的newAlphaOut值和是否为负
			k.Logger(ctx).Debug("Owner cut calculation - new alpha out",
				"netuid", netuid,
				"new_alpha_out_before_check", newAlphaOut.String(),
				"is_negative", newAlphaOut.IsNegative(),
			)

			if newAlphaOut.IsNegative() {
				k.Logger(ctx).Debug("Owner cut calculation - resetting negative alpha out to zero",
					"netuid", netuid,
					"negative_value", newAlphaOut.String(),
				)
				newAlphaOut = math.ZeroInt() // Prevent negative values
			}
			k.eventKeeper.SetSubnetAlphaOut(ctx, netuid, newAlphaOut)

			// 添加日志：记录最终设置的subnet_alpha_out值
			k.Logger(ctx).Debug("Owner cut calculation - final alpha out",
				"netuid", netuid,
				"final_alpha_out", newAlphaOut.String(),
			)
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
func (k Keeper) AddToPendingEmission(ctx sdk.Context, rewards map[uint16]types.SubnetRewards) error {
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
