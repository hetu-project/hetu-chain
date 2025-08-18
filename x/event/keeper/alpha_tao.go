package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetSubnetAlphaIn 设置子网的 AlphaIn 值
func (k Keeper) SetSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// SetSubnetTaoIn 设置子网的 TaoIn 值
func (k Keeper) SetSubnetTaoIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}
