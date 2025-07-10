package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// ModuleName 模块名称
	ModuleName = "yuma"

	// StoreKey 存储键
	StoreKey = ModuleName

	// QuerierRoute 查询路由
	QuerierRoute = ModuleName

	// 存储键前缀
	SubnetInfoPrefix          = "subnet_info/"
	NeuronInfoPrefix          = "neuron_info/"
	WeightsPrefix             = "weights/"
	BondsPrefix               = "bonds/"
	LastEpochBlockPrefix      = "last_epoch_block/"
	LastWeightsUpdatePrefix   = "last_weights_update/"
	LastValidatorEpochPrefix  = "last_validator_epoch/"
	ValidatorEpochCountPrefix = "validator_epoch_count/"
)

// 前缀函数
func GetSubnetInfoPrefix(netuid uint16) []byte {
	return append([]byte(SubnetInfoPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetNeuronInfoPrefix(netuid uint16) []byte {
	return append([]byte(NeuronInfoPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetWeightsPrefix(netuid uint16) []byte {
	return append([]byte(WeightsPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetBondsPrefix(netuid uint16) []byte {
	return append([]byte(BondsPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

// 键函数
func GetSubnetInfoKey(netuid uint16) []byte {
	return append([]byte(SubnetInfoPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetNeuronInfoKey(netuid uint16, uid uint16) []byte {
	return append(GetNeuronInfoPrefix(netuid), sdk.Uint64ToBigEndian(uint64(uid))...)
}

func GetWeightsKey(netuid uint16, uid uint16) []byte {
	return append(GetWeightsPrefix(netuid), sdk.Uint64ToBigEndian(uint64(uid))...)
}

func GetBondsKey(netuid uint16, uid uint16) []byte {
	return append(GetBondsPrefix(netuid), sdk.Uint64ToBigEndian(uint64(uid))...)
}

func GetLastEpochBlockKey(netuid uint16) []byte {
	return append([]byte(LastEpochBlockPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetLastWeightsUpdateKey(netuid uint16, uid uint16) []byte {
	prefix := append([]byte(LastWeightsUpdatePrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
	return append(prefix, sdk.Uint64ToBigEndian(uint64(uid))...)
}

func GetLastValidatorEpochKey(netuid uint16) []byte {
	return append([]byte(LastValidatorEpochPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}

func GetValidatorEpochCountKey(netuid uint16) []byte {
	return append([]byte(ValidatorEpochCountPrefix), sdk.Uint64ToBigEndian(uint64(netuid))...)
}
