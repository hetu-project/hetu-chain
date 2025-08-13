package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/stakework/types"
)

// Keeper simplified yuma keeper
type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	eventKeeper types.EventKeeper
}

// NewKeeper creates a new keeper
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

// Logger returns the module's logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetEventKeeper gets the event keeper
func (k Keeper) GetEventKeeper() types.EventKeeper {
	return k.eventKeeper
}
