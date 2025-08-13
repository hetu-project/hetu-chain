package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         storetypes.StoreKey
		memKey           storetypes.StoreKey
		accountKeeper    blockinflationtypes.AccountKeeper
		bankKeeper       blockinflationtypes.BankKeeper
		eventKeeper      blockinflationtypes.EventKeeper
		stakeworkKeeper  blockinflationtypes.StakeworkKeeper
		feeCollectorName string
		subspace         paramstypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	ak blockinflationtypes.AccountKeeper,
	bk blockinflationtypes.BankKeeper,
	ek blockinflationtypes.EventKeeper,
	stakeworkKeeper blockinflationtypes.StakeworkKeeper,
	feeCollectorName string,
	subspace paramstypes.Subspace,
) *Keeper {
	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		accountKeeper:    ak,
		bankKeeper:       bk,
		eventKeeper:      ek,
		stakeworkKeeper:  stakeworkKeeper,
		feeCollectorName: feeCollectorName,
		subspace:         subspace,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", blockinflationtypes.ModuleName))
}

// GetParams returns the current blockinflation module parameters
func (k Keeper) GetParams(ctx sdk.Context) blockinflationtypes.Params {
	var params blockinflationtypes.Params

	defer func() {
		if r := recover(); r != nil {
			k.Logger(ctx).Warn("Panic in GetParamSet, writing default params", "panic", r)
			// Write default parameters to KVStore
			k.SetParams(ctx, blockinflationtypes.DefaultParams())
			params = blockinflationtypes.DefaultParams()
		}
	}()
	k.subspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the blockinflation module parameters
func (k Keeper) SetParams(ctx sdk.Context, params blockinflationtypes.Params) {
	k.subspace.SetParamSet(ctx, &params)
}

// GetTotalIssuance returns the total issuance
func (k Keeper) GetTotalIssuance(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(blockinflationtypes.TotalIssuanceKey)
	if bz == nil {
		// Return zero coin if not found
		return sdk.NewCoin("ahetu", math.ZeroInt())
	}

	var totalIssuance sdk.Coin
	k.cdc.MustUnmarshal(bz, &totalIssuance)
	return totalIssuance
}

// SetTotalIssuance sets the total issuance
func (k Keeper) SetTotalIssuance(ctx sdk.Context, totalIssuance sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalIssuance)
	store.Set(blockinflationtypes.TotalIssuanceKey, bz)
}

// GetTotalBurned returns the total burned tokens
func (k Keeper) GetTotalBurned(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(blockinflationtypes.TotalBurnedKey)
	if bz == nil {
		// Return zero coin if not found
		return sdk.NewCoin("ahetu", math.ZeroInt())
	}

	var totalBurned sdk.Coin
	k.cdc.MustUnmarshal(bz, &totalBurned)
	return totalBurned
}

// SetTotalBurned sets the total burned tokens
func (k Keeper) SetTotalBurned(ctx sdk.Context, totalBurned sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalBurned)
	store.Set(blockinflationtypes.TotalBurnedKey, bz)
}

// GetPendingSubnetRewards returns the pending subnet rewards
func (k Keeper) GetPendingSubnetRewards(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(blockinflationtypes.PendingSubnetRewardsKey)
	if bz == nil {
		// Return zero coin if not found
		return sdk.NewCoin("ahetu", math.ZeroInt())
	}

	var pendingRewards sdk.Coin
	k.cdc.MustUnmarshal(bz, &pendingRewards)
	return pendingRewards
}

// SetPendingSubnetRewards sets the pending subnet rewards
func (k Keeper) SetPendingSubnetRewards(ctx sdk.Context, pendingRewards sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pendingRewards)
	store.Set(blockinflationtypes.PendingSubnetRewardsKey, bz)
}

// AddToPendingSubnetRewards adds tokens to the pending subnet rewards pool
func (k Keeper) AddToPendingSubnetRewards(ctx sdk.Context, amount sdk.Coin) {
	params := k.GetParams(ctx)

	// Validate denom
	if amount.Denom != params.MintDenom {
		k.Logger(ctx).Error("invalid denom for pending subnet rewards",
			"expected", params.MintDenom,
			"got", amount.Denom,
		)
		return
	}

	currentPending := k.GetPendingSubnetRewards(ctx)
	newPending := currentPending.Add(amount)
	k.SetPendingSubnetRewards(ctx, newPending)

	k.Logger(ctx).Info("added to pending subnet rewards",
		"amount", amount.String(),
		"total_pending", newPending.String(),
	)
}

// BurnTokens burns tokens and updates total burned
func (k Keeper) BurnTokens(ctx sdk.Context, amount sdk.Coin) error {
	params := k.GetParams(ctx)

	// Validate denom
	if amount.Denom != params.MintDenom {
		return fmt.Errorf("invalid denom: expected %s, got %s", params.MintDenom, amount.Denom)
	}

	// Burn tokens from module account
	if err := k.bankKeeper.BurnCoins(ctx, blockinflationtypes.ModuleName, sdk.Coins{amount}); err != nil {
		return err
	}

	// Update total burned
	currentBurned := k.GetTotalBurned(ctx)
	newBurned := currentBurned.Add(amount)
	k.SetTotalBurned(ctx, newBurned)

	// Update total issuance (decrease)
	currentIssuance := k.GetTotalIssuance(ctx)
	newIssuance := currentIssuance.Sub(amount)
	k.SetTotalIssuance(ctx, newIssuance)

	k.Logger(ctx).Info("burned tokens",
		"amount", amount.String(),
		"total_burned", newBurned.String(),
		"total_issuance", newIssuance.String(),
	)

	return nil
}
