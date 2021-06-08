package keeper

import (
	"context"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = &grpcQuerier{}

type grpcQuerier struct {
	keeper ContractSource
}

// NewGrpcQuerier constructor
func NewGrpcQuerier(keeper ContractSource) *grpcQuerier {
	return &grpcQuerier{keeper: keeper}
}

// ContractAddress query PoE contract address for given type
func (g grpcQuerier) ContractAddress(c context.Context, request *types.QueryContractAddressRequest) (*types.QueryContractAddressResponse, error) {
	addr, err := g.keeper.GetPoEContractAddress(sdk.UnwrapSDKContext(c), request.ContractType)
	switch {
	case wasmtypes.ErrNotFound.Is(err):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryContractAddressResponse{Address: addr.String()}, nil
}
