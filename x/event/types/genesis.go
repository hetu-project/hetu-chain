// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

import (
	"fmt"
)

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
	return &GenesisState{
		Subnets:          make([]Subnet, 0),
		ValidatorStakes:  make([]ValidatorStake, 0),
		Delegations:      make([]Delegation, 0),
		ValidatorWeights: make([]ValidatorWeight, 0),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	seenNetuid := make(map[uint16]bool)
	for _, s := range gs.Subnets {
		if seenNetuid[s.Netuid] {
			return fmt.Errorf("duplicate subnet netuid: %d", s.Netuid)
		}
		seenNetuid[s.Netuid] = true
		if s.Netuid == 0 {
			return fmt.Errorf("subnet netuid must be > 0")
		}
		if s.Owner == "" {
			return fmt.Errorf("subnet owner must not be empty for netuid %d", s.Netuid)
		}
	}
	for _, vs := range gs.ValidatorStakes {
		if vs.Validator == "" {
			return fmt.Errorf("validator stake: empty validator for netuid %d", vs.Netuid)
		}
	}
	for _, d := range gs.Delegations {
		if d.Validator == "" || d.Staker == "" {
			return fmt.Errorf("delegation: empty validator/staker for netuid %d", d.Netuid)
		}
	}
	for _, vw := range gs.ValidatorWeights {
		if vw.Validator == "" {
			return fmt.Errorf("validator weights: empty validator for netuid %d", vw.Netuid)
		}
	}
	return nil
}
