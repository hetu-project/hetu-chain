package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// InitGenesis initializes the blockinflation module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data *types.GenesisState) {
	// Set parameters
	k.SetParams(ctx, data.Params)

	// Set total issuance
	if !data.TotalIssuance.Amount.IsNil() {
		k.SetTotalIssuance(ctx, data.TotalIssuance)
	}

	// Set total burned
	if !data.TotalBurned.Amount.IsNil() {
		k.SetTotalBurned(ctx, data.TotalBurned)
	}

	k.Logger(ctx).Info("initialized blockinflation genesis state",
		"total_issuance", data.TotalIssuance.String(),
		"total_burned", data.TotalBurned.String(),
	)
}

// ExportGenesis returns the blockinflation module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) *types.GenesisState {
	genesis := types.DefaultGenesisState()
	genesis.Params = k.GetParams(ctx)
	genesis.TotalIssuance = k.GetTotalIssuance(ctx)
	genesis.TotalBurned = k.GetTotalBurned(ctx)

	return genesis
}
