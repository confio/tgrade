package types

import (
	"sort"

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

// PoEContractTypeFrom converts string representation to PoEContractType.
// Returns PoEContractTypeUndefined for unknown or misspelled values.
func PoEContractTypeFrom(s string) PoEContractType {
	v, ok := PoEContractType_value[s]
	if !ok {
		return PoEContractTypeUndefined
	}
	return PoEContractType(v)
}

// IteratePoEContractTypes for each defined poe contract type the given callback is called deterministic.
// When the callback returns true the loop is aborted early
func IteratePoEContractTypes(cb func(tp PoEContractType) bool) {
	names := make([]string, 0, len(PoEContractType_name)-1)
	for _, v := range PoEContractType_name {
		if v == PoEContractTypeUndefined.String() {
			continue
		}
		names = append(names, v)
	}
	sort.Strings(names)
	for _, v := range names {
		if cb(PoEContractTypeFrom(v)) {
			return
		}
	}
}
