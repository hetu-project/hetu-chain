package v1

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	checkpointingkeeper "github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	checkpointingKeeper checkpointingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Set the epoch windows parameter
		setEpochWindows(ctx, checkpointingKeeper)

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// setEpochWindows sets the epoch windows parameter
func setEpochWindows(ctx sdk.Context, ck checkpointingkeeper.Keeper) {
	// Get the current epoch windows parameter from the keeper
	currentEpochWindows := ck.GetEpochWindows(ctx)

	// Create a new params object with the current epoch windows
	params := types.NewParams(currentEpochWindows)

	// Set the params in the keeper
	err := ck.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}
}
