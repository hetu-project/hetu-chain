package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
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

// GetEventKeeper 获取 event keeper
func (k Keeper) GetEventKeeper() types.EventKeeper {
	return k.eventKeeper
}
