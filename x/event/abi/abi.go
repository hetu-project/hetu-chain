package abi

import _ "embed"

//go:embed SubnetRegistry.json
var SubnetRegistryABI []byte

//go:embed StakingSelfNative.json
var StakingSelfABI []byte

//go:embed StakingDelegatedNative.json
var StakingDelegatedABI []byte

//go:embed Weights.json
var WeightsABI []byte

// New contract ABIs
//
//go:embed SubnetManager.json
var SubnetManagerABI []byte

//go:embed NeuronManager.json
var NeuronManagerABI []byte

//go:embed GlobalStaking.json
var GlobalStakingABI []byte

//go:embed AlphaToken.json
var AlphaTokenABI []byte

//go:embed SubnetAMM.json
var SubnetAMMABI []byte

//go:embed WHETU.json
var WHETUABI []byte
