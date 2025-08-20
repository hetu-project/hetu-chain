// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/hetu-project/hetu/v1/x/event/client/cli"
	"github.com/hetu-project/hetu/v1/x/event/keeper"
	"github.com/hetu-project/hetu/v1/x/event/types"
	pb "github.com/hetu-project/hetu/v1/x/event/types/generated"
	"google.golang.org/grpc"
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

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}
func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesisState())
	if err != nil {
		panic(err)
	}
	return bz
}

func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := json.Unmarshal(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := pb.RegisterQueryHandlerFromEndpoint(
		context.Background(),
		mux,
		"localhost:9090", // Use the gRPC server address directly
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		panic(err)
	}
}

func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

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

// QuerierRoute returns the event module's querier route name
func (am AppModule) QuerierRoute() string {
	return "event"
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	// 注册gRPC查询服务
	pb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))

	// 注册legacy querier
	queryRouter := cfg.QueryServer().(interface {
		RegisterLegacyHandler(string, func(sdk.Context, []string, abci.RequestQuery) ([]byte, error))
	})
	legacyAmino := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(legacyAmino)
	queryRouter.RegisterLegacyHandler(types.ModuleName, keeper.NewLegacyQuerier(am.keeper, legacyAmino))
}

func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (am AppModule) InitGenesis(ctx sdk.Context, _ codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	if err := json.Unmarshal(gs, &genState); err != nil {
		panic(err)
	}

	// Initialize subnet data
	for _, subnet := range genState.Subnets {
		am.keeper.SetSubnet(ctx, subnet)
	}

	// Initialize validator stake data
	for _, stake := range genState.ValidatorStakes {
		am.keeper.SetValidatorStake(ctx, stake)
	}

	// Initialize delegation data
	for _, deleg := range genState.Delegations {
		am.keeper.SetDelegation(ctx, deleg)
	}

	// Initialize validator weight data
	for _, weight := range genState.ValidatorWeights {
		am.keeper.SetValidatorWeight(ctx, weight.Netuid, weight.Validator, weight.Weights)
	}

	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, _ codec.JSONCodec) json.RawMessage {
	// Export all data
	subnets := am.keeper.GetAllSubnets(ctx)

	var validatorStakes []types.ValidatorStake
	for _, subnet := range subnets {
		stakes := am.keeper.GetAllValidatorStakesByNetuid(ctx, subnet.Netuid)
		validatorStakes = append(validatorStakes, stakes...)
	}

	var delegations []types.Delegation
	for _, subnet := range subnets {
		for _, stake := range am.keeper.GetAllValidatorStakesByNetuid(ctx, subnet.Netuid) {
			delegs := am.keeper.GetDelegationsByValidator(ctx, subnet.Netuid, stake.Validator)
			delegations = append(delegations, delegs...)
		}
	}

	var validatorWeights []types.ValidatorWeight
	for _, subnet := range subnets {
		for _, stake := range am.keeper.GetAllValidatorStakesByNetuid(ctx, subnet.Netuid) {
			if weight, found := am.keeper.GetValidatorWeight(ctx, subnet.Netuid, stake.Validator); found {
				validatorWeights = append(validatorWeights, weight)
			}
		}
	}

	genState := types.NewGenesisState(subnets, validatorStakes, delegations, validatorWeights)
	bz, err := json.Marshal(genState)
	if err != nil {
		panic(err)
	}
	return bz
}

func (am AppModule) BeginBlock(ctx context.Context) error          { return nil }
func (am AppModule) EndBlock(_ sdk.Context) []abci.ValidatorUpdate { return []abci.ValidatorUpdate{} }

func (AppModule) GenerateGenesisState(_ *module.SimulationState)               {}
func (am AppModule) RegisterStoreDecoder(_ interface{})                        {}
func (am AppModule) WeightedOperations(_ module.SimulationState) []interface{} { return nil }
func (AppModule) ConsensusVersion() uint64                                     { return 1 }
func (am AppModule) IsAppModule()                                              {}
func (am AppModule) IsOnePerModuleType()                                       {}
