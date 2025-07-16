package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// BeginBlocker of blockinflation module
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Mint and allocate block inflation
	if err := k.MintAndAllocateBlockInflation(ctx); err != nil {
		k.Logger(ctx).Error("failed to mint and allocate block inflation", "error", err.Error())
		return err
	}

	return nil
}
