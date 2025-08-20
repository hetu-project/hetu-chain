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

	// 每 20 个区块同步一次所有子网的 AMM 池状态（原来是100个区块）
	if ctx.BlockHeight()%20 == 0 {
		k.Logger(ctx).Info("Periodic AMM pool sync", "height", ctx.BlockHeight())
		k.SyncAllAMMPools(ctx)
	}

	// 处理子网注册事件
	k.ProcessBeginBlockEvents(ctx)

	return nil
}
