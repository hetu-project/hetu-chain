package keeper_test

import (
	"encoding/json"
	"math"
	"math/big"
	"time"

	"github.com/hetu-project/hetu/v1/utils"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/hetu-project/hetu/v1/app"
	"github.com/hetu-project/hetu/v1/crypto/ethsecp256k1"
	"github.com/hetu-project/hetu/v1/encoding"
	"github.com/hetu-project/hetu/v1/testutil"
	utiltx "github.com/hetu-project/hetu/v1/testutil/tx"
	evmostypes "github.com/hetu-project/hetu/v1/types"
	evmtypes "github.com/hetu-project/hetu/v1/x/evm/types"
	"github.com/hetu-project/hetu/v1/x/feemarket/types"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
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

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	var err error
	suite.ctx, err = testutil.Commit(suite.ctx, suite.app, t, nil)
	suite.Require().NoError(err)
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.FeeMarketKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// setupTestWithContext sets up a test chain with an example Cosmos send msg,
// given a local (validator config) and a global (feemarket param) minGasPrice
func setupTestWithContext(valMinGasPrice string, minGasPrice sdkmath.LegacyDec, baseFee sdkmath.Int) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	privKey, msg := setupTest(valMinGasPrice + s.denom)
	params := types.DefaultParams()
	params.MinGasPrice = minGasPrice
	err := s.app.FeeMarketKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	s.app.FeeMarketKeeper.SetBaseFee(s.ctx, baseFee.BigInt())
	s.Commit()

	return privKey, msg
}

func setupTest(localMinGasPrices string) (*ethsecp256k1.PrivKey, banktypes.MsgSend) {
	setupChain(localMinGasPrices)

	address, privKey := utiltx.NewAccAddressAndKey()
	amount, ok := sdkmath.NewIntFromString("10000000000000000000")
	s.Require().True(ok)
	initBalance := sdk.Coins{sdk.Coin{
		Denom:  s.denom,
		Amount: amount,
	}}
	err := testutil.FundAccount(s.ctx, s.app.BankKeeper, address, initBalance)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: address.String(),
		ToAddress:   address.String(),
		Amount: sdk.Coins{sdk.Coin{
			Denom:  s.denom,
			Amount: sdkmath.NewInt(10000),
		}},
	}
	// s.Commit()
	return privKey, msg
}

func setupChain(localMinGasPricesStr string) {
	// Initialize the app, so we can use SetMinGasPrices to set the
	// validator-specific min-gas-prices setting
	db := dbm.NewMemDB()
	newapp := app.NewEvmos(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		encoding.MakeConfig(),
		simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
		baseapp.SetMinGasPrices(localMinGasPricesStr),
		baseapp.SetChainID(utils.TestnetChainID+"-1"),
	)

	genesisState := app.NewTestGenesisState(newapp.AppCodec())
	genesisState[types.ModuleName] = newapp.AppCodec().MustMarshalJSON(types.DefaultGenesisState())

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	s.Require().NoError(err)

	// Initialize the chain
	newapp.InitChain(
		&abci.RequestInitChain{
			ChainId:         "hetu_560000-1",
			Validators:      []abci.ValidatorUpdate{},
			AppStateBytes:   stateBytes,
			ConsensusParams: app.DefaultConsensusParams,
		},
	)

	s.app = newapp
	s.SetupApp(false)
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	ethTxParams := &evmtypes.EvmTxArgs{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        to,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Input:     data,
		Accesses:  accesses,
	}
	msgEthereumTx := evmtypes.NewTx(ethTxParams)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}
