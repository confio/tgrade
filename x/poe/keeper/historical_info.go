package keeper

import (
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibccoretypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"strconv"
)

var _ ibccoretypes.StakingKeeper = &Keeper{}

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := getHistoricalInfoKey(height)

	value := store.Get(key)
	if value == nil {
		return stakingtypes.HistoricalInfo{}, false
	}

	return stakingtypes.MustUnmarshalHistoricalInfo(k.marshaler, value), true
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(ctx sdk.Context, height int64, hi *stakingtypes.HistoricalInfo) {
	store := ctx.KVStore(k.storeKey)
	key := getHistoricalInfoKey(height)
	value := k.marshaler.MustMarshalBinaryBare(hi)
	store.Set(key, value)
}

// DeleteHistoricalInfo deletes the historical info at a given height
func (k Keeper) DeleteHistoricalInfo(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.storeKey)
	key := getHistoricalInfoKey(height)

	store.Delete(key)
}

// iterateHistoricalInfo provides an interator over all stored HistoricalInfo
//  objects. For each HistoricalInfo object, cb will be called. If the cb returns
// true, the iterator will close and stop.
func (k Keeper) iterateHistoricalInfo(ctx sdk.Context, cb func(stakingtypes.HistoricalInfo) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.HistoricalInfoKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		histInfo := stakingtypes.MustUnmarshalHistoricalInfo(k.marshaler, iterator.Value())
		if cb(histInfo) {
			break
		}
	}
}

// getAllHistoricalInfo returns all stored HistoricalInfo objects.
func (k Keeper) getAllHistoricalInfo(ctx sdk.Context) []stakingtypes.HistoricalInfo {
	var infos []stakingtypes.HistoricalInfo

	k.iterateHistoricalInfo(ctx, func(histInfo stakingtypes.HistoricalInfo) bool {
		infos = append(infos, histInfo)
		return false
	})

	return infos
}

// TrackHistoricalInfo saves the latest historical-info and deletes the oldest
// heights that are below pruning height
func (k Keeper) TrackHistoricalInfo(ctx sdk.Context) {
	entryNum := k.HistoricalEntries(ctx)

	// Prune store to ensure we only have parameter-defined historical entries.
	// In most cases, this will involve removing a single historical entry.
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store.
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry.
	for i := ctx.BlockHeight() - int64(entryNum); i >= 0; i-- {
		_, found := k.GetHistoricalInfo(ctx, i)
		if found {
			k.DeleteHistoricalInfo(ctx, i)
		} else {
			break
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return
	}

	// Create HistoricalInfo struct
	var valSet stakingtypes.Validators // not used by IBC so we keep it empty
	historicalEntry := stakingtypes.NewHistoricalInfo(ctx.BlockHeader(), valSet)

	// Set latest HistoricalInfo at current height
	k.SetHistoricalInfo(ctx, ctx.BlockHeight(), &historicalEntry)
}

// getHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func getHistoricalInfoKey(height int64) []byte {
	return append(types.HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}
