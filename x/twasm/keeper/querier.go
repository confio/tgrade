package keeper

import (
	"context"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = &grpcQuerier{}

// queryKeeper is a subset of the keeper's methods
type queryKeeper interface {
	IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool)
	IterateContractCallbacksByType(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}
type grpcQuerier struct {
	keeper queryKeeper
}

func NewGrpcQuerier(keeper queryKeeper) *grpcQuerier {
	return &grpcQuerier{keeper: keeper}
}

func (q grpcQuerier) PrivilegedContracts(c context.Context, _ *types.QueryPrivilegedContractsRequest) (*types.QueryPrivilegedContractsResponse, error) {
	var result types.QueryPrivilegedContractsResponse
	q.keeper.IteratePrivileged(sdk.UnwrapSDKContext(c), func(address sdk.AccAddress) bool {
		result.Contracts = append(result.Contracts, address.String())
		return false
	})
	return &result, nil
}

func (q grpcQuerier) ContractsByCallbackType(c context.Context, req *types.QueryContractsByCallbackTypeRequest) (*types.QueryContractsByCallbackTypeResponse, error) {
	var result types.QueryContractsByCallbackTypeResponse
	cType := types.PrivilegedCallbackTypeFrom(req.CallbackType)
	if cType == nil {
		return nil, status.Error(codes.NotFound, "callback type")
	}
	q.keeper.IterateContractCallbacksByType(sdk.UnwrapSDKContext(c), *cType, func(_ uint8, contractAddr sdk.AccAddress) bool {
		result.Contracts = append(result.Contracts, contractAddr.String())
		return false
	})
	return &result, nil
}
