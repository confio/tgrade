package twasm

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContractAddress generator function. Same as in wasm
func ContractAddress(codeID, instanceID uint64) sdk.AccAddress {
	return wasmkeeper.BuildContractAddress(codeID, instanceID)
}
