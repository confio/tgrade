package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math"
)

//AddRegisteredPrivilege add privilege type to list
func (d *TgradeContractDetails) AddRegisteredPrivilege(t PrivilegeType, pos uint8) {
	d.RegisteredPrivileges = append(d.RegisteredPrivileges, &RegisteredPrivilege{
		PrivilegeType: t.String(),
		Position:      uint32(pos),
	})
}

// RemoveRegisteredPrivilege remove privilege type from list
func (d *TgradeContractDetails) RemoveRegisteredPrivilege(t PrivilegeType, pos uint8) {
	src := &RegisteredPrivilege{
		PrivilegeType: t.String(),
		Position:      uint32(pos),
	}
	for i, v := range d.RegisteredPrivileges {
		if src.Equal(v) {
			d.RegisteredPrivileges = append(d.RegisteredPrivileges[0:i], d.RegisteredPrivileges[i+1:]...)
		}
	}
}

// HasRegisteredPrivilege returs true when given type was registered by this contract
func (d *TgradeContractDetails) HasRegisteredPrivilege(c PrivilegeType) bool {
	for _, v := range d.RegisteredPrivileges {
		if v.PrivilegeType == c.String() {
			return true
		}
	}
	return false
}

func (d TgradeContractDetails) IterateRegisteredPrivileges(cb func(c PrivilegeType, pos uint8) bool) {
	for _, v := range d.RegisteredPrivileges {
		if cb(*PrivilegeTypeFrom(v.PrivilegeType), uint8(v.Position)) {
			return
		}
	}
}

// ValidateBasic syntax checks
func (d TgradeContractDetails) ValidateBasic() error {
	unique := make(map[PrivilegeType]struct{})
	for i, c := range d.RegisteredPrivileges {
		if err := c.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "privilege %d", i)
		}
		privilegeType := *PrivilegeTypeFrom(c.PrivilegeType)
		if _, exists := unique[privilegeType]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "privilege %q", privilegeType.String())
		}
		unique[privilegeType] = struct{}{}
	}
	return nil
}

// ValidateBasic syntax checks
func (r RegisteredPrivilege) ValidateBasic() error {
	if r.Position > math.MaxUint8 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "position exceeds max")
	}
	if r.Position == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrEmpty, "position")
	}
	tp := PrivilegeTypeFrom(r.PrivilegeType)
	if tp == nil {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "privilege type")
	}
	return nil
}
