// SPDX-License-Identifier: MIT
package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
	stakeworktypes "github.com/hetu-project/hetu/v1/x/stakework/types"
)

// SubnetInfo defines the expected subnet info structure
// Deprecated: Use eventtypes.SubnetInfo directly
type SubnetInfo = eventtypes.SubnetInfo

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	GetSupply(ctx context.Context, denom string) sdk.Coin
}

// EventKeeper defines the expected interface for the event module keeper
// Deprecated: Use eventtypes.EventKeeper directly
type EventKeeper = eventtypes.EventKeeper

// StakeworkKeeper defines the expected interface for the stakework module keeper
type StakeworkKeeper interface {
	ShouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool
	RunEpoch(ctx sdk.Context, netuid uint16, raoEmission uint64) (*stakeworktypes.EpochResult, error)
}

// ERC20Keeper defines the expected interface for the ERC20 module keeper
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)
}
