package types

import wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

type TWasmConfig struct {
	WasmConfig wasmtypes.WasmConfig
}

func DefaultTWasmConfig() TWasmConfig {
	return TWasmConfig{WasmConfig: wasmtypes.DefaultWasmConfig()}
}
