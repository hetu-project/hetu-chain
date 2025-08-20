package types

// Module name and store keys
const (
	// ModuleName defines the module name
	ModuleName = "subnet"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_subnet"
)

// Subnet key prefixes
const (
	// SubnetKeyPrefix is the prefix for subnet storage
	SubnetKeyPrefix = "subnet:"

	// SubnetHyperparamsKeyPrefix is the prefix for subnet hyperparameters storage
	SubnetHyperparamsKeyPrefix = "subnet_hyperparams:"

	// NeuronKeyPrefix is the prefix for neuron storage
	NeuronKeyPrefix = "neuron:"
)

// GetSubnetKey returns the store key for a specific subnet
func GetSubnetKey(netuid uint16) []byte {
	return []byte(SubnetKeyPrefix + string(netuid))
}

// GetSubnetHyperparamsKey returns the store key for a specific subnet's hyperparameters
func GetSubnetHyperparamsKey(netuid uint16) []byte {
	return []byte(SubnetHyperparamsKeyPrefix + string(netuid))
}

// GetNeuronKey returns the store key for a specific neuron in a subnet
func GetNeuronKey(netuid uint16, uid string) []byte {
	return []byte(NeuronKeyPrefix + string(netuid) + ":" + uid)
}

// GetNeuronsBySubnetPrefix returns the prefix for all neurons in a subnet
func GetNeuronsBySubnetPrefix(netuid uint16) []byte {
	return []byte(NeuronKeyPrefix + string(netuid) + ":")
}
