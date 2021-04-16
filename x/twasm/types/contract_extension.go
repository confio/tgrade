package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math"
)

func (d *TgradeContractDetails) AddRegisteredCallback(t PrivilegedCallbackType, pos uint8) {
	d.RegisteredCallbacks = append(d.RegisteredCallbacks, &RegisteredCallback{
		CallbackType: t.String(),
		Position:     uint32(pos),
	})
}

func (d *TgradeContractDetails) RemoveRegisteredCallback(t PrivilegedCallbackType, pos uint8) {
	src := &RegisteredCallback{
		CallbackType: t.String(),
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
		if v.CallbackType == c.String() {
			return true
		}
	}
	return false
}

func (d TgradeContractDetails) IterateRegisteredCallbacks(cb func(c PrivilegedCallbackType, pos uint8) bool) {
	for _, v := range d.RegisteredCallbacks {
		if cb(*PrivilegedCallbackTypeFrom(v.CallbackType), uint8(v.Position)) {
			return
		}
	}
}

// ValidateBasic syntax checks
func (d TgradeContractDetails) ValidateBasic() error {
	unique := make(map[PrivilegedCallbackType]struct{})
	for i, c := range d.RegisteredCallbacks {
		if err := c.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "callback %d", i)
		}
		callbackType := *PrivilegedCallbackTypeFrom(c.CallbackType)
		if _, exists := unique[callbackType]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "callback %s", callbackType.String())
		}
		unique[callbackType] = struct{}{}
	}
	return nil
}

// ValidateBasic syntax checks
func (r RegisteredCallback) ValidateBasic() error {
	if r.Position > math.MaxUint8 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "position exceeds max")
	}
	tp := PrivilegedCallbackTypeFrom(r.CallbackType)
	if tp == nil {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "callback type")
	}
	return nil
}
