package checkpointing

import (
	"context"
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
	last_block := (height - 1) % EpochWindows
	if last_block == 0 && height != 1 {
		ctx.Logger().Info("Epoch begins", "height", height)
		// new a checkpoint
		// err := k.NewCheckpoint(ctx)
	}

	return nil
}

func EndBlocker(ctx context.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	if conflict := k.GetConflictingCheckpointReceived(ctx); conflict {
		panic(types.ErrConflictingCheckpoint)
	}
}
