// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	evmostypes "github.com/hetu-project/hetu/v1/types"

	utils "github.com/hetu-project/hetu/v1/utils"
	"github.com/hetu-project/hetu/v1/x/inflation/types"
)

// 200M token at year 4 allocated to the team
var teamAlloc = math.NewInt(200_000_000).Mul(evmostypes.PowerReduction)

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(
	ctx sdk.Context,
	coin sdk.Coin,
	params types.Params,
) (
	staking, incentives, communityPool sdk.Coins,
	err error,
) {
	// skip as no coins need to be minted
	if coin.Amount.IsNil() || !coin.Amount.IsPositive() {
		return nil, nil, nil, nil
	}

	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		return nil, nil, nil, err
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	return k.AllocateExponentialInflation(ctx, coin, params)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.Coins{coin}
	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
//   - usage incentives -> `x/incentives` module
//   - community pool -> `sdk `distr` module community pool
func (k Keeper) AllocateExponentialInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
	params types.Params,
) (
	staking, incentives, communityPool sdk.Coins,
	err error,
) {
	distribution := params.InflationDistribution

	// Allocate staking rewards into fee collector account
	staking = sdk.Coins{k.GetProportions(ctx, mintedCoin, distribution.StakingRewards)}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		staking,
	); err != nil {
		return nil, nil, nil, err
	}

	// Allocate usage incentives to incentives module account
	incentives = sdk.Coins{k.GetProportions(ctx, mintedCoin, distribution.UsageIncentives)}

	if err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		incentives,
	); err != nil {
		return nil, nil, nil, err
	}

	// Allocate community pool amount (remaining module balance) to community
	// pool address
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	inflationBalance := k.bankKeeper.GetAllBalances(ctx, moduleAddr)

	err = k.distrKeeper.FundCommunityPool(
		ctx,
		inflationBalance,
		moduleAddr,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return staking, incentives, communityPool, nil
}

// GetAllocationProportion calculates the proportion of coins that is to be
// allocated during inflation for a given distribution.
func (k Keeper) GetProportions(
	_ sdk.Context,
	coin sdk.Coin,
	distribution math.LegacyDec,
) sdk.Coin {
	return sdk.Coin{
		Denom:  coin.Denom,
		Amount: math.LegacyNewDecFromInt(coin.Amount).Mul(distribution).TruncateInt(),
	}
}

// BondedRatio the fraction of the staking tokens which are currently bonded
// It doesn't consider team allocation for inflation
func (k Keeper) BondedRatio(ctx sdk.Context) math.LegacyDec {
	stakeSupply, err := k.stakingKeeper.StakingTokenSupply(ctx)
	if err != nil {
		return math.LegacyZeroDec()
	}

	isMainnet := utils.IsMainnet(ctx.ChainID())

	if !stakeSupply.IsPositive() || (isMainnet && stakeSupply.LTE(teamAlloc)) {
		return math.LegacyZeroDec()
	}

	// don't count team allocation in bonded ratio's stake supple
	if isMainnet {
		stakeSupply = stakeSupply.Sub(teamAlloc)
	}

	totalBondedTokens, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return math.LegacyZeroDec()
	}

	return math.LegacyNewDecFromInt(totalBondedTokens).QuoInt(stakeSupply)
}

// GetCirculatingSupply returns the bank supply of the mintDenom excluding the
// team allocation in the first year
func (k Keeper) GetCirculatingSupply(ctx sdk.Context, mintDenom string) math.LegacyDec {
	circulatingSupply := math.LegacyNewDecFromInt(k.bankKeeper.GetSupply(ctx, mintDenom).Amount)
	teamAllocation := math.LegacyNewDecFromInt(teamAlloc)

	// Consider team allocation only on mainnet chain id
	if utils.IsMainnet(ctx.ChainID()) {
		circulatingSupply = circulatingSupply.Sub(teamAllocation)
	}

	return circulatingSupply
}

// GetInflationRate returns the inflation rate for the current period.
func (k Keeper) GetInflationRate(ctx sdk.Context, mintDenom string) math.LegacyDec {
	epp := k.GetEpochsPerPeriod(ctx)
	if epp == 0 {
		return math.LegacyZeroDec()
	}

	epochMintProvision := k.GetEpochMintProvision(ctx)
	if epochMintProvision.IsZero() {
		return math.LegacyZeroDec()
	}

	epochsPerPeriod := math.LegacyNewDec(epp)

	circulatingSupply := k.GetCirculatingSupply(ctx, mintDenom)
	if circulatingSupply.IsZero() {
		return math.LegacyZeroDec()
	}

	// EpochMintProvision * 365 / circulatingSupply * 100
	return epochMintProvision.Mul(epochsPerPeriod).Quo(circulatingSupply).Mul(math.LegacyNewDec(100))
}

// GetEpochMintProvision retrieves necessary params KV storage
// and calculate EpochMintProvision
func (k Keeper) GetEpochMintProvision(ctx sdk.Context) math.LegacyDec {
	return types.CalculateEpochMintProvision(
		k.GetParams(ctx),
		k.GetPeriod(ctx),
		k.GetEpochsPerPeriod(ctx),
		k.BondedRatio(ctx),
	)
}
