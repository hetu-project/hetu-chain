package checkpointing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called at the beginning of every block.
// Upon each BeginBlock, if reaching the first block after the epoch begins
// then we store the current validator set with BLS keys
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	epochWindows := int64(k.GetEpochWindows(ctx))
	epochNum := types.CurrentEpochNumber(height, epochWindows)
	if types.FirstBlockInEpoch(height, epochWindows) {
		sdkCtx.Logger().Info("Epoch begins", "block height", height, "epoch", epochNum)
		err := k.InitValidatorBLSSet(sdkCtx, epochNum)
		if err != nil {
			sdkCtx.Logger().Error("failed to store validator BLS set", "error", err.Error())
			return nil
		}
	}

	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	appHash := sdkCtx.BlockHeader().AppHash
	epochWindows := int64(k.GetEpochWindows(ctx))
	epochNum := types.CurrentEpochNumber(height, epochWindows)
	if types.LastBlockInEpoch(height, epochWindows) {
		// 1. new a checkpoint and save
		newCkptMeta, err := k.BuildRawCheckpoint(ctx, epochNum, appHash)
		if err != nil {
			sdkCtx.Logger().Error("BuildRawCheckpoint", "failed to build checkpoint:", err.Error())
			return nil
		}

		// Skip the first epoch, no need to aggregate and seal
		if epochNum == 1 {
			return nil
		}

		// Get validator set from the stored data
		valSet := k.GetValidatorBlsKeySet(ctx, epochNum-1)
		if valSet == nil || len(valSet.ValSet) == 0 {
			sdkCtx.Logger().Error("No validator set found for epoch", "epoch", epochNum-1)
			return nil
		}

		// 2. Send checkpoint to validators for signing via their dispatcher URLs
		if err := requestBLSSignatures(valSet, newCkptMeta); err != nil {
			sdkCtx.Logger().Error("Failed to request BLS sign", "err", err.Error())
			return nil
		}

		// 3. aggregate and seal previous checkpoint, update checkpoint
		ckptWithMeta, err := k.GetRawCheckpoint(ctx, epochNum-1)
		if err != nil {
			sdkCtx.Logger().Error("GetRawCheckpoint", "failed to get checkpoint:", err.Error())
			return nil
		}

		if ckptWithMeta.Status == types.Sealed {
			sdkCtx.Logger().Error("Checkpoint already sealed", "epoch", epochNum-1)
			return nil
		}

		// if err := k.SealCheckpoint(ctx, ckptWithMeta); err != nil {
		// 	sdkCtx.Logger().Error("failed to update checkpoint", "err", err.Error())
		// 	return nil
		// }
	}
	return nil
}

func requestBLSSignatures(valSet *types.ValidatorWithBlsKeySet, ckpt_with_meta *types.RawCheckpointWithMeta) error {
	type Request struct {
		ValidatorAddress   string                       `json:"validator_address"`
		CheckpointWithMeta *types.RawCheckpointWithMeta `json:"checkpoint_with_meta"`
	}

	// Filter valSet.ValSet to preserve only different validation sets for dispatcher_url
	uniqueValidators := filterUniqueValidators(valSet.ValSet)
	ch := make(chan error, len(uniqueValidators))
	for _, val := range uniqueValidators {
		go func(val *types.ValidatorWithBlsKey) {
			req := Request{
				ValidatorAddress:   val.ValidatorAddress,
				CheckpointWithMeta: ckpt_with_meta,
			}
			reqBody, err := json.Marshal(req)
			if err != nil {
				ch <- err
				return
			}

			resp, err := http.Post(val.DispatcherUrl+"/reqblssign", "application/json", bytes.NewBuffer(reqBody))
			if err != nil {
				ch <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				ch <- fmt.Errorf("failed to request BLS sign from %s, status code: %d", val.DispatcherUrl, resp.StatusCode)
				return
			}

			ch <- nil
		}(val)
	}

	for range uniqueValidators {
		if err := <-ch; err != nil {
			return err
		}
	}
	return nil
}

// filterUniqueValidators Filter the validation sets to preserve only the different validation sets of dispatcher_url
func filterUniqueValidators(validators []*types.ValidatorWithBlsKey) []*types.ValidatorWithBlsKey {
	uniqueValidators := make(map[string]*types.ValidatorWithBlsKey)
	for _, val := range validators {
		if _, exists := uniqueValidators[val.DispatcherUrl]; !exists {
			uniqueValidators[val.DispatcherUrl] = val
		}
	}

	result := make([]*types.ValidatorWithBlsKey, 0, len(uniqueValidators))
	for _, val := range uniqueValidators {
		result = append(result, val)
	}
	return result
}
