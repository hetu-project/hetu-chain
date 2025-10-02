package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// BeginBlocker of blockinflation module
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(blockinflationtypes.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Mint and allocate block inflation
	if err := k.MintAndAllocateBlockInflation(ctx); err != nil {
		k.Logger(ctx).Error("failed to mint and allocate block inflation",
			"err", err,
			"height", ctx.BlockHeight())
		return err
	}

	// Synchronize the AMM pool status of all subnets every 20 blocks (previously 100 blocks)
	if ctx.BlockHeight()%20 == 0 {
		k.Logger(ctx).Info("Periodic AMM pool sync", "height", ctx.BlockHeight())
		k.SyncAllAMMPools(ctx)
	}

	// Handling subnet registration events
	k.ProcessBeginBlockEvents(ctx)

	return nil
}
