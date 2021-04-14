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

	k.Logger(ctx).Info("Set privileged", "contractAddr", contractAddr.String())
	event := sdk.NewEvent(
		types.EventTypeSetPrivileged,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
	)
	ctx.EventManager().EmitEvent(event)

	return nil
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
	k.clearPrivilegedFlag(ctx, contractAddr)

	// iterate callbacks and remove
	k.iterateContractCallbacksByContract(ctx, contractAddr, func(callbackType types.PrivilegedCallbackType, pos uint8) bool {
		k.removePrivilegedContractCallbacks(ctx, callbackType, contractAddr)
		return false
	})

	k.Logger(ctx).Info("Unset privileged", "contractAddr", contractAddr.String())
	event := sdk.NewEvent(
		types.EventTypeUnsetPrivileged,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
	)
	ctx.EventManager().EmitEvent(event)

	return nil
}

// add to second index for privileged contracts
func (k Keeper) setPrivilegedFlag(ctx sdk.Context, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(privilegedContractsSecondaryIndexKey(contractAddr), []byte{1})
}

// remove entry from second index for privileged contracts
func (k Keeper) clearPrivilegedFlag(ctx sdk.Context, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(privilegedContractsSecondaryIndexKey(contractAddr))
}

// IsPrivileged returns if a contract has the privileges flag set.
func (k Keeper) IsPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(privilegedContractsSecondaryIndexKey(contractAddr))
}

// IteratePrivileged iterates through the list of privileged contacts by type and position ASC
func (k Keeper) IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), privilegedContractsSecondaryIndexPrefix)
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		if cb(iter.Key()) {
			return
		}
	}
}

// appendToPrivilegedContractCallbacks registers given contract for a callback type
func (k Keeper) appendToPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	// find last position value for callback type
	var pos uint8
	it := prefix.NewStore(store, getContractCallbacksSecondaryIndexPrefix(callbackType)).ReverseIterator(nil, nil)
	if it.Valid() {
		key := it.Key()
		pos = key[0]
	}
	newPos := pos + 1
	if newPos <= pos {
		panic("Overflow in in callback positions")
	}
	store.Set(contractCallbacksSecondaryIndexKey(callbackType, newPos, []byte{}), contractAddr)

	k.Logger(ctx).Info("Add callback", "contractAddr", contractAddr.String(), "callback_type", callbackType.String())
	event := sdk.NewEvent(
		types.EventTypeRegisterCallback,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
		sdk.NewAttribute(types.AttributeKeyCallbackType, callbackType.String()),
	)
	ctx.EventManager().EmitEvent(event)
}

// removePrivilegedContractCallbacks unregisters the given contract for a callback type
func (k Keeper) removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	start := append([]byte{0}, contractAddr...)
	end := append([]byte{math.MaxUint8}, contractAddr...)
	prefixStore := prefix.NewStore(store, getContractCallbacksSecondaryIndexPrefix(callbackType))

	var found bool
	for it := prefixStore.Iterator(start, end); it.Valid(); it.Next() {
		itKey := it.Key()
		if !sdk.AccAddress(it.Value()).Equals(contractAddr) {
			continue
		}
		prefixStore.Delete(itKey)
		found = true
	}
	if !found {
		return
	}
	k.Logger(ctx).Info("Remove callback", "contractAddr", contractAddr.String(), "callback_type", callbackType.String())
	event := sdk.NewEvent(
		types.EventTypeRegisterCallback,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
		sdk.NewAttribute(types.AttributeKeyCallbackType, callbackType.String()),
	)
	ctx.EventManager().EmitEvent(event)
}

// ExistsPrivilegedContractCallbacks returns if any contract is registered for the given type
func (k Keeper) ExistsPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType) bool {
	store := ctx.KVStore(k.storeKey)

	start := []byte{0}
	end := []byte{math.MaxUint8}
	prefixStore := prefix.NewStore(store, getContractCallbacksSecondaryIndexPrefix(callbackType))

	it := prefixStore.Iterator(start, end)
	return it.Valid()
}

// IterateContractCallbacksByType iterates through all contracts for the given type by position and address ASC
func (k Keeper) IterateContractCallbacksByType(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), getContractCallbacksSecondaryIndexPrefix(callbackType))
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		if cb(parseContractPosition(iter.Key()), iter.Value()) {
			return
		}
	}
}

// iterateContractCallbacksByContract iterates through all registered callbacks for the given contract. Ordered by type and position asc
func (k Keeper) iterateContractCallbacksByContract(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(t types.PrivilegedCallbackType, pos uint8) bool) {
	store := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, contractCallbacksSecondaryIndexPrefix)
	for it := prefixStore.Iterator(nil, nil); it.Valid(); it.Next() {
		t, pos := splitUnprefixedContractCallbacksSecondaryIndexKey(it.Key())
		addr := it.Value()
		if sdk.AccAddress(addr).Equals(contractAddress) { // index is not optimized for this. so we find all and have to check
			if cb(t, pos) {
				return
			}
		}
	}
}

func privilegedContractsSecondaryIndexKey(contractAddr sdk.AccAddress) []byte {
	return append(privilegedContractsSecondaryIndexPrefix, contractAddr...)
}

// contractCallbacksSecondaryIndexKey returns the key for privileged contract callbacks
// `<prefix><callbackType><position>
func contractCallbacksSecondaryIndexKey(callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) []byte {
	prefix := getContractCallbacksSecondaryIndexPrefix(callbackType)
	prefixLen := len(prefix)
	const posLen = 1 // 1 byte for position
	r := make([]byte, prefixLen+posLen)
	copy(r[0:], prefix)
	copy(r[prefixLen:], []byte{pos})
	return r
}

// getContractCallbacksSecondaryIndexPrefix return `<prefix><callbackType>`
func getContractCallbacksSecondaryIndexPrefix(callbackType types.PrivilegedCallbackType) []byte {
	return append(contractCallbacksSecondaryIndexPrefix, byte(callbackType))
}

// splits source of type `<callbackType><position>`
func splitUnprefixedContractCallbacksSecondaryIndexKey(key []byte) (types.PrivilegedCallbackType, uint8) {
	if len(key) != 1+1 {
		panic(fmt.Sprintf("unexpected key lenght %d", len(key)))
	}
	return types.PrivilegedCallbackType(key[0]), parseContractPosition(key[1:])
}

// splits source of type `<position>`
func parseContractPosition(key []byte) uint8 {
	if len(key) != 1 {
		panic(fmt.Sprintf("unexpected key lenght %d", len(key)))
	}
	return key[0]
}
