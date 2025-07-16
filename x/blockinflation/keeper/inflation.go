package keeper

import (
	"fmt"
	stdmath "math"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// CalculateBlockEmission calculates the block emission based on Bittensor's algorithm
func (k Keeper) CalculateBlockEmission(ctx sdk.Context) (math.Int, error) {
	params := k.GetParams(ctx)
	totalIssuance := k.GetTotalIssuance(ctx)

	// Check if we've reached total supply
	if totalIssuance.Amount.GTE(params.TotalSupply) {
		return math.ZeroInt(), nil
	}

	// Convert to float64 for calculation
	totalIssuanceFloat := float64(totalIssuance.Amount.Int64())
	totalSupplyFloat := float64(params.TotalSupply.Int64())
	defaultBlockEmissionFloat := float64(params.DefaultBlockEmission.Int64())

	// Calculate the ratio: total_issuance / (2 * total_supply)
	ratio := totalIssuanceFloat / (2.0 * totalSupplyFloat)

	// Calculate log2(1 / (1 - ratio))
	// This is equivalent to: log2(1 / (1 - total_issuance / (2 * total_supply)))
	if ratio >= 1.0 {
		return math.ZeroInt(), nil
	}

	logArg := 1.0 / (1.0 - ratio)
	logResult := stdmath.Log2(logArg)

	// Floor the log result
	flooredLog := stdmath.Floor(logResult)
	flooredLogInt := int64(flooredLog)

	// Calculate 2^flooredLog
	multiplier := stdmath.Pow(2.0, float64(flooredLogInt))

	// Calculate block emission percentage: 1 / multiplier
	blockEmissionPercentage := 1.0 / multiplier

	// Calculate actual block emission
	blockEmission := blockEmissionPercentage * defaultBlockEmissionFloat

	// Convert back to math.Int
	blockEmissionInt := math.NewInt(int64(blockEmission))

	k.Logger(ctx).Debug("calculated block emission",
		"total_issuance", totalIssuance.String(),
		"total_supply", params.TotalSupply.String(),
		"ratio", fmt.Sprintf("%.6f", ratio),
		"log_result", fmt.Sprintf("%.6f", logResult),
		"floored_log", flooredLogInt,
		"multiplier", fmt.Sprintf("%.6f", multiplier),
		"emission_percentage", fmt.Sprintf("%.6f", blockEmissionPercentage),
		"block_emission", blockEmissionInt.String(),
	)

	return blockEmissionInt, nil
}

// MintAndAllocateBlockInflation mints coins and allocates them to fee collector
func (k Keeper) MintAndAllocateBlockInflation(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Skip if inflation is disabled
	if !params.EnableBlockInflation {
		return nil
	}

	// Calculate block emission
	blockEmission, err := k.CalculateBlockEmission(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate block emission: %w", err)
	}

	// Skip if no emission
	if !blockEmission.IsPositive() {
		return nil
	}

	// Get subnet count for reward calculation
	subnetCount := k.eventKeeper.GetSubnetCount(ctx)

	// Calculate subnet reward ratio
	subnetRewardRatio := types.CalculateSubnetRewardRatio(params, subnetCount)

	// Calculate subnet reward amount
	subnetRewardAmount := math.LegacyNewDecFromInt(blockEmission).Mul(subnetRewardRatio).TruncateInt()

	// Calculate remaining amount for fee collector
	feeCollectorAmount := blockEmission.Sub(subnetRewardAmount)

	// Create minted coin
	mintedCoin := sdk.Coin{
		Denom:  params.MintDenom,
		Amount: blockEmission,
	}

	// Mint coins
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{mintedCoin}); err != nil {
		return fmt.Errorf("failed to mint coins: %w", err)
	}

	// Add subnet reward to pending pool
	if subnetRewardAmount.IsPositive() {
		subnetRewardCoin := sdk.Coin{
			Denom:  params.MintDenom,
			Amount: subnetRewardAmount,
		}
		k.AddToPendingSubnetRewards(ctx, subnetRewardCoin)
	}

	// Send remaining amount to fee collector
	if feeCollectorAmount.IsPositive() {
		feeCollectorCoin := sdk.Coin{
			Denom:  params.MintDenom,
			Amount: feeCollectorAmount,
		}
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			k.feeCollectorName,
			sdk.Coins{feeCollectorCoin},
		); err != nil {
			return fmt.Errorf("failed to send coins to fee collector: %w", err)
		}
	}

	// Update total issuance
	currentIssuance := k.GetTotalIssuance(ctx)
	newIssuance := currentIssuance.Add(mintedCoin)
	k.SetTotalIssuance(ctx, newIssuance)

	k.Logger(ctx).Info("minted and allocated block inflation",
		"block_height", ctx.BlockHeight(),
		"minted_amount", mintedCoin.String(),
		"subnet_count", subnetCount,
		"subnet_reward_ratio", subnetRewardRatio.String(),
		"subnet_reward_amount", subnetRewardAmount.String(),
		"fee_collector_amount", feeCollectorAmount.String(),
		"total_issuance", newIssuance.String(),
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"block_inflation_minted",
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("minted_amount", mintedCoin.Amount.String()),
			sdk.NewAttribute("mint_denom", mintedCoin.Denom),
			sdk.NewAttribute("subnet_count", fmt.Sprintf("%d", subnetCount)),
			sdk.NewAttribute("subnet_reward_ratio", subnetRewardRatio.String()),
			sdk.NewAttribute("subnet_reward_amount", subnetRewardAmount.String()),
			sdk.NewAttribute("fee_collector_amount", feeCollectorAmount.String()),
			sdk.NewAttribute("total_issuance", newIssuance.Amount.String()),
		),
	)

	return nil
}
