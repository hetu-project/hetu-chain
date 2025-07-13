package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines the distribution module's genesis state.
type GenesisState struct {
	// Params defines all the parameters of the distribution module.
	Params Params `json:"params" yaml:"params"`
	// SubnetRewardPool defines the accumulated subnet rewards pool
	SubnetRewardPool sdk.Coins `json:"subnet_reward_pool" yaml:"subnet_reward_pool"`
	// SubnetRewardDistribution defines the pending subnet reward distributions
	SubnetRewardDistribution []SubnetRewardDistribution `json:"subnet_reward_distribution" yaml:"subnet_reward_distribution"`
}

// SubnetRewardDistribution defines a pending subnet reward distribution
type SubnetRewardDistribution struct {
	// SubnetID is the ID of the subnet
	SubnetID string `json:"subnet_id" yaml:"subnet_id"`
	// ValidatorAddress is the validator address
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`
	// Amount is the reward amount
	Amount sdk.Coin `json:"amount" yaml:"amount"`
	// Epoch is the epoch when this reward was calculated
	Epoch int64 `json:"epoch" yaml:"epoch"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                   DefaultParams(),
		SubnetRewardPool:         sdk.NewCoins(),
		SubnetRewardDistribution: []SubnetRewardDistribution{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate subnet reward pool
	if err := gs.SubnetRewardPool.Validate(); err != nil {
		return err
	}

	// Validate subnet reward distributions
	for _, distribution := range gs.SubnetRewardDistribution {
		if err := distribution.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate performs basic validation on subnet reward distribution
func (srd SubnetRewardDistribution) Validate() error {
	if srd.SubnetID == "" {
		return ErrEmptySubnetID
	}
	if srd.ValidatorAddress == "" {
		return ErrEmptyValidatorAddress
	}
	if err := srd.Amount.Validate(); err != nil {
		return err
	}
	if srd.Epoch < 0 {
		return ErrInvalidEpoch
	}
	return nil
}
