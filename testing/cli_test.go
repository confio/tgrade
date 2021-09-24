//go:build system_test
// +build system_test

package testing

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestUnsafeResetAll(t *testing.T) {
	// scenario:
	// 	given a non empty wasm dir exists in the node home
	//  when `unsafe-reset-all` is executed
	// 	then the dir and all files in it are removed

	wasmDir := filepath.Join(workDir, sut.nodePath(0), "wasm")
	require.NoError(t, os.MkdirAll(wasmDir, os.ModePerm))

	_, err := ioutil.TempFile(wasmDir, "testing")
	require.NoError(t, err)

	// when
	sut.ForEachNodeExecAndWait(t, []string{"unsafe-reset-all"})

	// then
	sut.withEachNodeHome(func(i int, home string) {
		if _, err := os.Stat(wasmDir); !os.IsNotExist(err) {
			t.Fatal("expected wasm dir to be removed")
		}
	})
}
