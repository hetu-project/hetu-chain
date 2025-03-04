package checkpointing

import (
	"context"
	"fmt"
	"time"

	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EpochWindows = 500
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the first block after the epoch begins
// then we store the current validator set with BLS keys
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	height := ctx.BlockHeight()
	epochNum := types.CurrentEpochNumber(height, EpochWindows)
	if types.FirstBlockInEpoch(height, EpochWindows) {
		ctx.Logger().Info("Epoch begins", "block height", height, "epoch", epochNum)
		err := k.InitValidatorBLSSet(ctx.Context(), epochNum)
		if err != nil {
			panic(fmt.Errorf("failed to store validator BLS set: %w", err))
		}
	}

	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	if conflict := k.GetConflictingCheckpointReceived(ctx); conflict {
		panic(types.ErrConflictingCheckpoint)
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	appHash := sdkCtx.BlockHeader().AppHash
	epochNum := types.CurrentEpochNumber(height, EpochWindows)
	if types.LastBlockInEpoch(height, EpochWindows) {
		// 1. new a checkpoint and save
		k.BuildRawCheckpoint(ctx, epochNum, appHash)

		// 2. aggregate and seal last checkpoint, update checkpoint
		ckpt, err := k.GetRawCheckpoint(ctx, epochNum - 1)
		if err != nil {
			sdkCtx.Logger().Error("GetRawCheckpoint", "ailed to get checkpoint:", err.Error())
			return nil
		}

		if ckpt.Status == types.Sealed {
			sdkCtx.Logger().Info("Checkpoint already sealed", "epoch", epochNum - 1)
			return nil
		}
		// todo: aggregate checkpoint
		// ckpt.Accumulate()
		// if err := k.AggregateCheckpoint(ctx, epochNum); err != nil {
		// 	return fmt.Errorf("failed to aggregate checkpoint: %w", err)
		// }
		if err := k.SealCheckpoint(ctx, ckpt); err != nil {
			return fmt.Errorf("failed to update checkpoint: %w", err)
		}
	}
	return nil
}