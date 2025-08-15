package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// InitGenesis initializes the blockinflation module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data *blockinflationtypes.GenesisState) {
	if data == nil {
		panic("blockinflation: nil genesis state")
	}

	// Validate parameters
	if data.Params.MintDenom == "" {
		panic("blockinflation: params.mint_denom must not be empty")
	}
	// TODO: If Params exposes a Validate() or ValidateBasic(), call it here.
	// if err := data.Params.Validate(); err != nil { panic(fmt.Errorf("invalid blockinflation params: %w", err)) }

	// Set parameters
	k.SetParams(ctx, data.Params)

	// Debug: Params being set
	k.Logger(ctx).Debug("InitGenesis: setting params", "params", data.Params)

	// Verify parameters were set correctly
	params := k.GetParams(ctx)
	k.Logger(ctx).Debug("InitGenesis: retrieved params", "params", params)

	// Set total issuance
	if err := sdk.ValidateDenom(data.TotalIssuance.Denom); err != nil {
		panic(fmt.Errorf("blockinflation: invalid total_issuance denom: %w", err))
	}
	if data.TotalIssuance.Amount.IsNegative() {
		panic("blockinflation: total_issuance amount must be >= 0")
	}
	if data.TotalIssuance.Denom != data.Params.MintDenom {
		panic(fmt.Errorf("blockinflation: total_issuance denom %q must match params.mint_denom %q", data.TotalIssuance.Denom, data.Params.MintDenom))
	}
	k.SetTotalIssuance(ctx, data.TotalIssuance)

	// Set total burned
	if err := sdk.ValidateDenom(data.TotalBurned.Denom); err != nil {
		panic(fmt.Errorf("blockinflation: invalid total_burned denom: %w", err))
	}
	if data.TotalBurned.Amount.IsNegative() {
		panic("blockinflation: total_burned amount must be >= 0")
	}
	if data.TotalBurned.Denom != data.Params.MintDenom {
		panic(fmt.Errorf("blockinflation: total_burned denom %q must match params.mint_denom %q", data.TotalBurned.Denom, data.Params.MintDenom))
	}
	k.SetTotalBurned(ctx, data.TotalBurned)

	// Set pending subnet rewards
	if err := sdk.ValidateDenom(data.PendingSubnetRewards.Denom); err != nil {
		panic(fmt.Errorf("blockinflation: invalid pending_subnet_rewards denom: %w", err))
	}
	if data.PendingSubnetRewards.Amount.IsNegative() {
		panic("blockinflation: pending_subnet_rewards amount must be >= 0")
	}
	if data.PendingSubnetRewards.Denom != data.Params.MintDenom {
		panic(fmt.Errorf("blockinflation: pending_subnet_rewards denom %q must match params.mint_denom %q", data.PendingSubnetRewards.Denom, data.Params.MintDenom))
	}
	k.SetPendingSubnetRewards(ctx, data.PendingSubnetRewards)

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
