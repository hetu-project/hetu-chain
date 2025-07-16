package types

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryParamsRequest is the request type for the Query/Params RPC method.
type QueryParamsRequest struct{}

// QueryParamsResponse is the response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	// Params defines the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

// QueryPendingSubnetRewardsRequest is the request type for the Query/PendingSubnetRewards RPC method.
type QueryPendingSubnetRewardsRequest struct{}

// QueryPendingSubnetRewardsResponse is the response type for the Query/PendingSubnetRewards RPC method.
type QueryPendingSubnetRewardsResponse struct {
	// PendingSubnetRewards defines the pending subnet rewards.
	PendingSubnetRewards string `protobuf:"bytes,1,opt,name=pending_subnet_rewards,json=pendingSubnetRewards,proto3" json:"pending_subnet_rewards,omitempty"`
}

// QueryClient is the client API for Query service.
type QueryClient interface {
	// Params queries the parameters of the module.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error)
	// PendingSubnetRewards queries the pending subnet rewards.
	PendingSubnetRewards(ctx context.Context, in *QueryPendingSubnetRewardsRequest, opts ...interface{}) (*QueryPendingSubnetRewardsResponse, error)
}

// NewQueryClient creates a new query client
func NewQueryClient(client interface{}) QueryClient {
	return &queryClient{client: client}
}

type queryClient struct {
	client interface{}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...interface{}) (*QueryParamsResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, this would make a gRPC call
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}

func (c *queryClient) PendingSubnetRewards(ctx context.Context, in *QueryPendingSubnetRewardsRequest, opts ...interface{}) (*QueryPendingSubnetRewardsResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, this would make a gRPC call
	return nil, status.Errorf(codes.Unimplemented, "method PendingSubnetRewards not implemented")
}
