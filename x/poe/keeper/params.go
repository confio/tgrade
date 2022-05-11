package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

// HistoricalEntries = number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) (res uint32) {
	k.paramStore.Get(ctx, types.KeyHistoricalEntries, &res)
	return
}

// GetInitialValidatorEngagementPoints get number of engagement for any new validator joining post genesis
func (k Keeper) GetInitialValidatorEngagementPoints(ctx sdk.Context) (res uint64) {
	k.paramStore.Get(ctx, types.KeyInitialValEngagementPoints, &res)
	return
}

func (k Keeper) MinimumDelegationAmounts(ctx sdk.Context) (res sdk.Coins) {
	k.paramStore.Get(ctx, types.KeyMinDelegationAmounts, &res)
	return
}

// GetParams returns all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.HistoricalEntries(ctx),
		k.GetInitialValidatorEngagementPoints(ctx),
		k.MinimumDelegationAmounts(ctx),
	)
}

// SetParams set the params
func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
