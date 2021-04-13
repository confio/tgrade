package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func ContractDetails(c wasmtypes.ContractInfo) (*TgradeContractDetails, error) {
	if c.Extension == nil {
		return &TgradeContractDetails{}, nil
	}
	e, ok := c.Extension.GetCachedValue().(*TgradeContractDetails)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "extension: %T", c.Extension)
	}
	return e, nil
}

func (d *TgradeContractDetails) AddRegisteredCallback(t PrivilegedCallbackType, pos uint8) {
	d.RegisteredCallbacks = append(d.RegisteredCallbacks, &RegisteredCallback{
		CallbackType: uint32(t),
		Position:     uint32(pos),
	})
}

func (d *TgradeContractDetails) RemoveRegisteredCallback(t PrivilegedCallbackType, pos uint8) {
	src := &RegisteredCallback{
		CallbackType: uint32(t),
		Position:     uint32(pos),
	}
	for i, v := range d.RegisteredCallbacks {
		if src.Equal(v) {
			d.RegisteredCallbacks = append(d.RegisteredCallbacks[0:i], d.RegisteredCallbacks[i+1:]...)
		}
	}
}

func (d *TgradeContractDetails) HasRegisteredContractCallback(c PrivilegedCallbackType) bool {
	for _, v := range d.RegisteredCallbacks {
		if PrivilegedCallbackType(v.CallbackType) == c {
			return true
		}
	}
	return false
}

func (d TgradeContractDetails) IterateRegisteredCallbacks(cb func(t PrivilegedCallbackType, pos uint8) bool) {
	for _, v := range d.RegisteredCallbacks {
		if cb(PrivilegedCallbackType(v.CallbackType), uint8(v.Position)) {
			return
		}
	}
}
