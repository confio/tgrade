// +build system_test

package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Scenario: add WASM code as part of genesis and pin it in VM cache forever
//           for faster execution.
func TestGenesisCodePin(t *testing.T) {
	sut.ResetChain(t)
	// WASM code 1-5 is present.
	sut.ModifyGenesisCLI(t,
		[]string{"wasm-genesis-flags", "set-pinned", "1"},
		[]string{"wasm-genesis-flags", "set-pinned", "3"},
	)

	// At the time of writing this test, there is no public interface to
	// check if code is cached or not. Instead, we are checking the genesis
	// file only.
	// No mutation, we are only interested in checking the content.
	sut.ModifyGenesisJson(t, func(raw []byte) []byte {
		codeIDs := gjson.GetBytes(raw, "app_state.wasm.pinned_code_ids").Array()
		require.Len(t, codeIDs, 2)
		require.Equal(t, codeIDs[0].Int(), int64(1))
		require.Equal(t, codeIDs[1].Int(), int64(3))
		return raw
	})
}
