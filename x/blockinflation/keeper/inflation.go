package keeper

import (
	"fmt"
	stdmath "math"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// CalculateBlockEmission calculates the block emission based on Bittensor's algorithm
// Improved version with high-precision calculations to avoid floating-point precision issues
func (k Keeper) CalculateBlockEmission(ctx sdk.Context) (math.Int, error) {
	params := k.GetParams(ctx)
	totalIssuance := k.GetTotalIssuance(ctx)

	// Check if we've reached total supply
	if totalIssuance.Amount.GTE(params.TotalSupply) {
		return math.ZeroInt(), nil
	}

	// Use high-precision math.LegacyDec calculations instead of float64
	totalIssuanceDec := totalIssuance.Amount.ToLegacyDec()
	totalSupplyDec := params.TotalSupply.ToLegacyDec()
	defaultBlockEmissionDec := params.DefaultBlockEmission.ToLegacyDec()

	// Calculate the ratio: total_issuance / total_supply
	ratio := totalIssuanceDec.Quo(totalSupplyDec)

	// If ratio >= 1.0, return 0
	if ratio.GTE(math.LegacyOneDec()) {
		return math.ZeroInt(), nil
	}

	// Calculate log2(1 / (1 - ratio)) using high-precision arithmetic
	// logArg = 1 / (1 - ratio)
	oneMinusRatio := math.LegacyOneDec().Sub(ratio)
	if oneMinusRatio.LTE(math.LegacyZeroDec()) {
		return math.ZeroInt(), nil
	}

	logArg := math.LegacyOneDec().Quo(oneMinusRatio)

	// Convert to float64 for log2 calculation (this is the only place we need float64)
	logArgFloat := logArg.MustFloat64()
	logResult := stdmath.Log2(logArgFloat)

	// Floor the log result
	flooredLog := stdmath.Floor(logResult)
	flooredLogInt := int64(flooredLog)

	// Calculate 2^flooredLog
	multiplier := stdmath.Pow(2.0, float64(flooredLogInt))

	// Calculate block emission percentage: 1 / multiplier
	blockEmissionPercentage := math.LegacyOneDec().Quo(math.LegacyNewDecWithPrec(int64(multiplier*1000), 3))

	// Calculate actual block emission using high-precision arithmetic
	blockEmission := defaultBlockEmissionDec.Mul(blockEmissionPercentage)

	// Convert back to math.Int with proper rounding
	blockEmissionInt := blockEmission.TruncateInt()

	k.Logger(ctx).Debug("calculated block emission (high-precision)",
		"total_issuance", totalIssuance.String(),
		"total_supply", params.TotalSupply.String(),
		"ratio", ratio.String(),
		"log_arg", logArg.String(),
		"log_result", fmt.Sprintf("%.6f", logResult),
		"floored_log", flooredLogInt,
		"multiplier", fmt.Sprintf("%.6f", multiplier),
		"emission_percentage", blockEmissionPercentage.String(),
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
	subnetCount := uint64(len(k.eventKeeper.GetAllSubnetNetuids(ctx)))

	k.Logger(ctx).Info("Test whether the subnet was successfully obtained",
		"subnetCount", subnetCount,
	)

	// Calculate subnet reward ratio
	subnetRewardRatio := blockinflationtypes.CalculateSubnetRewardRatio(params, subnetCount)

	// Calculate subnet reward amount
	subnetRewardAmount := math.LegacyNewDecFromInt(blockEmission).Mul(subnetRewardRatio).TruncateInt()

	k.Logger(ctx).Info("Test whether the subnet reward ratio was successfully obtained",
		"subnetRewardRatio", subnetRewardRatio.String(),
	)

	// Calculate remaining amount for fee collector
	feeCollectorAmount := blockEmission.Sub(subnetRewardAmount)

	// Create minted coin
	mintedCoin := sdk.Coin{
		Denom:  params.MintDenom,
		Amount: blockEmission,
	}
	// Mint coins
	if err := k.bankKeeper.MintCoins(ctx, blockinflationtypes.ModuleName, sdk.Coins{mintedCoin}); err != nil {
		return fmt.Errorf("failed to mint coins: %w", err)
	}
	// Add subnet reward to pending pool
	if subnetRewardAmount.IsPositive() {
		subnetRewardCoin := sdk.Coin{
			Denom:  params.MintDenom,
			Amount: subnetRewardAmount,
		}
		k.AddToPendingSubnetRewards(ctx, subnetRewardCoin)

		// Execute coinbase logic to distribute rewards to subnets
		if err := k.RunCoinbase(ctx, subnetRewardAmount); err != nil {
			k.Logger(ctx).Error("failed to execute coinbase", "error", err)
			// Don't return error here to avoid blocking inflation
		}
	}

	// Send remaining amount to fee collector
	if feeCollectorAmount.IsPositive() {
		feeCollectorCoin := sdk.Coin{
			Denom:  params.MintDenom,
			Amount: feeCollectorAmount,
		}
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			blockinflationtypes.ModuleName,
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
