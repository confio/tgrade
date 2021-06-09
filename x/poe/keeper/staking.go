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

type StakingAdapter struct {
	k              *Keeper
	contractKeeper wasmtypes.ContractOpsKeeper
}

func NewStakingAdapter(k *Keeper, contractKeeper wasmtypes.ContractOpsKeeper) StakingAdapter {
	return StakingAdapter{k: k, contractKeeper: contractKeeper}

}

func (s StakingAdapter) BondDenom(ctx sdk.Context) (res string) {
	panic("implement me") // TODO (Alex): can be satisfied by contract
}

func (s StakingAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	panic("implement me") // TODO (Alex): can be satisfied by contract
}

func (s StakingAdapter) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	panic("implement me") // TODO (Alex): can be satisfied by contract
}

func (s StakingAdapter) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	return nil
}

func (s StakingAdapter) GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool) {
	return
}

func (s StakingAdapter) HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {
	return false
}

func (s StakingAdapter) DelegationRewards(c context.Context, req *types.QueryDelegationRewardsRequest) (*types.QueryDelegationRewardsResponse, error) {
	return nil, nil
}
