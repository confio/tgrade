package keeper

import (
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math"
)

// SetPrivileged does
// - pin to cache
// - set privileged flag
// - call Sudo with PrivilegeChangeMsg{Promoted{}}
func (k Keeper) SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	contractInfo := k.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contractAddr")
	}
	// add to cache
	if err := k.contractKeeper.PinCode(ctx, contractInfo.CodeID); err != nil {
		return sdkerrors.Wrapf(err, "pin")
	}

	// set privileged flag
	k.setPrivilegedFlag(ctx, contractAddr)

	// call contract and let it register for callbacks
	msg := contract.TgradeSudoMsg{PrivilegeChange: &contract.PrivilegeChangeMsg{Promoted: &struct{}{}}}
	msgBz, err := json.Marshal(&msg)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	if _, err = k.Sudo(ctx, contractAddr, msgBz); err != nil {
		return sdkerrors.Wrap(err, "sudo")
	}
	return nil
}

// add to second index for privileged contracts
func (k Keeper) setPrivilegedFlag(ctx sdk.Context, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPrivilegedContractsSecondaryIndexKey(contractAddr), []byte{1})
}

// UnsetPrivileged does:
// - call Sudo with PrivilegeChangeMsg{Demoted{}}
// - remove contract from cache
// - remove privileged flag
// - remove all callbacks for the contract
func (k Keeper) UnsetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	contractInfo := k.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contractAddr")
	}

	// call contract to unregister for callbacks
	msg := contract.TgradeSudoMsg{PrivilegeChange: &contract.PrivilegeChangeMsg{Demoted: &struct{}{}}}
	msgBz, err := json.Marshal(&msg)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	if _, err = k.Sudo(ctx, contractAddr, msgBz); err != nil {
		return sdkerrors.Wrap(err, "sudo")
	}

	// remove from cache
	if err := k.contractKeeper.UnpinCode(ctx, contractInfo.CodeID); err != nil {
		return sdkerrors.Wrapf(err, "unpin")
	}

	// remove privileged flag
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPrivilegedContractsSecondaryIndexKey(contractAddr))

	// iterate callbacks and remove
	k.iterateContractCallbacksByContract(ctx, contractAddr, func(callbackType types.PriviledgedCallbackType, pos uint8) bool {
		k.removePrivilegedContractCallbacks(ctx, callbackType, contractAddr)
		return false
	})
	return nil
}

// IsPrivileged returns if a contract has the privileges flag set.
func (k Keeper) IsPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPrivilegedContractsSecondaryIndexKey(contractAddr))
}

// IteratePrivileged iterates through the list of privileged contacts by type and position ASC
func (k Keeper) IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.PrivilegedContractsSecondaryIndexPrefix)
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		if cb(iter.Key()) {
			return
		}
	}
}

// appendToPrivilegedContractCallbacks registers given contract for a callback type
func (k Keeper) appendToPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PriviledgedCallbackType, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	// find last position value for callback type
	var pos uint8
	it := prefix.NewStore(store, types.GetContractCallbacksSecondaryIndexPrefix(callbackType)).ReverseIterator(nil, nil)
	if it.Valid() {
		key := it.Key()
		pos = key[0]
	}
	newPos := pos + 1
	if newPos <= pos {
		panic("Overflow in in callback positions")
	}
	store.Set(types.GetContractCallbacksSecondaryIndexKey(callbackType, newPos, contractAddress), []byte{1})
}

// removePrivilegedContractCallbacks unregisters the given contract for a callback type
func (k Keeper) removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PriviledgedCallbackType, contractAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	start := append([]byte{0}, contractAddress...)
	end := append([]byte{math.MaxUint8}, contractAddress...)
	prefixStore := prefix.NewStore(store, types.GetContractCallbacksSecondaryIndexPrefix(callbackType))

	for it := prefixStore.Iterator(start, end); it.Valid(); it.Next() {
		prefixStore.Delete(it.Key())
	}
}

// ExistsPrivilegedContractCallbacks returns if any contract is registered for the given type
func (k Keeper) ExistsPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PriviledgedCallbackType) bool {
	store := ctx.KVStore(k.storeKey)

	start := []byte{0}
	end := []byte{math.MaxUint8}
	prefixStore := prefix.NewStore(store, types.GetContractCallbacksSecondaryIndexPrefix(callbackType))

	for it := prefixStore.Iterator(start, end); it.Valid(); it.Next() {
		return true
	}
	return false
}

// IterateContractCallbacksByType iterates through all contracts for the given type by position and address ASC
func (k Keeper) IterateContractCallbacksByType(ctx sdk.Context, callbackType types.PriviledgedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetContractCallbacksSecondaryIndexPrefix(callbackType))
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		pos, address := splitPositionContract(iter.Key())
		// cb returns true to stop early
		if cb(pos, address) {
			return
		}
	}
}

func splitPositionContract(key []byte) (uint8, sdk.AccAddress) {
	return key[0], key[1:]
}

// iterateContractCallbacksByContract iterates through all registered callbacks for the given contract. Ordered by type and position asc
func (k Keeper) iterateContractCallbacksByContract(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(t types.PriviledgedCallbackType, pos uint8) bool) {
	store := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.ContractCallbacksSecondaryIndexPrefix)
	for it := prefixStore.Iterator(nil, nil); it.Valid(); it.Next() {
		t, pos, addr := splitUnprefixedContractCallbacksSecondaryIndexKey(it.Key())
		if addr.Equals(contractAddress) { // index is not optimized for this. so we find all and have to check
			if cb(t, pos) {
				return
			}
		}
	}
}

// splits source of type `<callbackType><position><contractAddr>`
func splitUnprefixedContractCallbacksSecondaryIndexKey(s []byte) (types.PriviledgedCallbackType, uint8, sdk.AccAddress) {
	if len(s) != 1+1+sdk.AddrLen {
		panic(fmt.Sprintf("unexpected key lenght %d", len(s)))
	}
	return types.PriviledgedCallbackType(s[0]), s[1], s[2:]
}
