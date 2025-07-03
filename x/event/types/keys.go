// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

const (
	// ModuleName defines the module name
	ModuleName = "event"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for event
	RouterKey = ModuleName
)

// prefix bytes for the event persistent store
const (
	prefixSubnetRegistered = iota + 1
	prefixSubnetMultiParamUpdated
	prefixTaoStaked
	prefixTaoUnstaked
)

// KeyPrefixSubnetRegistered defines prefix key for storing SubnetRegistered events
var KeyPrefixSubnetRegistered = []byte{prefixSubnetRegistered}

// KeyPrefixSubnetMultiParamUpdated defines prefix key for storing SubnetMultiParamUpdated events
var KeyPrefixSubnetMultiParamUpdated = []byte{prefixSubnetMultiParamUpdated}

// KeyPrefixTaoStaked defines prefix key for storing TaoStaked events
var KeyPrefixTaoStaked = []byte{prefixTaoStaked}

// KeyPrefixTaoUnstaked defines prefix key for storing TaoUnstaked events
var KeyPrefixTaoUnstaked = []byte{prefixTaoUnstaked}
