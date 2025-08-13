package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
	stakeworktypes "github.com/hetu-project/hetu/v1/x/stakework/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetSupply(ctx context.Context, denom string) sdk.Coin
}

// EventKeeper defines the expected event keeper for getting subnet count
type EventKeeper interface {
	GetSubnetCount(ctx sdk.Context) uint64
	GetAllSubnetNetuids(ctx sdk.Context) []uint16
	GetSubnetsToEmitTo(ctx sdk.Context) []uint16
	GetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16) (uint64, bool)
	GetAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	GetMovingAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	UpdateMovingPrice(ctx sdk.Context, netuid uint16, movingAlpha math.LegacyDec, halvingBlocks uint64)
	GetSubnet(ctx sdk.Context, netuid uint16) (eventtypes.Subnet, bool)
	GetSubnetAlphaIn(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOut(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.ValidatorStake
	AddSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	SetSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTAO(ctx sdk.Context, netuid uint16, amount math.Int)
	GetSubnetAlphaInEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTaoInEmission(ctx sdk.Context, netuid uint16) math.Int
	AddSubnetAlphaInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTaoInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	GetPendingOwnerCut(ctx sdk.Context, netuid uint16) math.Int
	SetPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	SetBlocksSinceLastStep(ctx sdk.Context, netuid uint16, value uint64)
	GetBlocksSinceLastStep(ctx sdk.Context, netuid uint16) uint64
	SetLastMechanismStepBlock(ctx sdk.Context, netuid uint16, blockHeight int64)
	GetLastMechanismStepBlock(ctx sdk.Context, netuid uint16) int64
	SetPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	GetPendingEmission(ctx sdk.Context, netuid uint16) math.Int
}

// StakeworkKeeper defines the expected interface for interacting with stakework module
// Only declare the methods that are needed
// Note: ctx is sdk.Context
// RunEpoch returns *types.EpochResult, error
// ShouldRunEpoch returns bool
// You can supplement methods according to stakework keeper implementation

type StakeworkKeeper interface {
	RunEpoch(ctx sdk.Context, netuid uint16, emission uint64) (*stakeworktypes.EpochResult, error)
	ShouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool
}
