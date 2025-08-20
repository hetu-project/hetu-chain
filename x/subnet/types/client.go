package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	// Subnets returns all subnets with pagination
	Subnets(ctx context.Context, in *QuerySubnetsRequest, opts ...interface{}) (*QuerySubnetsResponse, error)
	// Subnet returns subnet info by ID
	Subnet(ctx context.Context, in *QuerySubnetRequest, opts ...interface{}) (*QuerySubnetResponse, error)
	// SubnetNeurons returns all neurons in a subnet with pagination
	SubnetNeurons(ctx context.Context, in *QuerySubnetNeuronsRequest, opts ...interface{}) (*QuerySubnetNeuronsResponse, error)
	// SubnetPool returns pool info for a subnet
	SubnetPool(ctx context.Context, in *QuerySubnetPoolRequest, opts ...interface{}) (*QuerySubnetPoolResponse, error)
}

// NewQueryClient creates a new QueryClient instance
// This is a temporary implementation until the proto files are compiled
func NewQueryClient(clientCtx client.Context) QueryClient {
	return nil
}

// RegisterQueryHandlerClient registers the query handler client
// This is a temporary implementation until the proto files are compiled
func RegisterQueryHandlerClient(ctx context.Context, mux *runtime.ServeMux, client QueryClient) error {
	return nil
}
