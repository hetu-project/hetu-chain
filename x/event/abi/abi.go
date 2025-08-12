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

// 新增合约 ABI
//
//go:embed SubnetManager.json
var SubnetManagerABI []byte

//go:embed NeuronManager.json
var NeuronManagerABI []byte

//go:embed GlobalStaking.json
var GlobalStakingABI []byte
