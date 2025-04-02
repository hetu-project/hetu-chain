package v2

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// MigrateStore migrates the x/checkpointing module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/checkpointing module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	legacySubspace types.Subspace,
) error {
	store := ctx.KVStore(storeKey)

	// Get current parameters from legacy subspace
	var params types.Params
	legacySubspace = legacySubspace.WithKeyTable(types.ParamKeyTable())
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	// Store the epoch windows parameter
	store.Set(types.KeyEpochWindows, sdk.Uint64ToBigEndian(params.EpochWindows))

	return nil
}
