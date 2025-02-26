package keeper_test

import (
	"math"
	"time"

	"github.com/hetu-project/hetu/v1/utils"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/hetu-project/hetu/v1/crypto/ethsecp256k1"
	"github.com/hetu-project/hetu/v1/encoding"
	"github.com/hetu-project/hetu/v1/testutil"
	utiltx "github.com/hetu-project/hetu/v1/testutil/tx"
	evmostypes "github.com/hetu-project/hetu/v1/types"
	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
	"github.com/hetu-project/hetu/v1/x/feemarket/types"

	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	t := suite.T()
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), "hetu_560000-1", suite.consAddress, nil, nil,
	)

	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, header)
	suite.ctx = suite.ctx.WithBlockGasMeter(storetypes.NewGasMeter(math.MaxUint64))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	acc := &evmostypes.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 18, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, sdk.ValAddress(validator.GetOperator()))
	// require.NoError(t, err)

	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = utils.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	encodingConfig := encoding.MakeConfig()
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
}
