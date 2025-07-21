package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/stakework/types"
)

// Keeper 简化的 yuma keeper
type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	eventKeeper types.EventKeeper
}

// NewKeeper 创建新的 keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	eventKeeper types.EventKeeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		eventKeeper: eventKeeper,
	}
}

// Logger 返回模块的 logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/stakework")
}

// GetEventKeeper 获取 event keeper
func (k Keeper) GetEventKeeper() types.EventKeeper {
	return k.eventKeeper
}
