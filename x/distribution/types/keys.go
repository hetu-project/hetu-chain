package types

const (
	// ModuleName defines the module name
	ModuleName = "distribution"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_distribution"
)

var (
	// SubnetRewardPoolKey defines the key for subnet reward pool
	SubnetRewardPoolKey = []byte{0x01}

	// SubnetRewardDistributionKey defines the key for subnet reward distribution
	SubnetRewardDistributionKey = []byte{0x02}
)
