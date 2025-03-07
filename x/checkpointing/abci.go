package checkpointing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
		err := k.InitValidatorBLSSet(sdkCtx, epochNum)
		if err != nil {
			panic(fmt.Errorf("failed to store validator BLS set: %w", err))
		}
	}

	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()
	appHash := sdkCtx.BlockHeader().AppHash
	epochNum := types.CurrentEpochNumber(height, EpochWindows)
	if types.LastBlockInEpoch(height, EpochWindows) {
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
		if err := requestBLSSignatures(valSet, newCkptMeta.Ckpt); err != nil {
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

		// Calculate total power
		var totalPower int64 = 0
		for _, val := range valSet.ValSet {
			totalPower += int64(val.VotingPower)
		}

		// Convert to types.ValidatorSet
		validatorSet := make(types.ValidatorSet, len(valSet.ValSet))
		for i, val := range valSet.ValSet {
			validatorSet[i] = types.Validator{
				Addr:  common.HexToAddress(val.ValidatorAddress).Bytes(),
				Power: int64(val.VotingPower),
			}
		}

		// Generate message to be signed
		msg := types.GetSignBytes(epochNum-1, *ckptWithMeta.Ckpt.BlockHash)
		// todo: mock signatures, replace with actual signature collection
		_, blsSig := datagen.GenRandomPubkeysAndSigs(1, msg)

		// Simulate getting signatures and accumulating them
		for _, val := range valSet.ValSet {
			valAddr := common.HexToAddress(val.ValidatorAddress)
			blsPubkey := val.BlsPubKey

			// Accumulate the signature
			err = ckptWithMeta.Accumulate(validatorSet, valAddr, blsPubkey, blsSig[0], totalPower)
			if err != nil {
				sdkCtx.Logger().Error("Failed to accumulate BLS", "validator", val.ValidatorAddress, "err", err.Error())
				continue
			}
		}

		if err := k.SealCheckpoint(ctx, ckptWithMeta); err != nil {
			sdkCtx.Logger().Error("failed to update checkpoint", "err", err.Error())
			return nil
		}
	}
	return nil
}

func requestBLSSignatures(valSet *types.ValidatorWithBlsKeySet, ckpt *types.RawCheckpoint) error {
	type Request struct {
		ValidatorAddress string               `json:"validator_address"`
		Checkpoint       *types.RawCheckpoint `json:"checkpoint"`
	}

	ch := make(chan error, len(valSet.ValSet))
	for _, val := range valSet.ValSet {
		go func(val *types.ValidatorWithBlsKey) {
			req := Request{
				ValidatorAddress: val.ValidatorAddress,
				Checkpoint:       ckpt,
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

	for range valSet.ValSet {
		if err := <-ch; err != nil {
			return err
		}
	}
	return nil
}
