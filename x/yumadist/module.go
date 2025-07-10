package yumadist

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/hetu-project/hetu-chain/x/yumadist/keeper"
)

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string                                       { return "yumadist" }
func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino)               {}
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage { return nil }
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

func (am AppModule) Name() string                                        { return am.AppModuleBasic.Name() }
func (am AppModule) RegisterServices(cfg module.Configurator)            {}
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry)         {}
func (am AppModule) Route() sdk.Route                                    { return sdk.Route{} }
func (am AppModule) QuerierRoute() string                                { return "yumadist" }
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier { return nil }
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage { return nil }
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock)             {}
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
