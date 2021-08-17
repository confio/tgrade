package stakingadapter

import (
	"context"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibccoretypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"time"
)

// StakingAdapter connect to POE contract

var _ wasmtypes.StakingKeeper = &StakingAdapter{}
var _ wasmtypes.DistributionKeeper = &StakingAdapter{}
var _ wasmkeeper.ValidatorSetSource = &StakingAdapter{}
var _ ibccoretypes.StakingKeeper = &StakingAdapter{}
var _ evidencetypes.StakingKeeper = &StakingAdapter{}
var _ slashingtypes.StakingKeeper = &StakingAdapter{}
var _ minttypes.StakingKeeper = &StakingAdapter{}
var _ govtypes.StakingKeeper = &StakingAdapter{}

type poeKeeper interface{}

type StakingAdapter struct {
	k              poeKeeper
	contractKeeper wasmtypes.ContractOpsKeeper
}

func NewStakingAdapter(k poeKeeper, contractKeeper wasmtypes.ContractOpsKeeper) StakingAdapter {
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

func (s StakingAdapter) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	panic("implement me")
}

func (s StakingAdapter) UnbondingTime(ctx sdk.Context) time.Duration {
	panic("implement me")
}

func (s StakingAdapter) ValidatorByConsAddr(ctx sdk.Context, address sdk.ConsAddress) stakingtypes.ValidatorI {
	panic("implement me")
}

func (s StakingAdapter) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) (updates []abci.ValidatorUpdate, err error) {
	ctx.Logger().Error("NOT IMPLEMENTED")
	return nil, nil
}

func (s StakingAdapter) IterateValidators(ctx sdk.Context, f func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	panic("implement me")
}

func (s StakingAdapter) Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI {
	panic("implement me")
}

func (s StakingAdapter) Slash(ctx sdk.Context, address sdk.ConsAddress, i int64, i2 int64, dec sdk.Dec) {
	panic("implement me")
}

func (s StakingAdapter) Jail(ctx sdk.Context, address sdk.ConsAddress) {
	panic("implement me")
}

func (s StakingAdapter) Unjail(ctx sdk.Context, address sdk.ConsAddress) {
	panic("implement me")
}

func (s StakingAdapter) Delegation(ctx sdk.Context, address sdk.AccAddress, address2 sdk.ValAddress) stakingtypes.DelegationI {
	panic("implement me")
}

func (s StakingAdapter) MaxValidators(ctx sdk.Context) uint32 {
	panic("implement me")
}

func (s StakingAdapter) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	panic("implement me")
}

func (s StakingAdapter) BondedRatio(ctx sdk.Context) sdk.Dec {
	panic("implement me")
}

func (s2 StakingAdapter) IterateBondedValidatorsByPower(s sdk.Context, f func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	panic("implement me")
}

func (s2 StakingAdapter) TotalBondedTokens(s sdk.Context) sdk.Int {
	panic("implement me")
}

func (s2 StakingAdapter) IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool)) {
	panic("implement me")
}
