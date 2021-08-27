// +build system_test

package testing

import (
	"testing"

	"github.com/tidwall/gjson"
)

// Scenario: add WASM code as part of genesis and pin it in VM cache forever
//           for faster execution.
func TestGenesisCodePin(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// WASM code 1-5 is present.
	sut.ModifyGenesisCLI(t,
		[]string{"wasm-genesis-flags", "set-pinned", "1"},
		[]string{"wasm-genesis-flags", "set-pinned", "3"},
	)
	sut.StartChain(t)

	// TODO - how to check if code is pinned?
	qResult := cli.CustomQuery("q", "wasm", "list-codes")
	codes := gjson.Get(qResult, "code_infos").Array()
}
