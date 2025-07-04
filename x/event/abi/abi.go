package abi

import _ "embed"

//go:embed SubnetRegistry.json
var SubnetRegistryABI []byte

//go:embed SubnetParamManager.json
var SubnetParamManagerABI []byte

//go:embed TaoStaking.json
var TaoStakingABI []byte
