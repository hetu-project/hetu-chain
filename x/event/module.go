// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/hetu-project/hetu/v1/x/event/keeper"
	"github.com/hetu-project/hetu/v1/x/event/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasABCIGenesis = AppModule{}

	_ appmodule.HasBeginBlocker = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

type AppModuleBasic struct {
	cdc codec.Codec
}

func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino)   {}
func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router)                        {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}
func (AppModuleBasic) GetTxCmd() *cobra.Command                                                  { return nil }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                                               { return nil }

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

func (AppModule) QuerierRoute() string                          { return "" }
func (am AppModule) RegisterServices(cfg module.Configurator)   {}
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	// TODO: 调用 keeper 初始化链上事件数据
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// TODO: 导出链上事件数据
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (am AppModule) BeginBlock(ctx context.Context) error          { return nil }
func (am AppModule) EndBlock(_ sdk.Context) []abci.ValidatorUpdate { return []abci.ValidatorUpdate{} }

func (AppModule) GenerateGenesisState(_ *module.SimulationState)               {}
func (am AppModule) RegisterStoreDecoder(_ interface{})                        {}
func (am AppModule) WeightedOperations(_ module.SimulationState) []interface{} { return nil }
func (AppModule) ConsensusVersion() uint64                                     { return 1 }
func (am AppModule) IsAppModule()                                              {}
func (am AppModule) IsOnePerModuleType()                                       {}
