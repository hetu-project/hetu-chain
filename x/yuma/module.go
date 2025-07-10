package yuma

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/evmos/evmos/v14/x/yuma/keeper"
	"github.com/evmos/evmos/v14/x/yuma/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic 基础应用模块
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

// Name 返回模块名称
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec 注册legacy amino编解码器
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces 注册接口
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis 返回默认创世状态
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis 验证创世状态
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("解析%s创世状态失败: %w", types.ModuleName, err)
	}
	return types.ValidateGenesis(genState)
}

// RegisterRESTRoutes 注册REST路由
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

// RegisterGRPCGatewayRoutes 注册gRPC网关路由
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd 获取事务命令
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // 如果有交易命令，在这里添加
}

// GetQueryCmd 获取查询命令
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // 如果有查询命令，在这里添加
}

// AppModule 应用模块
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule 创建新的应用模块
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// Name 返回模块名称
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Route 返回模块路由
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

// QuerierRoute 返回查询路由
func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler 返回legacy查询处理器
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices 注册服务
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// RegisterInvariants 注册不变量
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis 初始化创世状态
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis 导出创世状态
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion 返回共识版本
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock 在区块开始时执行
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	// 在每个区块开始时运行epoch检查
	am.runEpochCheck(ctx)
}

// EndBlock 在区块结束时执行
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// GenerateGenesisState 生成创世状态用于模拟
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// 模拟创世状态生成
}

// ProposalContents 返回提案内容
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams 返回随机化参数
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder 注册存储解码器
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations 返回加权操作
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// runEpochCheck 运行epoch检查
func (am AppModule) runEpochCheck(ctx sdk.Context) {
	// 获取所有子网
	subnets := am.keeper.GetAllSubnets(ctx)

	// 为每个子网检查是否需要运行epoch
	for _, subnet := range subnets {
		currentBlock := uint64(ctx.BlockHeight())

		// 检查是否到了运行epoch的时间
		if currentBlock >= subnet.LastEpoch+uint64(subnet.Tempo) {
			err := am.keeper.RunEpoch(ctx, subnet.Netuid)
			if err != nil {
				am.keeper.Logger(ctx).Error("运行epoch失败",
					"netuid", subnet.Netuid,
					"error", err)
			} else {
				am.keeper.Logger(ctx).Info("成功运行epoch",
					"netuid", subnet.Netuid,
					"block", currentBlock)
			}
		}
	}
}

// InitGenesis 初始化创世状态
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// 设置参数
	k.SetParams(ctx, genState.Params)

	// 设置子网
	for _, subnet := range genState.Subnets {
		k.SetSubnetInfo(ctx, subnet)
	}

	// 设置神经元
	for _, neuron := range genState.Neurons {
		k.SetNeuronInfo(ctx, neuron)
	}

	// 设置权重
	for _, weight := range genState.Weights {
		k.SetWeights(ctx, weight.Netuid, weight.Uid, weight.Weights)
	}
}

// ExportGenesis 导出创世状态
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Subnets = k.GetAllSubnets(ctx)

	// 导出所有神经元
	for _, subnet := range genesis.Subnets {
		neurons := k.GetSubnetNeurons(ctx, subnet.Netuid)
		genesis.Neurons = append(genesis.Neurons, neurons...)

		// 导出权重数据
		for _, neuron := range neurons {
			weights := k.GetWeights(ctx, neuron.Netuid, neuron.Uid)
			if len(weights) > 0 {
				genesis.Weights = append(genesis.Weights, types.WeightData{
					Netuid:  neuron.Netuid,
					Uid:     neuron.Uid,
					Weights: weights,
				})
			}
		}
	}

	return *genesis
}

// NewHandler 创建新的处理器
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// 在这里添加消息处理逻辑
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("未识别的%s消息类型: %T", types.ModuleName, msg))
		}
	}
}
