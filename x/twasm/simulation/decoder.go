package simulation

// DONTCOVER

import (
	"bytes"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding twasm type.
// Not fully implemented but falls back into default output
func NewDecodeStore() func(kvA kv.Pair, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], wasmtypes.ContractStorePrefix):
			var addrA, pathA string
			if len(kvA.Key) > wasmtypes.ContractAddrLen+1 {
				addrA = sdk.AccAddress(kvA.Key[1 : wasmtypes.ContractAddrLen+1]).String()
				pathA = string(kvA.Key[wasmtypes.ContractAddrLen+1:])
			} else {
				addrA = fmt.Sprintf("%q", kvA.Key[1:])
			}
			var addrB, pathB string
			if len(kvB.Key) > wasmtypes.ContractAddrLen+1 {
				addrB = sdk.AccAddress(kvB.Key[1 : wasmtypes.ContractAddrLen+1]).String()
				pathB = string(kvB.Key[wasmtypes.ContractAddrLen+1:])
			} else {
				addrB = fmt.Sprintf("%q", kvB.Key[1:])
			}
			return fmt.Sprintf("Contract storage A %s/%s=>%s\nContract storage B %s/%s=>%s", addrA, pathA, string(kvA.Value),
				addrB, pathB, string(kvB.Value))
		default:
			return fmt.Sprintf("store A %q => %q\nstore B %q => %q\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)
		}
	}
}
