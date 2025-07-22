package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// BeginBlocker of blockinflation module
func (k Keeper) BeginBlocker(ctx sdk.Context, subspace paramstypes.Subspace) error {
	defer telemetry.ModuleMeasureSince(blockinflationtypes.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Mint and allocate block inflation
	if err := k.MintAndAllocateBlockInflation(ctx, subspace); err != nil {
		k.Logger(ctx).Error("failed to mint and allocate block inflation", "error", err.Error())
		return err
	}

	return nil
}
