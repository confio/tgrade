package types

import wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

const (
	// DefaultParamspace for params keeper
	DefaultParamspace = ModuleName
)

func DefaultParams() wasmtypes.Params {
	return wasmtypes.DefaultParams()
}
