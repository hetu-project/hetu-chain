package config

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/hetu-project/hetu-hub/v1/encoding"
	"github.com/hetu-project/hetu-hub/v1/x/evm"
	"github.com/hetu-project/hetu-hub/v1/x/feemarket"
)

func MakeConfigForTest(moduleManager module.BasicManager) sdktestutil.TestEncodingConfig {
	config := encoding.MakeConfig()
	if moduleManager == nil {
		moduleManager = module.NewBasicManager(
			auth.AppModuleBasic{},
			bank.AppModuleBasic{},
			distr.AppModuleBasic{},
			gov.NewAppModuleBasic([]govclient.ProposalHandler{paramsclient.ProposalHandler}),
			staking.AppModuleBasic{},
			// Ethermint modules
			evm.AppModuleBasic{},
			feemarket.AppModuleBasic{},
		)
	}
	moduleManager.RegisterLegacyAminoCodec(config.Amino)
	moduleManager.RegisterInterfaces(config.InterfaceRegistry)
	return config
}
