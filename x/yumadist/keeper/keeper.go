package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	yumatypes "github.com/hetu-project/hetu-chain/x/yuma/types"
)

type Keeper struct {
	distrKeeper distrkeeper.Keeper
	yumaKeeper  yumatypes.YumaKeeper
	bankKeeper  keeper.Keeper
	storeKey    sdk.StoreKey
	cdc         codec.BinaryCodec
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	distrKeeper distrkeeper.Keeper,
	yumaKeeper yumatypes.YumaKeeper,
	bankKeeper keeper.Keeper,
) Keeper {
	return Keeper{
		distrKeeper: distrKeeper,
		yumaKeeper:  yumaKeeper,
		bankKeeper:  bankKeeper,
		storeKey:    storeKey,
		cdc:         cdc,
	}
}
