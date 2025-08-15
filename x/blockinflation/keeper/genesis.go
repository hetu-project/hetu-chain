package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// InitGenesis initializes the blockinflation module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data *blockinflationtypes.GenesisState) {
	// Set parameters
	k.SetParams(ctx, data.Params)

	// Debug: Params being set
	k.Logger(ctx).Debug("InitGenesis: setting params", "params", data.Params)

	// Verify parameters were set correctly
	params := k.GetParams(ctx)
	k.Logger(ctx).Debug("InitGenesis: retrieved params", "params", params)

	// Set total issuance
	if !data.TotalIssuance.Amount.IsNil() {
		k.SetTotalIssuance(ctx, data.TotalIssuance)
	}

	// Set total burned
	if !data.TotalBurned.Amount.IsNil() {
		k.SetTotalBurned(ctx, data.TotalBurned)
	}

	// Set pending subnet rewards
	if !data.PendingSubnetRewards.Amount.IsNil() {
		k.SetPendingSubnetRewards(ctx, data.PendingSubnetRewards)
	}

	k.Logger(ctx).Info("blockinflation: initialized genesis state",
		"total_issuance", data.TotalIssuance.String(),
		"total_burned", data.TotalBurned.String(),
		"pending_subnet_rewards", data.PendingSubnetRewards.String(),
	)
}

// ExportGenesis returns the blockinflation module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) *blockinflationtypes.GenesisState {
	return &blockinflationtypes.GenesisState{
		Params:               k.GetParams(ctx),
		TotalIssuance:        k.GetTotalIssuance(ctx),
		TotalBurned:          k.GetTotalBurned(ctx),
		PendingSubnetRewards: k.GetPendingSubnetRewards(ctx),
	}
}
