package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines the blockinflation module's genesis state.
type GenesisState struct {
	// Params defines all the parameters of the blockinflation module.
	Params Params `json:"params" yaml:"params"`
	// TotalIssuance defines the total issuance at genesis
	TotalIssuance sdk.Coin `json:"total_issuance" yaml:"total_issuance"`
	// TotalBurned defines the total burned tokens at genesis
	TotalBurned sdk.Coin `json:"total_burned" yaml:"total_burned"`
	// PendingSubnetRewards defines the pending subnet rewards pool
	PendingSubnetRewards sdk.Coin `json:"pending_subnet_rewards" yaml:"pending_subnet_rewards"`
}

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() *GenesisState {
	p := DefaultParams()
	denom := p.MintDenom
	return &GenesisState{
		Params:               p,
		TotalIssuance:        sdk.NewCoin(denom, math.ZeroInt()),
		TotalBurned:          sdk.NewCoin(denom, math.ZeroInt()),
		PendingSubnetRewards: sdk.NewCoin(denom, math.ZeroInt()),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate total issuance
	if err := gs.TotalIssuance.Validate(); err != nil {
		return err
	}

	// Validate total burned
	if err := gs.TotalBurned.Validate(); err != nil {
		return err
	}

	// Validate pending subnet rewards
	if err := gs.PendingSubnetRewards.Validate(); err != nil {
		return err
	}

	// Ensure all coins use the configured mint denom
	if gs.TotalIssuance.Denom != gs.Params.MintDenom {
		return fmt.Errorf("total_issuance denom %q must match params.mint_denom %q", gs.TotalIssuance.Denom, gs.Params.MintDenom)
	}
	if gs.TotalBurned.Denom != gs.Params.MintDenom {
		return fmt.Errorf("total_burned denom %q must match params.mint_denom %q", gs.TotalBurned.Denom, gs.Params.MintDenom)
	}
	if gs.PendingSubnetRewards.Denom != gs.Params.MintDenom {
		return fmt.Errorf("pending_subnet_rewards denom %q must match params.mint_denom %q", gs.PendingSubnetRewards.Denom, gs.Params.MintDenom)
	}

	return nil
}
