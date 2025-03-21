package keeper

import (
	"context"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	hetutypes "github.com/hetu-project/hetu/v1/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

var _ types.MsgServer = msgServer{}

// RegistValidator registers validator's BLS public key
func (m msgServer) RegistValidator(goCtx context.Context, msg *types.MsgRegistValidator) (*types.MsgRegistValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := hetutypes.ValidateAddress(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid eth account address %w", err)
	}

	// Query the validator's stake from the staking contract
	validatorAddr := common.HexToAddress(msg.ValidatorAddress)
	stake, err := m.k.GetValidatorStake(ctx, validatorAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to query validator stake: %w", err)
	}

	// Check if the validator has staked enough tokens
	if stake.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("validator has not staked any tokens")
	}

	// Store validator BLS public key
	err = m.k.CreateRegistration(ctx, *msg.BlsPubkey, validatorAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegistValidatorResponse{}, nil
}

// RegistStakeContract handles the registration of a validator's staking contract address
func (m msgServer) RegistStakeContract(
	goCtx context.Context,
	msg *types.MsgRegistStakeContract,
) (*types.MsgRegistStakeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddr := common.HexToAddress(msg.ContractAddress)

	err := m.k.StoreValidatorContractAddresses(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegistStakeContractResponse{}, nil
}

// BLSCallback handles the BLS signature upload
func (m msgServer) BLSCallback(goCtx context.Context, msg *types.MsgBLSCallback) (*types.MsgBLSCallbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return m.k.BLSCallback(ctx, msg)
}
