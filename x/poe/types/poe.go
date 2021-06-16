package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (t PoEContractType) ValidateBasic() error {
	if t == PoEContractTypeUndefined {
		return wasmtypes.ErrInvalid
	}
	if _, ok := PoEContractType_name[int32(t)]; !ok {
		return wasmtypes.ErrNotFound
	}
	return nil
}

func PoEContractTypeFrom(s string) PoEContractType {
	v, ok := PoEContractType_value[s]
	if !ok {
		return PoEContractTypeUndefined
	}
	return PoEContractType(v)
}
