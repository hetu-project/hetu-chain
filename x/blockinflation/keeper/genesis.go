package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// InitGenesis initializes the blockinflation module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, subspace paramstypes.Subspace, cdc codec.JSONCodec, data *blockinflationtypes.GenesisState) {
	// Set parameters
	k.SetParams(ctx, subspace, data.Params)

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

	k.Logger(ctx).Info("initialized blockinflation genesis state",
		"total_issuance", data.TotalIssuance.String(),
		"total_burned", data.TotalBurned.String(),
		"pending_subnet_rewards", data.PendingSubnetRewards.String(),
	)
}

// ExportGenesis returns the blockinflation module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context, subspace paramstypes.Subspace, cdc codec.JSONCodec) *blockinflationtypes.GenesisState {
	genesis := blockinflationtypes.DefaultGenesisState()
	genesis.Params = k.GetParams(ctx, subspace)
	genesis.TotalIssuance = k.GetTotalIssuance(ctx)
	genesis.TotalBurned = k.GetTotalBurned(ctx)
	genesis.PendingSubnetRewards = k.GetPendingSubnetRewards(ctx)

	return genesis
}
