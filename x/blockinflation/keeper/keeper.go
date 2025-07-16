package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramstypes.Subspace

		// keepers
		accountKeeper    types.AccountKeeper
		bankKeeper       types.BankKeeper
		feeCollectorName string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	ps paramstypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeCollectorName string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramstore:       ps,
		accountKeeper:    ak,
		bankKeeper:       bk,
		feeCollectorName: feeCollectorName,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the current blockinflation module parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the blockinflation module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetTotalIssuance returns the total issuance
func (k Keeper) GetTotalIssuance(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TotalIssuanceKey)
	if bz == nil {
		return sdk.NewCoin(k.GetParams(ctx).MintDenom, math.ZeroInt())
	}

	var totalIssuance sdk.Coin
	k.cdc.MustUnmarshal(bz, &totalIssuance)
	return totalIssuance
}

// SetTotalIssuance sets the total issuance
func (k Keeper) SetTotalIssuance(ctx sdk.Context, totalIssuance sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalIssuance)
	store.Set(types.TotalIssuanceKey, bz)
}

// GetTotalBurned returns the total burned tokens
func (k Keeper) GetTotalBurned(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TotalBurnedKey)
	if bz == nil {
		return sdk.NewCoin(k.GetParams(ctx).MintDenom, math.ZeroInt())
	}

	var totalBurned sdk.Coin
	k.cdc.MustUnmarshal(bz, &totalBurned)
	return totalBurned
}

// SetTotalBurned sets the total burned tokens
func (k Keeper) SetTotalBurned(ctx sdk.Context, totalBurned sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&totalBurned)
	store.Set(types.TotalBurnedKey, bz)
}

// BurnTokens burns tokens and updates total burned
func (k Keeper) BurnTokens(ctx sdk.Context, amount sdk.Coin) error {
	params := k.GetParams(ctx)

	// Validate denom
	if amount.Denom != params.MintDenom {
		return fmt.Errorf("invalid denom: expected %s, got %s", params.MintDenom, amount.Denom)
	}

	// Burn tokens from module account
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{amount}); err != nil {
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
