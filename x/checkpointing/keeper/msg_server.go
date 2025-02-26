package keeper

import (
	"context"
	"fmt"

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

	// query the validator's account balance from the staking contract

	err := hetutypes.ValidateAddress(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid eth account address %w", err)
	}

	// store validator BLS public key
	err = m.k.CreateRegistration(ctx, *msg.BlsPubkey, common.HexToAddress(msg.ValidatorAddress))
	if err != nil {
		return nil, err
	}

	return &types.MsgRegistValidatorResponse{}, nil
}
