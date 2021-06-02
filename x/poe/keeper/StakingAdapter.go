package keeper

import (
	"context"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingAdapter connect to POE contract

var _ wasmtypes.StakingKeeper = &StakingAdapter{}
var _ wasmtypes.DistributionKeeper = &StakingAdapter{}

type TWasmKeeper interface {
}

type StakingAdapter struct {
	twasmKeeper TWasmKeeper
}

func NewStakingAdapter(twasmKeeper TWasmKeeper) StakingAdapter {
	return StakingAdapter{twasmKeeper: twasmKeeper}
}

func (s StakingAdapter) BondDenom(ctx sdk.Context) (res string) {
	panic("implement me")
}

func (s StakingAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	panic("implement me")
}

func (s StakingAdapter) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	panic("implement me")
}

func (s StakingAdapter) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	panic("implement me")
}

func (s StakingAdapter) GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool) {
	panic("implement me")
}

func (s StakingAdapter) HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {
	panic("implement me")
}

func (s StakingAdapter) DelegationRewards(c context.Context, req *types.QueryDelegationRewardsRequest) (*types.QueryDelegationRewardsResponse, error) {
	panic("implement me")
}
