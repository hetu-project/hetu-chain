// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/event/types"
)

// Keeper of this module maintains collections of events.
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
}

// NewKeeper returns a new instance of event Keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
