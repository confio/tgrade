package keeper

import (
	"context"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var _ types.QueryServer = &grpcQuerier{}

// ContractSource subset of poe keeper
type ContractSource interface {
	GetPoEContractAddress(sdk.Context, types.PoEContractType) (sdk.AccAddress, error)
}

type grpcQuerier struct {
	keeper          ContractSource
	contractQuerier types.SmartQuerier
}

// NewGrpcQuerier constructor
func NewGrpcQuerier(keeper ContractSource, contractQuerier types.SmartQuerier) *grpcQuerier {
	return &grpcQuerier{keeper: keeper, contractQuerier: contractQuerier}
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

// Validators query all validators that match the given status.
func (g grpcQuerier) Validators(c context.Context, request *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if request.Pagination != nil {
		return nil, status.Error(codes.Unimplemented, "pagination not supported, yet")
	}
	if request.Status != "" {
		return nil, status.Error(codes.Unimplemented, "status not supported, yet")
	}

	ctx := sdk.UnwrapSDKContext(c)
	addr, err := g.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	valsRsp, err := contract.ListValidators(ctx, g.contractQuerier, addr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	vals := make([]stakingtypes.Validator, len(valsRsp))
	for i, v := range valsRsp {
		vals[i], err = v.ToValidator()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &stakingtypes.QueryValidatorsResponse{
		Validators: vals,
	}, nil
}

// Validator queries validator info for a given validator address.
// returns NotFound error code when none exists for the given address
func (g grpcQuerier) Validator(c context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	contractAddr, err := g.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	valRsp, err := contract.QueryValidator(ctx, g.contractQuerier, contractAddr, opAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if valRsp == nil {
		return nil, status.Errorf(codes.NotFound, "validator %s not found", req.ValidatorAddr)
	}
	val, err := valRsp.ToValidator()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &stakingtypes.QueryValidatorResponse{Validator: val}, nil
}

// UnbondingPeriod query the global unbonding period
func (g grpcQuerier) UnbondingPeriod(c context.Context, request *types.QueryUnbondingPeriodRequest) (*types.QueryUnbondingPeriodResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	contractAddr, err := g.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp, err := contract.QueryStakingUnbondingPeriod(ctx, g.contractQuerier, contractAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &types.QueryUnbondingPeriodResponse{
		Height: uint64(rsp.Height),
		Time:   time.Duration(rsp.Time) * time.Second,
	}, nil
}
