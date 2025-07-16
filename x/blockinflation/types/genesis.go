package types

import (
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
	return &GenesisState{
		Params:               DefaultParams(),
		TotalIssuance:        sdk.NewCoin("ahetu", math.ZeroInt()),
		TotalBurned:          sdk.NewCoin("ahetu", math.ZeroInt()),
		PendingSubnetRewards: sdk.NewCoin("ahetu", math.ZeroInt()),
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

	return nil
}
