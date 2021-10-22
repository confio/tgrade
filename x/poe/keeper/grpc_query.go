package keeper

import (
	"context"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

var _ types.QueryServer = &grpcQuerier{}

// ContractSource subset of poe keeper
type ContractSource interface {
	GetPoEContractAddress(sdk.Context, types.PoEContractType) (sdk.AccAddress, error)
}

type ViewKeeper interface {
	ContractSource
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetBondDenom(ctx sdk.Context) string
}
type grpcQuerier struct {
	keeper          ViewKeeper
	contractQuerier types.SmartQuerier
}

// NewGrpcQuerier constructor
func NewGrpcQuerier(keeper ViewKeeper, contractQuerier types.SmartQuerier) *grpcQuerier {
	return &grpcQuerier{keeper: keeper, contractQuerier: contractQuerier}
}

// ContractAddress query PoE contract address for given type
func (q grpcQuerier) ContractAddress(c context.Context, req *types.QueryContractAddressRequest) (*types.QueryContractAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := q.keeper.GetPoEContractAddress(sdk.UnwrapSDKContext(c), req.ContractType)
	switch {
	case wasmtypes.ErrNotFound.Is(err):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryContractAddressResponse{Address: addr.String()}, nil
}

// Validators query all validators that match the given status.
func (q grpcQuerier) Validators(c context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Pagination != nil {
		return nil, status.Error(codes.Unimplemented, "pagination not supported, yet")
	}
	if req.Status != "" {
		return nil, status.Error(codes.Unimplemented, "status not supported, yet")
	}

	ctx := sdk.UnwrapSDKContext(c)
	addr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	valsRsp, err := contract.ListValidators(ctx, q.contractQuerier, addr)
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
func (q grpcQuerier) Validator(c context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
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
	contractAddr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	valRsp, err := contract.QueryValidator(ctx, q.contractQuerier, contractAddr, opAddr)
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
func (q grpcQuerier) UnbondingPeriod(c context.Context, req *types.QueryUnbondingPeriodRequest) (*types.QueryUnbondingPeriodResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	contractAddr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp, err := contract.QueryStakingUnbondingPeriod(ctx, q.contractQuerier, contractAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &types.QueryUnbondingPeriodResponse{
		Time: time.Duration(rsp) * time.Second,
	}, nil
}

func (q grpcQuerier) ValidatorDelegation(c context.Context, req *types.QueryValidatorDelegationRequest) (*types.QueryValidatorDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	stakingContractAddr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	amount, err := contract.QueryTG4Member(ctx, q.contractQuerier, stakingContractAddr, opAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if amount == nil {
		return nil, status.Error(codes.NotFound, "not a validator operator address")
	}
	return &types.QueryValidatorDelegationResponse{
		Balance: sdk.NewCoin(q.keeper.GetBondDenom(ctx), sdk.NewInt(int64(*amount))),
	}, nil
}

func (q grpcQuerier) ValidatorUnbondingDelegations(c context.Context, req *types.QueryValidatorUnbondingDelegationsRequest) (*types.QueryValidatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	stakingContractAddr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := contract.QueryStakingUnbonding(ctx, q.contractQuerier, stakingContractAddr, opAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// add all unbonded amounts
	var unbodings []stakingtypes.UnbondingDelegationEntry
	for _, v := range res.Claims {
		unbodings = append(unbodings, stakingtypes.UnbondingDelegationEntry{
			InitialBalance: v.Amount,
			CompletionTime: time.Unix(0, int64(v.ReleaseAt)).UTC(),
			Balance:        v.Amount,
			CreationHeight: int64(v.CreationHeight),
		})
	}
	return &types.QueryValidatorUnbondingDelegationsResponse{Entries: unbodings}, nil
}

func (q grpcQuerier) HistoricalInfo(c context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	hi, found := q.keeper.GetHistoricalInfo(sdk.UnwrapSDKContext(c), req.Height)
	if !found {
		return nil, status.Errorf(codes.NotFound, "historical info for height %d not found", req.Height)
	}
	return &stakingtypes.QueryHistoricalInfoResponse{Hist: &hi}, nil
}
