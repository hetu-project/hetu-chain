package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	yumatypes "github.com/hetu-project/hetu-chain/x/yuma/types"
)

func (k Keeper) CustomAllocateTokens(ctx sdk.Context, totalTokens sdk.DecCoins) {
	subnetCount := k.yumaKeeper.GetActiveSubnetCount(ctx)
	yumaRatio := CalculateYumaRatio(subnetCount)
	bondDenom := k.distrKeeper.GetParams(ctx).BondDenom
	yumaRewards := sdk.NewDecCoins(sdk.NewDecCoinFromDec(
		bondDenom,
		totalTokens.AmountOf(bondDenom).Mul(yumaRatio),
	))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		distrtypes.ModuleName,
		yumatypes.ModuleName,
		yumaRewards.TruncateDecimal(),
	); err != nil {
		panic(err)
	}
	remaining := totalTokens.Sub(yumaRewards)
	k.distrKeeper.AllocateTokens(ctx, remaining)
}

func CalculateYumaRatio(subnetCount int64) sdk.Dec {
	minRatio := sdk.MustNewDecFromStr("0.5")
	maxRatio := sdk.MustNewDecFromStr("0.7")
	base := sdk.NewDec(subnetCount).QuoInt64(100)
	ratio := minRatio.Add(base)
	if ratio.GT(maxRatio) {
		return maxRatio
	}
	return ratio
}
