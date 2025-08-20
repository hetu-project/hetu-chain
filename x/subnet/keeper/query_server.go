package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hetu-project/hetu/v1/x/subnet/types"
)

var _ types.QueryServer = Keeper{}

// Subnets implements the Query/Subnets gRPC method
func (k Keeper) Subnets(c context.Context, req *types.QuerySubnetsRequest) (*types.QuerySubnetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	allSubnets := k.GetAllSubnets(ctx)

	// Apply pagination
	start, end := pagination(req.Pagination, len(allSubnets))
	if start >= len(allSubnets) {
		allSubnets = []types.SubnetInfo{}
	} else if end >= len(allSubnets) {
		allSubnets = allSubnets[start:]
	} else {
		allSubnets = allSubnets[start:end]
	}

	return &types.QuerySubnetsResponse{
		Subnets:    allSubnets,
		Pagination: req.Pagination,
	}, nil
}

// Subnet implements the Query/Subnet gRPC method
func (k Keeper) Subnet(c context.Context, req *types.QuerySubnetRequest) (*types.QuerySubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	subnet, hyperparams, found := k.GetSubnet(ctx, uint16(req.Netuid))
	if !found {
		return nil, status.Errorf(codes.NotFound, "subnet with ID %d not found", req.Netuid)
	}

	return &types.QuerySubnetResponse{
		Subnet:      subnet,
		Hyperparams: hyperparams,
	}, nil
}

// SubnetNeurons implements the Query/SubnetNeurons gRPC method
func (k Keeper) SubnetNeurons(c context.Context, req *types.QuerySubnetNeuronsRequest) (*types.QuerySubnetNeuronsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Check if subnet exists
	_, _, found := k.GetSubnet(ctx, uint16(req.Netuid))
	if !found {
		return nil, status.Errorf(codes.NotFound, "subnet with ID %d not found", req.Netuid)
	}

	// Get all neurons for the subnet
	allNeurons := k.GetSubnetNeurons(ctx, uint16(req.Netuid))

	// Apply pagination
	start, end := pagination(req.Pagination, len(allNeurons))
	if start >= len(allNeurons) {
		allNeurons = []types.NeuronInfo{}
	} else if end >= len(allNeurons) {
		allNeurons = allNeurons[start:]
	} else {
		allNeurons = allNeurons[start:end]
	}

	return &types.QuerySubnetNeuronsResponse{
		Neurons:    allNeurons,
		Pagination: req.Pagination,
	}, nil
}

// SubnetPool implements the Query/SubnetPool gRPC method
func (k Keeper) SubnetPool(c context.Context, req *types.QuerySubnetPoolRequest) (*types.QuerySubnetPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	poolInfo, found := k.GetSubnetPool(ctx, uint16(req.Netuid))
	if !found {
		return nil, status.Errorf(codes.NotFound, "subnet with ID %d not found", req.Netuid)
	}

	return &types.QuerySubnetPoolResponse{
		Pool: poolInfo,
	}, nil
}

// Helper function to handle pagination
func pagination(pageReq *query.PageRequest, total int) (start int, end int) {
	if pageReq == nil {
		return 0, total
	}

	start = int(pageReq.Offset)
	if start < 0 {
		start = 0
	}

	limit := int(pageReq.Limit)
	if limit <= 0 {
		limit = 100 // Default limit
	}

	end = start + limit
	return start, end
}
