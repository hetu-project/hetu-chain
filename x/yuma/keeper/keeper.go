package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/hetu-project/hetu/v1/x/yuma/types"
)

// Keeper yuma模块的keeper
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	paramstore paramtypes.Subspace

	// 其他模块的keeper依赖
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	eventKeeper   types.EventKeeper // 你的event模块keeper
}

// NewKeeper 创建新的keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	eventKeeper types.EventKeeper,
) *Keeper {
	// 设置参数子空间
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		eventKeeper:   eventKeeper,
	}
}

// Logger 返回模块日志器
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams 获取参数
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramstore.GetParamSet(ctx, &p)
	return p
}

// SetParams 设置参数
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetSubnetInfo 获取子网信息
func (k Keeper) GetSubnetInfo(ctx sdk.Context, netuid uint16) (types.SubnetInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetSubnetInfoKey(netuid))
	if bz == nil {
		return types.SubnetInfo{}, false
	}

	var subnet types.SubnetInfo
	k.cdc.MustUnmarshal(bz, &subnet)
	return subnet, true
}

// SetSubnetInfo 设置子网信息
func (k Keeper) SetSubnetInfo(ctx sdk.Context, subnet types.SubnetInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&subnet)
	store.Set(types.GetSubnetInfoKey(subnet.Netuid), bz)
}

// GetNeuronInfo 获取神经元信息
func (k Keeper) GetNeuronInfo(ctx sdk.Context, netuid uint16, hotkey string) (types.NeuronInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetNeuronInfoKey(netuid, hotkey))
	if bz == nil {
		return types.NeuronInfo{}, false
	}

	var neuron types.NeuronInfo
	k.cdc.MustUnmarshal(bz, &neuron)
	return neuron, true
}

// SetNeuronInfo 设置神经元信息
func (k Keeper) SetNeuronInfo(ctx sdk.Context, neuron types.NeuronInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&neuron)
	store.Set(types.GetNeuronInfoKey(neuron.Netuid, neuron.Hotkey), bz)
}

// GetWeights 获取权重
func (k Keeper) GetWeights(ctx sdk.Context, netuid uint16, uid uint16) []uint16 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetWeightsKey(netuid, uid))
	if bz == nil {
		return []uint16{}
	}

	var weights []uint16
	k.cdc.MustUnmarshal(bz, &weights)
	return weights
}

// SetWeights 设置权重
func (k Keeper) SetWeights(ctx sdk.Context, netuid uint16, uid uint16, weights []uint16) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&weights)
	store.Set(types.GetWeightsKey(netuid, uid), bz)
}

// GetAllSubnets 获取所有子网
func (k Keeper) GetAllSubnets(ctx sdk.Context) []types.SubnetInfo {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.SubnetInfoPrefix)
	defer iterator.Close()

	var subnets []types.SubnetInfo
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.SubnetInfo
		k.cdc.MustUnmarshal(iterator.Value(), &subnet)
		subnets = append(subnets, subnet)
	}

	return subnets
}

// GetSubnetNeurons 获取子网的所有神经元
func (k Keeper) GetSubnetNeurons(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	store := ctx.KVStore(k.storeKey)

	// 构建前缀：NeuronInfoPrefix + netuid
	prefix := append(types.NeuronInfoPrefix, sdk.Uint64ToBigEndian(uint64(netuid))...)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var neurons []types.NeuronInfo
	for ; iterator.Valid(); iterator.Next() {
		var neuron types.NeuronInfo
		k.cdc.MustUnmarshal(iterator.Value(), &neuron)
		neurons = append(neurons, neuron)
	}

	return neurons
}

// GetSubnetSize 获取子网大小（神经元数量）
func (k Keeper) GetSubnetSize(ctx sdk.Context, netuid uint16) uint16 {
	neurons := k.GetSubnetNeurons(ctx, netuid)
	return uint16(len(neurons))
}

// IsValidNeuron 检查神经元是否有效
func (k Keeper) IsValidNeuron(ctx sdk.Context, netuid uint16, hotkey string) bool {
	_, found := k.GetNeuronInfo(ctx, netuid, hotkey)
	return found
}

// GetActiveNeurons 获取活跃的神经元
func (k Keeper) GetActiveNeurons(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	neurons := k.GetSubnetNeurons(ctx, netuid)

	var activeNeurons []types.NeuronInfo

	for _, neuron := range neurons {
		if neuron.Active {
			activeNeurons = append(activeNeurons, neuron)
		}
	}

	return activeNeurons
}

// UpdateSubnetFromEvents 从事件模块更新子网参数
func (k Keeper) UpdateSubnetFromEvents(ctx sdk.Context, netuid uint16) error {
	// 通过event keeper获取子网参数
	subnetParams, err := k.eventKeeper.GetSubnetParameters(ctx, netuid)
	if err != nil {
		return fmt.Errorf("获取子网参数失败: %w", err)
	}

	// 更新子网信息
	subnet, found := k.GetSubnetInfo(ctx, netuid)
	if !found {
		// 如果子网不存在，创建新的
		subnet = types.SubnetInfo{
			Netuid: netuid,
		}
	}

	// 更新参数（这里需要根据你的event模块结构调整）
	subnet.Tempo = subnetParams.Tempo
	subnet.Modality = subnetParams.Modality
	subnet.Owner = subnetParams.Owner
	// ... 其他参数更新

	k.SetSubnetInfo(ctx, subnet)
	return nil
}
