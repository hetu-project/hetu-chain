package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Set the AlphaIn value of the subnet for Set Subnet AlphaIn
func (k Keeper) SetSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// Set the TaoI value of the subnet using Set Subnet TaoI
func (k Keeper) SetSubnetTaoIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}
