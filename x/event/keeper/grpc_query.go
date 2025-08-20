package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	eventtypes "github.com/hetu-project/hetu/v1/hetu/event/v1"
)

// QueryServer implements the gRPC QueryServer interface
type QueryServer struct {
	Keeper
	eventtypes.UnimplementedQueryServer
}

// NewQueryServer creates a new QueryServer instance
func NewQueryServer(k Keeper) eventtypes.QueryServer {
	return &QueryServer{
		Keeper: k,
	}
}

// Subnets implements the Query/Subnets gRPC method
func (q QueryServer) Subnets(ctx context.Context, req *eventtypes.QuerySubnetsRequest) (*eventtypes.QuerySubnetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnets := q.Keeper.GetAllSubnetInfos(sdkCtx)

	// Convert to protobuf types
	pbSubnets := make([]*eventtypes.SubnetInfo, len(subnets))
	for i, subnet := range subnets {
		pbSubnets[i] = &eventtypes.SubnetInfo{
			Netuid:         uint32(subnet.Netuid),
			Owner:          subnet.Owner,
			AlphaToken:     subnet.AlphaToken,
			AmmPool:        subnet.AmmPool,
			LockedAmount:   subnet.LockedAmount,
			PoolInitialTao: subnet.PoolInitialTao,
			BurnedAmount:   subnet.BurnedAmount,
			CreatedAt:      subnet.CreatedAt,
			IsActive:       subnet.IsActive,
			Name:           subnet.Name,
			Description:    subnet.Description,
			ActivatedAt:    subnet.ActivatedAt,
			ActivatedBlock: subnet.ActivatedBlock,
		}
	}

	return &eventtypes.QuerySubnetsResponse{
		Subnets: pbSubnets,
	}, nil
}

// Subnet implements the Query/Subnet gRPC method
func (q QueryServer) Subnet(ctx context.Context, req *eventtypes.QuerySubnetRequest) (*eventtypes.QuerySubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnet, found := q.Keeper.GetSubnetInfo(sdkCtx, uint16(req.Netuid))
	if !found {
		return nil, status.Errorf(codes.NotFound, "subnet with netuid %d not found", req.Netuid)
	}

	// Convert to protobuf type
	pbSubnet := &eventtypes.SubnetInfo{
		Netuid:         uint32(subnet.Netuid),
		Owner:          subnet.Owner,
		AlphaToken:     subnet.AlphaToken,
		AmmPool:        subnet.AmmPool,
		LockedAmount:   subnet.LockedAmount,
		PoolInitialTao: subnet.PoolInitialTao,
		BurnedAmount:   subnet.BurnedAmount,
		CreatedAt:      subnet.CreatedAt,
		IsActive:       subnet.IsActive,
		Name:           subnet.Name,
		Description:    subnet.Description,
		ActivatedAt:    subnet.ActivatedAt,
		ActivatedBlock: subnet.ActivatedBlock,
	}

	return &eventtypes.QuerySubnetResponse{
		Subnet: pbSubnet,
	}, nil
}

// SubnetNeurons implements the Query/SubnetNeurons gRPC method
func (q QueryServer) SubnetNeurons(ctx context.Context, req *eventtypes.QuerySubnetNeuronsRequest) (*eventtypes.QuerySubnetNeuronsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	neurons := q.Keeper.GetAllNeuronInfosByNetuid(sdkCtx, uint16(req.Netuid))

	// Convert to protobuf types
	pbNeurons := make([]*eventtypes.NeuronInfo, len(neurons))
	for i, neuron := range neurons {
		pbNeurons[i] = &eventtypes.NeuronInfo{
			Account:                neuron.Account,
			Netuid:                 uint32(neuron.Netuid),
			IsActive:               neuron.IsActive,
			IsValidator:            neuron.IsValidator,
			RequestedValidatorRole: neuron.RequestedValidatorRole,
			Stake:                  neuron.Stake,
			RegistrationBlock:      neuron.RegistrationBlock,
			LastUpdate:             neuron.LastUpdate,
			AxonEndpoint:           neuron.AxonEndpoint,
			AxonPort:               neuron.AxonPort,
			PrometheusEndpoint:     neuron.PrometheusEndpoint,
			PrometheusPort:         neuron.PrometheusPort,
		}
	}

	return &eventtypes.QuerySubnetNeuronsResponse{
		Neurons: pbNeurons,
	}, nil
}

// SubnetPool implements the Query/SubnetPool gRPC method
func (q QueryServer) SubnetPool(ctx context.Context, req *eventtypes.QuerySubnetPoolRequest) (*eventtypes.QuerySubnetPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnet, found := q.Keeper.GetSubnet(sdkCtx, uint16(req.Netuid))
	if !found {
		return nil, status.Errorf(codes.NotFound, "subnet with netuid %d not found", req.Netuid)
	}

	return &eventtypes.QuerySubnetPoolResponse{
		Netuid:       uint32(subnet.Netuid),
		AmmPool:      subnet.AmmPool,
		LockedAmount: subnet.LockedAmount,
		BurnedAmount: subnet.BurnedAmount,
	}, nil
}

// ValidatorWeights implements the Query/ValidatorWeights gRPC method
func (q QueryServer) ValidatorWeights(ctx context.Context, req *eventtypes.QueryValidatorWeightsRequest) (*eventtypes.QueryValidatorWeightsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	weight, found := q.Keeper.GetValidatorWeight(sdkCtx, uint16(req.Netuid), req.Validator)
	if !found {
		return nil, status.Errorf(codes.NotFound, "validator weight for netuid %d and validator %s not found", req.Netuid, req.Validator)
	}

	// 将 map[string]uint64 转换为 []string
	weights := make([]string, 0, len(weight.Weights))
	for uid, w := range weight.Weights {
		weights = append(weights, fmt.Sprintf("%s:%d", uid, w))
	}

	return &eventtypes.QueryValidatorWeightsResponse{
		Netuid:    uint32(weight.Netuid),
		Validator: weight.Validator,
		Weights:   weights,
	}, nil
}
