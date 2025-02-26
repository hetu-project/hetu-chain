package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/crypto/bls12381"
	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

func (suite *KeeperTestSuite) TestRegistValidator() {
	testCases := []struct {
		name     string
		malleate func() *types.MsgRegistValidator
		expPass  bool
	}{
		{
			"pass - valid address and BLS public key",
			func() *types.MsgRegistValidator {
				_, pub := bls12381.GenKeyPair()
				validatorAddress := sdk.ValAddress(suite.address.Bytes()).String()
				return &types.MsgRegistValidator{
					BlsPubkey:        &pub,
					ValidatorAddress: validatorAddress,
				}
			},
			true,
		},
		{
			"error - invalid validator address",
			func() *types.MsgRegistValidator {
				blsPubkey := &bls12381.PublicKey{}
				*blsPubkey = bls12381.PublicKey([]byte("validBlsPubkey"))
				validatorAddress := "invalidAddress"
				return &types.MsgRegistValidator{
					BlsPubkey:        blsPubkey,
					ValidatorAddress: validatorAddress,
				}
			},
			false,
		},
		{
			"error - failed to create registration",
			func() *types.MsgRegistValidator {
				blsPubkey := &bls12381.PublicKey{}
				*blsPubkey = bls12381.PublicKey([]byte("validBlsPubkey"))
				validatorAddress := sdk.ValAddress(suite.address.Bytes()).String()
				suite.app.CheckpointingKeeper.CreateRegistration(suite.ctx, *blsPubkey, common.HexToAddress(validatorAddress))
				return &types.MsgRegistValidator{
					BlsPubkey:        blsPubkey,
					ValidatorAddress: validatorAddress,
				}
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()

			msg := tc.malleate()
			msgServer := keeper.NewMsgServerImpl(suite.app.CheckpointingKeeper)
			_, err := msgServer.RegistValidator(suite.ctx, msg)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
