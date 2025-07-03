// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

// GenesisState defines the event module's genesis state.
type GenesisState struct {
	SubnetRegisteredEvents        []SubnetRegisteredEvent        `json:"subnet_registered_events"`
	SubnetMultiParamUpdatedEvents []SubnetMultiParamUpdatedEvent `json:"subnet_multi_param_updated_events"`
	TaoStakedEvents               []TaoStakedEvent               `json:"tao_staked_events"`
	TaoUnstakedEvents             []TaoUnstakedEvent             `json:"tao_unstaked_events"`
}

// NewGenesisState creates a new genesis state instance
func NewGenesisState(
	subnetRegistered []SubnetRegisteredEvent,
	subnetMultiParamUpdated []SubnetMultiParamUpdatedEvent,
	taoStaked []TaoStakedEvent,
	taoUnstaked []TaoUnstakedEvent,
) *GenesisState {
	return &GenesisState{
		SubnetRegisteredEvents:        subnetRegistered,
		SubnetMultiParamUpdatedEvents: subnetMultiParamUpdated,
		TaoStakedEvents:               taoStaked,
		TaoUnstakedEvents:             taoUnstaked,
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
