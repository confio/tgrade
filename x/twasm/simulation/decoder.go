package simulation

import (
	"bytes"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
)

type AuthUnmarshaler interface {
	GetCodec() codec.Codec
}

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding auth type.
func NewDecodeStore(ak *twasmkeeper.Keeper) func(kvA kv.Pair, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		// case bytes.Equal(kvA.Key[:1], types.PrivilegedContractsSecondaryIndexPrefix):
		// case bytes.Equal(kvA.Key[:1], types.ContractCallbacksSecondaryIndexPrefix):
		// case bytes.Equal(kvA.Key[:1], wasmtypes.CodeKeyPrefix):
		// case bytes.Equal(kvA.Key[:1], wasmtypes.ContractKeyPrefix):
		// case bytes.Equal(kvA.Key[:1], wasmtypes.ContractStorePrefix):
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
