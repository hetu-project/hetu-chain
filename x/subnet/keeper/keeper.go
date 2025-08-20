package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/subnet/types"
)

type (
	Keeper struct {
		cdc                  codec.BinaryCodec
		storeKey             storetypes.StoreKey
		memKey               storetypes.StoreKey
		eventKeeper          types.EventKeeper
		blockInflationKeeper types.BlockInflationKeeper
		erc20Keeper          types.ERC20Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	eventKeeper types.EventKeeper,
	blockInflationKeeper types.BlockInflationKeeper,
	erc20Keeper types.ERC20Keeper,
) *Keeper {
	return &Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		eventKeeper:          eventKeeper,
		blockInflationKeeper: blockInflationKeeper,
		erc20Keeper:          erc20Keeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAllSubnets returns all subnets
func (k Keeper) GetAllSubnets(ctx sdk.Context) []types.SubnetInfo {
	netuids := k.eventKeeper.GetAllSubnetNetuids(ctx)
	subnets := make([]types.SubnetInfo, 0, len(netuids))

	for _, netuid := range netuids {
		subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
		if !found {
			continue
		}

		// Convert string amounts to math.Int
		lockedAmount, _ := subnetInfo.GetLockedAmountInt()
		poolInitialTao, _ := subnetInfo.GetPoolInitialTaoInt()
		burnedAmount, _ := subnetInfo.GetBurnedAmountInt()

		subnet := types.SubnetInfo{
			Netuid:         uint32(netuid),
			Owner:          subnetInfo.Owner,
			AlphaToken:     subnetInfo.AlphaToken,
			AmmPool:        subnetInfo.AmmPool,
			Name:           subnetInfo.Name,
			Description:    subnetInfo.Description,
			IsActive:       subnetInfo.IsActive,
			CreatedAt:      subnetInfo.CreatedAt,
			LockedAmount:   lockedAmount,
			PoolInitialTao: poolInitialTao,
			BurnedAmount:   burnedAmount,
		}

		subnets = append(subnets, subnet)
	}

	return subnets
}

// GetSubnet returns a subnet by ID
func (k Keeper) GetSubnet(ctx sdk.Context, netuid uint16) (types.SubnetInfo, types.SubnetHyperparams, bool) {
	subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
	if !found {
		return types.SubnetInfo{}, types.SubnetHyperparams{}, false
	}

	// Convert string amounts to math.Int
	lockedAmount, _ := subnetInfo.GetLockedAmountInt()
	poolInitialTao, _ := subnetInfo.GetPoolInitialTaoInt()
	burnedAmount, _ := subnetInfo.GetBurnedAmountInt()

	subnet := types.SubnetInfo{
		Netuid:         uint32(netuid),
		Owner:          subnetInfo.Owner,
		AlphaToken:     subnetInfo.AlphaToken,
		AmmPool:        subnetInfo.AmmPool,
		Name:           subnetInfo.Name,
		Description:    subnetInfo.Description,
		IsActive:       subnetInfo.IsActive,
		CreatedAt:      subnetInfo.CreatedAt,
		LockedAmount:   lockedAmount,
		PoolInitialTao: poolInitialTao,
		BurnedAmount:   burnedAmount,
	}

	// Get hyperparameters from params
	params := k.blockInflationKeeper.GetParams(ctx)

	// Create hyperparams with default values
	hyperparams := types.SubnetHyperparams{
		Tempo:                     100,     // Default value
		SubnetEmissionValue:       1000000, // Default value
		SubnetOwnerCut:            params.SubnetOwnerCut,
		MaxAllowedValidators:      128,                             // Default value
		MaxAllowedUids:            256,                             // Default value
		ImmunityPeriod:            100,                             // Default value
		MinAllowedWeights:         1,                               // Default value
		MaxWeightLimit:            1000,                            // Default value
		MaxWeightAge:              math.LegacyNewDecWithPrec(1, 1), // 0.1 Default value
		WeightConsensus:           math.LegacyNewDecWithPrec(1, 1), // 0.1 Default value
		WeightMaxAge:              100,                             // Default value
		ScalingLawPower:           math.LegacyNewDecWithPrec(5, 1), // 0.5 Default value
		ValidatorExcludeQuantile:  math.LegacyNewDecWithPrec(5, 2), // 0.05 Default value
		ValidatorPruneLen:         math.LegacyNewDecWithPrec(1, 1), // 0.1 Default value
		ValidatorLogitsDivergence: math.LegacyNewDecWithPrec(1, 1), // 0.1 Default value
		BlocksSinceLastStep:       0,                               // Default value
		LastMechanismStepBlock:    0,                               // Default value
		BlocksPerStep:             100,                             // Default value
		BondsMovingAverage:        10,                              // Default value
		SubnetMovingAlpha:         params.SubnetMovingAlpha,
		EmaPriceHalvingBlocks:     1000, // Default value
	}

	return subnet, hyperparams, true
}

// GetSubnetNeurons returns all neurons in a subnet
func (k Keeper) GetSubnetNeurons(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	neurons := k.eventKeeper.GetAllNeuronsBySubnet(ctx, netuid)
	result := make([]types.NeuronInfo, 0, len(neurons))

	for _, neuron := range neurons {
		// Convert string stake to math.Int
		stake, _ := neuron.GetStakeInt()

		// Default values for fields not in the event module's NeuronInfo
		emission := math.NewInt(0)
		incentive := math.NewInt(0)
		trust := math.LegacyNewDecWithPrec(0, 0)
		consensus := math.LegacyNewDecWithPrec(0, 0)
		dividends := math.NewInt(0)

		neuronInfo := types.NeuronInfo{
			Uid:        neuron.Account, // Using Account as UID
			Hotkey:     neuron.Account, // Using Account as Hotkey
			Coldkey:    "",             // No coldkey in event module's NeuronInfo
			Stake:      stake,
			LastUpdate: neuron.LastUpdate,
			Rank:       0, // Default value
			Emission:   emission,
			Incentive:  incentive,
			Trust:      trust,
			Consensus:  consensus,
			Dividends:  dividends,
			IsActive:   neuron.IsActive,
		}

		result = append(result, neuronInfo)
	}

	return result
}

// GetSubnetPool returns pool info for a subnet
func (k Keeper) GetSubnetPool(ctx sdk.Context, netuid uint16) (types.PoolInfo, bool) {
	// Check if subnet exists
	_, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return types.PoolInfo{}, false
	}

	// Get pool info
	taoIn := k.eventKeeper.GetSubnetTAO(ctx, netuid)
	alphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
	alphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)
	currentPrice := k.eventKeeper.GetAlphaPrice(ctx, netuid)
	movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)

	// TODO: Get total volume and mechanism type from contract
	totalVolume := math.NewInt(0)
	mechanismType := uint32(0)

	poolInfo := types.PoolInfo{
		Netuid:        uint32(netuid),
		TaoIn:         taoIn,
		AlphaIn:       alphaIn,
		AlphaOut:      alphaOut,
		CurrentPrice:  currentPrice,
		MovingPrice:   movingPrice,
		TotalVolume:   totalVolume,
		MechanismType: mechanismType,
	}

	return poolInfo, true
}
