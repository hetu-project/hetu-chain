package blockinflation

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"

	"github.com/hetu-project/hetu/v1/x/blockinflation/client/cli"
	"github.com/hetu-project/hetu/v1/x/blockinflation/keeper"
	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	pb "github.com/hetu-project/hetu/v1/x/blockinflation/types/generated"
	"google.golang.org/grpc"
)

var (
	_ module.AppModule          = AppModule{}
	_ module.AppModuleBasic     = AppModuleBasic{}
	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

// AppModuleBasic defines the basic application module used by the blockinflation module.
type AppModuleBasic struct{}

// Name returns the blockinflation module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the blockinflation module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the blockinflation module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(*types.DefaultGenesisState())
	if err != nil {
		panic(err)
	}
	return bz
}

// ValidateGenesis performs genesis state validation for the blockinflation module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := json.Unmarshal(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the blockinflation module.
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

// GetTxCmd returns the root tx command for the blockinflation module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the blockinflation module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for the blockinflation module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the blockinflation module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterInvariants registers the blockinflation module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Register invariants here if needed
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	pb.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the blockinflation module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(err)
	}
	am.keeper.InitGenesis(ctx, cdc, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the blockinflation
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx, cdc)
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}
	return bz
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns the begin blocker for the blockinflation module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.keeper.BeginBlocker(sdkCtx)
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements the appmodule.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}
