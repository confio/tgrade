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

type Paginator struct {
	// Any raw entry from the returned list. Must be deserialized to the correct type depending on paginated query
	StartAfter []byte `json:"start_after,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
}
