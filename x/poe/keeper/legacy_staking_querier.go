package keeper

import (
	"context"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var _ stakingtypes.QueryServer = &legacyStakingGRPCQuerier{}

var neverReleasedDelegation = time.Date(2999, time.December, 31, 12, 0, 0, 0, time.UTC)

type stakingQuerierKeeper interface {
	ContractSource
	GetBondDenom(ctx sdk.Context) string
	HistoricalEntries(ctx sdk.Context) uint32
	UnbondingTime(ctx sdk.Context) time.Duration
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
}
type legacyStakingGRPCQuerier struct {
	keeper          stakingQuerierKeeper
	contractQuerier types.SmartQuerier
	queryServer     types.QueryServer
}

func NewLegacyStakingGRPCQuerier(poeKeeper stakingQuerierKeeper, q types.SmartQuerier) *legacyStakingGRPCQuerier {
	return &legacyStakingGRPCQuerier{keeper: poeKeeper, contractQuerier: q, queryServer: NewGrpcQuerier(poeKeeper, q)}
}

// Validators legacy support for querying all validators that match the given status
func (q legacyStakingGRPCQuerier) Validators(c context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	return q.queryServer.Validators(c, req)
}

// Validator legacy support for querying the validator info for a given validator address.
// returns NotFound error code when none exists for the given address
func (q legacyStakingGRPCQuerier) Validator(c context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	return q.queryServer.Validator(c, req)
}

// ValidatorDelegations legacy support for querying the delegate infos for a given validator.
// In PoE only validator operators do self delegations/ unbondings. Result set is either zero or one element.
func (q legacyStakingGRPCQuerier) ValidatorDelegations(c context.Context, req *stakingtypes.QueryValidatorDelegationsRequest) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
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
		return &stakingtypes.QueryValidatorDelegationsResponse{}, nil
	}
	return &stakingtypes.QueryValidatorDelegationsResponse{
		DelegationResponses: stakingtypes.DelegationResponses{
			{
				Delegation: stakingtypes.Delegation{
					DelegatorAddress: opAddr.String(),
					ValidatorAddress: opAddr.String(),
					Shares:           sdk.OneDec(),
				},
				Balance: sdk.NewCoin(q.keeper.GetBondDenom(ctx), sdk.NewInt(int64(*amount))),
			},
		},
	}, nil
}

// ValidatorUnbondingDelegations legacy support for querying the unbonding delegations of a validator.
// In PoE only validator operators do self delegations/ unbondings. Result set is either zero or one element.
func (q legacyStakingGRPCQuerier) ValidatorUnbondingDelegations(c context.Context, req *stakingtypes.QueryValidatorUnbondingDelegationsRequest) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
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
		var compl time.Time
		switch {
		case v.ReleaseAt.AtTime != nil:
			compl = time.Unix(0, int64(*v.ReleaseAt.AtTime)).UTC()
		case v.ReleaseAt.Never != nil:
			compl = neverReleasedDelegation
		case v.ReleaseAt.AtHeight != nil:
			// unhandled
		}
		unbodings = append(unbodings, stakingtypes.UnbondingDelegationEntry{
			CompletionTime: compl,
			Balance:        v.Amount,
		})
	}
	result := &stakingtypes.QueryValidatorUnbondingDelegationsResponse{UnbondingResponses: []stakingtypes.UnbondingDelegation{
		{
			DelegatorAddress: req.ValidatorAddr,
			ValidatorAddress: req.ValidatorAddr,
			Entries:          unbodings,
		},
	}}
	return result, nil
}

// Delegation legacy support for querying the delegate info for a given validator delegator pair
// Returns response or NotFound error when none exists.
func (q legacyStakingGRPCQuerier) Delegation(c context.Context, req *stakingtypes.QueryDelegationRequest) (*stakingtypes.QueryDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	if req.ValidatorAddr != req.DelegatorAddr { // return early on impossible case
		return nil, status.Error(
			codes.NotFound,
			"delegation with delegator not found",
		)

	}
	qr, err := q.ValidatorDelegations(c, &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: req.ValidatorAddr})
	if err != nil {
		return nil, err
	}
	if n := len(qr.DelegationResponses); n == 0 {
		return nil, status.Error(
			codes.NotFound,
			"delegation for delegator not found",
		)
	}
	return &stakingtypes.QueryDelegationResponse{
		DelegationResponse: &qr.DelegationResponses[0],
	}, nil
}

// UnbondingDelegation legacy support for querying the unbonding info for given validator delegator pair
// Returns response or NotFound error when none exists.
func (q legacyStakingGRPCQuerier) UnbondingDelegation(c context.Context, req *stakingtypes.QueryUnbondingDelegationRequest) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	if req.ValidatorAddr != req.DelegatorAddr { // return early on impossible case
		return nil, status.Errorf(
			codes.NotFound,
			"delegation with delegator %s not found for validator %s",
			req.DelegatorAddr, req.ValidatorAddr)

	}
	qr, err := q.ValidatorUnbondingDelegations(c, &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: req.ValidatorAddr})
	if err != nil {
		return nil, err
	}
	if n := len(qr.UnbondingResponses); n == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			"delegation for delegator not found",
		)
	}
	return &stakingtypes.QueryUnbondingDelegationResponse{
		Unbond: qr.UnbondingResponses[0],
	}, nil
}

// DelegatorDelegations legacy support for querying all delegations of a given delegator address.
// In PoE only validator operators do self delegations/ unbondings. Result set is either zero or one element.
func (q legacyStakingGRPCQuerier) DelegatorDelegations(c context.Context, req *stakingtypes.QueryDelegatorDelegationsRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	qr, err := q.ValidatorDelegations(c, &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: req.DelegatorAddr})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.QueryDelegatorDelegationsResponse{
		DelegationResponses: qr.DelegationResponses,
	}, nil
}

// DelegatorUnbondingDelegations legacy support for querying all unbonding delegations of a given delegator address
// In PoE only validator operators do self delegations/ unbondings. Result set is either zero or one element.
func (q legacyStakingGRPCQuerier) DelegatorUnbondingDelegations(c context.Context, req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.DelegatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	qr, err := q.ValidatorUnbondingDelegations(c, &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: req.DelegatorAddr})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingResponses: qr.UnbondingResponses,
	}, nil

}

func (q legacyStakingGRPCQuerier) Redelegations(c context.Context, req *stakingtypes.QueryRedelegationsRequest) (*stakingtypes.QueryRedelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &stakingtypes.QueryRedelegationsResponse{}, nil
}

func (q legacyStakingGRPCQuerier) DelegatorValidators(c context.Context, req *stakingtypes.QueryDelegatorValidatorsRequest) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	resp, err := q.Validator(c, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: req.DelegatorAddr,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.QueryDelegatorValidatorsResponse{
		Validators: []stakingtypes.Validator{resp.Validator},
	}, nil
}

func (q legacyStakingGRPCQuerier) DelegatorValidator(c context.Context, req *stakingtypes.QueryDelegatorValidatorRequest) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	resp, err := q.Validator(c, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: req.DelegatorAddr,
	})
	if err != nil {
		return nil, err
	}
	return &stakingtypes.QueryDelegatorValidatorResponse{
		Validator: resp.Validator,
	}, nil

}

// HistoricalInfo queries the historical info for given height
func (q legacyStakingGRPCQuerier) HistoricalInfo(c context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	hi, found := q.keeper.GetHistoricalInfo(sdk.UnwrapSDKContext(c), req.Height)
	if !found {
		return nil, status.Errorf(codes.NotFound, "historical info for height %d not found", req.Height)
	}
	return &stakingtypes.QueryHistoricalInfoResponse{Hist: &hi}, nil
}

func (q legacyStakingGRPCQuerier) Pool(c context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	logNotImplemented(sdk.UnwrapSDKContext(c), "Pool")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) Params(c context.Context, req *stakingtypes.QueryParamsRequest) (*stakingtypes.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	valsetContractAddr, err := q.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	valsetConfig, err := contract.QueryValsetConfig(ctx, q.contractQuerier, valsetContractAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stakingtypes.QueryParamsResponse{
		Params: stakingtypes.Params{
			UnbondingTime:     q.keeper.UnbondingTime(ctx),
			MaxValidators:     uint32(valsetConfig.MaxValidators),
			MaxEntries:        0,
			HistoricalEntries: q.keeper.HistoricalEntries(ctx),
			BondDenom:         q.keeper.GetBondDenom(ctx),
		},
	}, nil
}

func logNotImplemented(ctx sdk.Context, msg string) {
	ctx.Logger().Error("NOT IMPLEMENTED: ", "fn", msg)
}
