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

	// call contract and let it register for privileges
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
// - remove all privileges for the contract
func (k Keeper) UnsetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	// call contract to release privileges
	msg := contract.TgradeSudoMsg{PrivilegeChange: &contract.PrivilegeChangeMsg{Demoted: &struct{}{}}}
	msgBz, err := json.Marshal(&msg)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	if _, err = k.Sudo(ctx, contractAddr, msgBz); err != nil {
		return sdkerrors.Wrap(err, "sudo")
	}

	// load after sudo so that unregister messages were handled
	contractInfo := k.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contractAddr")
	}

	// remove from cache
	if err := k.contractKeeper.UnpinCode(ctx, contractInfo.CodeID); err != nil {
		return sdkerrors.Wrapf(err, "unpin")
	}

	// remove privileged flag
	k.clearPrivilegedFlag(ctx, contractAddr)

	// remove remaining privileges
	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return err
	}
	details.IterateRegisteredPrivileges(func(privilegeType types.PrivilegeType, pos uint8) bool {
		k.removePrivilegeRegistration(ctx, privilegeType, pos, contractAddr)
		details.RemoveRegisteredPrivilege(privilegeType, pos)
		return false
	})
	if err := k.setContractDetails(ctx, contractAddr, &details); err != nil {
		return sdkerrors.Wrap(err, "store contract info extension")
	}

	k.Logger(ctx).Info("Unset privileged", "contractAddr", contractAddr.String())
	event := sdk.NewEvent(
		types.EventTypeUnsetPrivileged,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
	)
	ctx.EventManager().EmitEvent(event)
	return nil
}

// importPrivileged import from genesis
func (k Keeper) importPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress, codeID uint64, details types.TgradeContractDetails) error {
	// add to cache
	if err := k.contractKeeper.PinCode(ctx, codeID); err != nil {
		return sdkerrors.Wrapf(err, "pin")
	}

	// set privileged flag
	k.setPrivilegedFlag(ctx, contractAddr)

	store := ctx.KVStore(k.storeKey)
	for _, c := range details.RegisteredPrivileges {
		var (
			privilegeType = types.PrivilegeTypeFrom(c.PrivilegeType)
			pos           = uint8(c.Position)
		)
		if privilegeType == nil {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "unknown privilege type: %q", c.PrivilegeType)
		}
		key := contractPrivilegesSecondaryIndexKey(*privilegeType, pos)
		if store.Has(key) {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis,
				"privilege exists already: %s for contract %s", privilegeType.String(), contractAddr.String())
		}
		k.storeContractPrivilegeRegistration(ctx, *privilegeType, pos, contractAddr)
	}
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

// appendToPrivilegedContracts registers given contract for a privilege type.
func (k Keeper) appendToPrivilegedContracts(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddr sdk.AccAddress) (uint8, error) {
	store := ctx.KVStore(k.storeKey)

	// find last position value for privilege type
	var pos uint8
	it := prefix.NewStore(store, getContractPrivilegesSecondaryIndexPrefix(privilegeType)).ReverseIterator(nil, nil)
	if it.Valid() {
		key := it.Key()
		pos = key[0]
		if privilegeType.IsSingleton() {
			return 0, wasmtypes.ErrDuplicate
		}
	}
	newPos := pos + 1
	if newPos <= pos {
		panic("Overflow in privilege positions")
	}
	k.storeContractPrivilegeRegistration(ctx, privilegeType, newPos, contractAddr)

	k.Logger(ctx).Info("Add privilege", "contractAddr", contractAddr.String(), "type", privilegeType.String())
	event := sdk.NewEvent(
		types.EventTypeRegisterPrivilege,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
		sdk.NewAttribute(types.AttributeKeyCallbackType, privilegeType.String()),
	)
	ctx.EventManager().EmitEvent(event)
	return newPos, nil
}

// storeContractPrivilegeRegistration persists the privilege registration the contract
func (k Keeper) storeContractPrivilegeRegistration(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(contractPrivilegesSecondaryIndexKey(privilegeType, pos), contractAddr)
}

// removePrivilegeRegistration unregisters the given contract for a privilege type
func (k Keeper) removePrivilegeRegistration(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := contractPrivilegesSecondaryIndexKey(privilegeType, pos)
	if !store.Has(key) {
		return false
	}
	store.Delete(key)
	k.Logger(ctx).Info("Remove privilege", "contractAddr", contractAddr.String(), "type", privilegeType.String())
	event := sdk.NewEvent(
		types.EventTypeReleasePrivilege,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(wasmtypes.AttributeKeyContract, contractAddr.String()),
		sdk.NewAttribute(types.AttributeKeyCallbackType, privilegeType.String()),
	)
	ctx.EventManager().EmitEvent(event)
	return true
}

// getPrivilegedContract returns the key stored at the given type and position. Result can be nil when none exists
func (k Keeper) getPrivilegedContract(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	key := contractPrivilegesSecondaryIndexKey(privilegeType, pos)
	return store.Get(key)
}

// ExistsAnyPrivilegedContract returns if any contract is registered for the given type
func (k Keeper) ExistsAnyPrivilegedContract(ctx sdk.Context, privilegeType types.PrivilegeType) bool {
	store := ctx.KVStore(k.storeKey)

	start := []byte{0}
	end := []byte{math.MaxUint8}
	prefixStore := prefix.NewStore(store, getContractPrivilegesSecondaryIndexPrefix(privilegeType))

	it := prefixStore.Iterator(start, end)
	return it.Valid()
}

// IteratePrivilegedContractsByType iterates through all contracts for the given type by position and address ASC
func (k Keeper) IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), getContractPrivilegesSecondaryIndexPrefix(privilegeType))
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		if cb(parseContractPosition(iter.Key()), iter.Value()) {
			return
		}
	}
}

// HasPrivilegedContract returns if the contract has the given privilege type registered.
// Returns error for unknown contract addresses.
func (k Keeper) HasPrivilegedContract(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType types.PrivilegeType) (bool, error) {
	d, err := k.getContractDetails(ctx, contractAddr)
	if err != nil {
		return false, err
	}
	return d.HasRegisteredPrivilege(privilegeType), nil
}

func privilegedContractsSecondaryIndexKey(contractAddr sdk.AccAddress) []byte {
	return append(privilegedContractsSecondaryIndexPrefix, contractAddr...)
}

// contractPrivilegesSecondaryIndexKey returns the key for contract privileges
// `<prefix><privilegeType><position>
func contractPrivilegesSecondaryIndexKey(privilegeType types.PrivilegeType, pos uint8) []byte {
	prefix := getContractPrivilegesSecondaryIndexPrefix(privilegeType)
	prefixLen := len(prefix)
	const posLen = 1 // 1 byte for position
	r := make([]byte, prefixLen+posLen)
	copy(r[0:], prefix)
	copy(r[prefixLen:], []byte{pos})
	return r
}

// getContractPrivilegesSecondaryIndexPrefix return `<prefix><privilegeType>`
func getContractPrivilegesSecondaryIndexPrefix(privilegeType types.PrivilegeType) []byte {
	return append(contractCallbacksSecondaryIndexPrefix, byte(privilegeType))
}

// splits source of type `<position>`
func parseContractPosition(key []byte) uint8 {
	if len(key) != 1 {
		panic(fmt.Sprintf("unexpected key lenght %d", len(key)))
	}
	return key[0]
}
