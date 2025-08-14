// SPDX-License-Identifier: MIT
package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
	stakeworktypes "github.com/hetu-project/hetu/v1/x/stakework/types"
)

// SubnetInfo defines the expected subnet info structure
type SubnetInfo struct {
	Netuid                uint16
	Owner                 string
	AlphaToken            string
	EMAPriceHalvingBlocks uint64
	Params                map[string]string
}

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
type EventKeeper interface {
	GetAllSubnetNetuids(ctx sdk.Context) []uint16
	GetSubnetsToEmitTo(ctx sdk.Context) []uint16
	GetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16) (uint64, bool)
	GetSubnet(ctx sdk.Context, netuid uint16) (eventtypes.Subnet, bool)
	GetAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	GetMovingAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	UpdateMovingPrice(ctx sdk.Context, netuid uint16, movingAlpha math.LegacyDec, halvingBlocks uint64)
	GetSubnetAlphaIn(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOut(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaInEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTaoInEmission(ctx sdk.Context, netuid uint16) math.Int
	AddSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTAO(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTaoInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	SetSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	GetPendingEmission(ctx sdk.Context, netuid uint16) math.Int
	SetPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	GetPendingOwnerCut(ctx sdk.Context, netuid uint16) math.Int
	SetPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)
	GetBlocksSinceLastStep(ctx sdk.Context, netuid uint16) uint64
	SetBlocksSinceLastStep(ctx sdk.Context, netuid uint16, blocks uint64)
	GetLastMechanismStepBlock(ctx sdk.Context, netuid uint16) int64
	SetLastMechanismStepBlock(ctx sdk.Context, netuid uint16, block int64)
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.ValidatorStake
}

// StakeworkKeeper defines the expected interface for the stakework module keeper
type StakeworkKeeper interface {
	ShouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool
	RunEpoch(ctx sdk.Context, netuid uint16, raoEmission uint64) (*stakeworktypes.EpochResult, error)
}

// ERC20Keeper defines the expected interface for the ERC20 module keeper
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)
}
