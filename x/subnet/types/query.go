package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
)

// QueryServer defines the gRPC querier service for subnet module
type QueryServer interface {
	// Subnets returns all subnets with pagination
	Subnets(context.Context, *QuerySubnetsRequest) (*QuerySubnetsResponse, error)
	// Subnet returns subnet info by ID
	Subnet(context.Context, *QuerySubnetRequest) (*QuerySubnetResponse, error)
	// SubnetNeurons returns all neurons in a subnet with pagination
	SubnetNeurons(context.Context, *QuerySubnetNeuronsRequest) (*QuerySubnetNeuronsResponse, error)
	// SubnetPool returns pool info for a subnet
	SubnetPool(context.Context, *QuerySubnetPoolRequest) (*QuerySubnetPoolResponse, error)
}

// QuerySubnetsRequest is request type for the Query/Subnets RPC method
type QuerySubnetsRequest struct {
	// pagination defines an optional pagination for the request.
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QuerySubnetsResponse is response type for the Query/Subnets RPC method
type QuerySubnetsResponse struct {
	Subnets    []SubnetInfo        `json:"subnets"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QuerySubnetRequest is request type for the Query/Subnet RPC method
type QuerySubnetRequest struct {
	Netuid uint32 `json:"netuid"`
}

// QuerySubnetResponse is response type for the Query/Subnet RPC method
type QuerySubnetResponse struct {
	Subnet      SubnetInfo        `json:"subnet"`
	Hyperparams SubnetHyperparams `json:"hyperparams"`
}

// QuerySubnetNeuronsRequest is request type for the Query/SubnetNeurons RPC method
type QuerySubnetNeuronsRequest struct {
	Netuid     uint32             `json:"netuid"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QuerySubnetNeuronsResponse is response type for the Query/SubnetNeurons RPC method
type QuerySubnetNeuronsResponse struct {
	Neurons    []NeuronInfo        `json:"neurons"`
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QuerySubnetPoolRequest is request type for the Query/SubnetPool RPC method
type QuerySubnetPoolRequest struct {
	Netuid uint32 `json:"netuid"`
}

// QuerySubnetPoolResponse is response type for the Query/SubnetPool RPC method
type QuerySubnetPoolResponse struct {
	Pool PoolInfo `json:"pool"`
}
