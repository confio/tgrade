package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/confio/tgrade/x/poe/contract"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/confio/tgrade/x/poe/types"
)

var _ types.QueryServer = &Querier{}

// ContractSource subset of poe keeper
type ContractSource interface {
	GetPoEContractAddress(sdk.Context, types.PoEContractType) (sdk.AccAddress, error)
}

type ViewKeeper interface {
	ContractSource
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetBondDenom(ctx sdk.Context) string
	DistributionContract(ctx sdk.Context) DistributionContract
	ValsetContract(ctx sdk.Context) ValsetContract
	StakeContract(ctx sdk.Context) StakeContract
	EngagementContract(ctx sdk.Context) EngagementContract
}

type Querier struct {
	keeper ViewKeeper
}

// NewQuerier constructor
func NewQuerier(keeper ViewKeeper) *Querier {
	return &Querier{keeper: keeper}
}

// ContractAddress query PoE contract address for given type
func (q Querier) ContractAddress(c context.Context, req *types.QueryContractAddressRequest) (*types.QueryContractAddressResponse, error) {
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
func (q Querier) Validators(c context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Status != "" {
		return nil, status.Error(codes.Unimplemented, "status not supported, yet")
	}

	pagination, err := contract.NewPaginator(req.Pagination)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(c)
	vals, cursor, err := q.keeper.ValsetContract(ctx).ListValidators(ctx, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pageResp *query.PageResponse
	if len(cursor) != 0 {
		pageResp = &query.PageResponse{
			NextKey: cursor,
		}
	}
	return &stakingtypes.QueryValidatorsResponse{
		Validators: vals,
		Pagination: pageResp,
	}, nil
}

// Validator queries validator info for a given validator address.
// returns NotFound error code when none exists for the given address
func (q Querier) Validator(c context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
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
	val, err := q.keeper.ValsetContract(ctx).QueryValidator(ctx, opAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if val == nil {
		return nil, status.Error(codes.NotFound, "by address")
	}
	return &stakingtypes.QueryValidatorResponse{Validator: *val}, nil
}

// UnbondingPeriod query the global unbonding period
func (q Querier) UnbondingPeriod(c context.Context, req *types.QueryUnbondingPeriodRequest) (*types.QueryUnbondingPeriodResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	period, err := q.keeper.StakeContract(ctx).QueryStakingUnbondingPeriod(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryUnbondingPeriodResponse{
		Time: period,
	}, nil
}

func (q Querier) ValidatorDelegation(c context.Context, req *types.QueryValidatorDelegationRequest) (*types.QueryValidatorDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	amount, err := q.keeper.StakeContract(ctx).QueryStakedAmount(ctx, opAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if amount == nil {
		return nil, status.Error(codes.NotFound, "not a validator operator address")
	}
	return &types.QueryValidatorDelegationResponse{
		Balance: sdk.NewCoin(q.keeper.GetBondDenom(ctx), *amount),
	}, nil
}

func (q Querier) ValidatorUnbondingDelegations(c context.Context, req *types.QueryValidatorUnbondingDelegationsRequest) (*types.QueryValidatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator address")
	}

	ctx := sdk.UnwrapSDKContext(c)
	unbodings, err := q.keeper.StakeContract(ctx).QueryStakingUnbonding(ctx, opAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &types.QueryValidatorUnbondingDelegationsResponse{Entries: unbodings}, nil
}

func (q Querier) HistoricalInfo(c context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	hi, found := q.keeper.GetHistoricalInfo(sdk.UnwrapSDKContext(c), req.Height)
	if !found {
		return nil, status.Errorf(codes.NotFound, "historical info for height %d not found", req.Height)
	}
	return &stakingtypes.QueryHistoricalInfoResponse{Hist: &hi}, nil
}

func (q Querier) ValidatorOutstandingReward(c context.Context, req *types.QueryValidatorOutstandingRewardRequest) (*types.QueryValidatorOutstandingRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}
	valAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "address invalid")
	}
	ctx := sdk.UnwrapSDKContext(c)
	reward, err := q.keeper.DistributionContract(ctx).ValidatorOutstandingReward(ctx, valAddr)
	if err != nil {
		if types.ErrNotFound.Is(err) {
			return nil, status.Error(codes.NotFound, "address")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorOutstandingRewardResponse{
		Reward: sdk.NewDecCoin(reward.Denom, reward.Amount),
	}, nil
}
