package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// GetEpochWindows returns the current epoch windows parameter
func (k Keeper) GetEpochWindows(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.KeyEpochWindows)
	if err != nil {
		panic(err)
	}

	if bz == nil {
		return types.DefaultParams().EpochWindows
	}

	return sdk.BigEndianToUint64(bz)
}

// SetEpochWindows sets the epoch windows parameter
func (k Keeper) SetEpochWindows(ctx context.Context, epochWindows uint64) error {
	if epochWindows == 0 {
		return types.ErrInvalidEpochWindows
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := sdk.Uint64ToBigEndian(epochWindows)
	return store.Set(types.KeyEpochWindows, bz)
}

// GetParams returns the total set of checkpointing parameters.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	return types.NewParams(k.GetEpochWindows(ctx))
}

// SetParams sets the total set of checkpointing parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	return k.SetEpochWindows(ctx, params.EpochWindows)
}
