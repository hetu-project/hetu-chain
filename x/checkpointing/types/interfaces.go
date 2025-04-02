package types

import (
	"context"

	"cosmossdk.io/core/address"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
)

// AccountKeeper defines the expected interface needed to retrieve account info.
type AccountKeeper interface {
	AddressCodec() address.Codec
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	GetAllAccounts(context.Context) (accounts []sdk.AccountI)
	GetModuleAccountAndPermissions(context.Context, string) (sdk.ModuleAccountI, []string)
}

// ERC20Keeper defines the expected ERC20 keeper interface for supporting
// Call CKPTStaking contract by ckpt module.
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error)
}

// Subspace defines an interface that implements the legacy Cosmos SDK x/params Subspace type.
// This is used to support the legacy Cosmos SDK x/params parameters format.
type Subspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	WithKeyTable(table paramtypes.KeyTable) paramtypes.Subspace
}
