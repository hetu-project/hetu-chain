package types

const (
	// ModuleName defines the module name
	ModuleName = "blockinflation"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_blockinflation"
)

var (
	// TotalIssuanceKey defines the key for total issuance
	TotalIssuanceKey = []byte{0x01}

	// TotalBurnedKey defines the key for total burned tokens
	TotalBurnedKey = []byte{0x02}

	// PendingSubnetRewardsKey defines the key for pending subnet rewards
	PendingSubnetRewardsKey = []byte{0x03}

	// PendingSubnetRewardsByNetUIDPrefix defines a map prefix for per-subnet rewards:
	// 0x10 | netuid(2 bytes) -> rewards
	PendingSubnetRewardsByNetUIDPrefix = []byte{0x10}
)
