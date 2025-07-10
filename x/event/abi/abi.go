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
