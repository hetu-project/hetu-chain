package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/hetu-project/hetu/v1/x/distribution/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramstypes.Subspace

		// keepers
		bankKeeper    types.BankKeeper
		stakingKeeper types.StakingKeeper
		yumaKeeper    types.YumaKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	ps paramstypes.Subspace,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	yk types.YumaKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		bankKeeper:    bk,
		stakingKeeper: sk,
		yumaKeeper:    yk,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the current distribution module parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the distribution module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// CalculateSubnetRewardRatio calculates the dynamic subnet reward ratio based on current subnet count
func (k Keeper) CalculateSubnetRewardRatio(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	subnetCount := k.yumaKeeper.GetSubnetCount(ctx)
	return params.CalculateSubnetRewardRatio(subnetCount)
}

// DistributeBlockRewards distributes block rewards according to the current subnet reward ratio
func (k Keeper) DistributeBlockRewards(ctx sdk.Context, blockRewards sdk.Coins) error {
	if blockRewards.IsZero() {
		return nil
	}

	// Calculate subnet reward ratio
	subnetRewardRatio := k.CalculateSubnetRewardRatio(ctx)
	if subnetRewardRatio.IsZero() {
		// No subnet rewards to distribute
		return nil
	}

	// Calculate subnet reward amount
	subnetRewardAmount := sdk.NewCoins()
	for _, coin := range blockRewards {
		amount := coin.Amount.ToLegacyDec().Mul(subnetRewardRatio)
		subnetRewardAmount = subnetRewardAmount.Add(sdk.NewCoin(coin.Denom, amount.TruncateInt()))
	}

	// Log the distribution
	k.Logger(ctx).Info("distributed block rewards",
		"total_rewards", blockRewards.String(),
		"subnet_reward_ratio", subnetRewardRatio.String(),
		"subnet_rewards", subnetRewardAmount.String(),
		"subnet_count", k.yumaKeeper.GetSubnetCount(ctx),
	)

	return nil
}
