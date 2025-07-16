package types

// 为模块提供基本的标识符，供其他模块和系统使用
const (
	// ModuleName defines the module name
	ModuleName = "stakework"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_stakework"
)
