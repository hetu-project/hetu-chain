package checkpointing

import (
	"context"
	"fmt"
	"time"

	"github.com/hetu-project/hetu/v1/testutil/datagen" // datagen for generating testing data
	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	EpochWindows = 5
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the first block after the epoch begins
// then we store the current validator set with BLS keys
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	epochNum := types.CurrentEpochNumber(height, EpochWindows)
	if types.FirstBlockInEpoch(height, EpochWindows) {
		sdkCtx.Logger().Info("Epoch begins", "block height", height, "epoch", epochNum)
		err := k.InitValidatorBLSSet(ctx, epochNum)
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

		// Skip the first epoch, no need to aggregate and seal
		if epochNum == 1 {
			return nil
		}
		// 2. aggregate and seal previous checkpoint, update checkpoint
		ckpt, err := k.GetRawCheckpoint(ctx, epochNum-1)
		if err != nil {
			sdkCtx.Logger().Error("GetRawCheckpoint", "failed to get checkpoint:", err.Error())
			return nil
		}

		if ckpt.Status == types.Sealed {
			sdkCtx.Logger().Error("Checkpoint already sealed", "epoch", epochNum-1)
			return nil
		}

		// Mock data generation for testing
		n := 4 // Number of validators
		totalPower := int64(10) * int64(n)
		msg := types.GetSignBytes(epochNum-1, *ckpt.Ckpt.BlockHash)
		blsPubkeys, blsSigs := datagen.GenRandomPubkeysAndSigs(n, msg)
		valSet := datagen.GenRandomValSet(n)

		// Accumulate signatures
		for i := 0; i < n; i++ {
			err = ckpt.Accumulate(valSet, common.Address(valSet[i].Addr), blsPubkeys[i], blsSigs[i], totalPower)
			if err != nil {
				sdkCtx.Logger().Error("Failed to accumulate BLS", "err", err.Error())
				return nil
			}
		}

		if err := k.SealCheckpoint(ctx, ckpt); err != nil {
			sdkCtx.Logger().Error("failed to update checkpoint", "err", err.Error())
			return nil
		}
	}
	return nil
}
