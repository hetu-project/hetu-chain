// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

// GenesisState defines the event module's genesis state.
type GenesisState struct {
	Subnets          []Subnet          `json:"subnets"`
	ValidatorStakes  []ValidatorStake  `json:"validator_stakes"`
	Delegations      []Delegation      `json:"delegations"`
	ValidatorWeights []ValidatorWeight `json:"validator_weights"`
}

// NewGenesisState creates a new genesis state instance
func NewGenesisState(
	subnets []Subnet,
	validatorStakes []ValidatorStake,
	delegations []Delegation,
	validatorWeights []ValidatorWeight,
) *GenesisState {
	return &GenesisState{
		Subnets:          subnets,
		ValidatorStakes:  validatorStakes,
		Delegations:      delegations,
		ValidatorWeights: validatorWeights,
	}
}

// DefaultGenesisState returns the default event genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil, nil, nil, nil)
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// 可根据需要添加更详细的校验逻辑
	return nil
}
