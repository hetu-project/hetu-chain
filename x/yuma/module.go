package yuma

import (
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/hetu-project/hetu/v1/x/yuma/keeper"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic 定义模块的基本接口
type AppModuleBasic struct{}

// Name 返回模块名称
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec 注册 legacy amino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// 简化的模块，暂时不需要注册消息类型
}

// RegisterInterfaces 注册接口
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// 简化的模块，暂时不需要注册接口
}

// DefaultGenesis 返回默认的 genesis 状态
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// 简化的模块，返回空的 genesis 状态
	return json.RawMessage(`{}`)
}

// ValidateGenesis 验证 genesis 状态
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	// 简化的模块，暂时不需要验证
	return nil
}

// RegisterGRPCGatewayRoutes 注册 gRPC gateway 路由
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// 简化的模块，暂时不需要 gRPC 路由
}

// GetTxCmd 返回交易命令
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	// 简化的模块，暂时没有交易命令
	return nil
}

// GetQueryCmd 返回查询命令
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	// 简化的模块，暂时没有查询命令
	return nil
}

// AppModule 实现 AppModule 接口
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule 创建新的 AppModule
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// RegisterServices 注册服务
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// 简化的模块，暂时不需要注册服务
}

// RegisterInvariants 注册不变量
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// 简化的模块不需要复杂的不变量检查
}

// InitGenesis 初始化 genesis 状态
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []interface{} {
	// 简化的初始化，主要设置基本状态
	InitGenesis(ctx, am.keeper)
	return []interface{}{}
}

// ExportGenesis 导出 genesis 状态
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// 简化的导出，返回空状态
	return json.RawMessage(`{}`)
}

// ConsensusVersion 返回共识版本
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock 在每个区块开始时执行
func (am AppModule) BeginBlock(ctx sdk.Context, _ interface{}) {
	// 检查是否需要运行 epoch
	am.checkAndRunEpochs(ctx)
}

// checkAndRunEpochs 检查并运行 epoch
func (am AppModule) checkAndRunEpochs(ctx sdk.Context) {
	// 获取所有子网
	subnets := am.keeper.GetEventKeeper().GetAllSubnets(ctx)

	for _, subnet := range subnets {
		// 尝试运行 epoch
		_, err := am.keeper.RunEpoch(ctx, subnet.Netuid, 1000000) // 默认 1M rao emission
		if err != nil {
			// 如果不是因为时间未到导致的错误，记录日志
			if err != keeper.ErrEpochNotDue {
				ctx.Logger().Error("Failed to run epoch", "netuid", subnet.Netuid, "error", err)
			}
		}
	}
}

// 依赖注入支持

// IsOnePerModuleType 标记为每个模块类型一个实例
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule 标记为应用模块
func (am AppModule) IsAppModule() {}

// ProvideModule 提供模块依赖
func ProvideModule(in depinject.Config) (keeper.Keeper, error) {
	// 这里需要根据实际的依赖注入配置来实现
	// 暂时返回一个简单的实现
	return keeper.Keeper{}, nil
}

// InitGenesis 初始化 genesis 状态
func InitGenesis(ctx sdk.Context, k keeper.Keeper) {
	// 简化的初始化，主要设置基本状态
	// 这里可以设置一些基本的配置
}

// ExportGenesis 导出 genesis 状态
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) interface{} {
	// 简化的导出，返回空状态
	return nil
}
