// +build system_test

package testing

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"strings"
	"testing"
)

// Scenario: add reflect contract to genesis and set privileged
// 			 trigger gov proposal to unset privileges
//			 then verify that callback permission was removed
func TestGovProposal(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	myKey := strings.Trim(cli.Keys("keys", "show", "-a", "node0"), "\n ")
	t.Logf("key: %q", myKey)
	commands := [][]string{
		{
			"wasm-genesis-message",
			"store",
			"x/poe/contract/tgrade_gov_reflect.wasm",
			"--instantiate-everybody=true",
			"--builder=foo/bar:latest",
			fmt.Sprintf("--run-as=%s", myKey),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"1",
			`{}`,
			"--label=testing",
			fmt.Sprintf("--run-as=%s", myKey),
		},
		{
			"wasm-genesis-flags",
			"set-privileged",
			"tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy",
		},
	}
	sut.ModifyGenesisCLI(t, commands...)
	sut.StartChain(t)

	qResult := cli.CustomQuery("q", "wasm", "callback-contracts", "gov_proposal_executor")
	contracts := gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 1, qResult)
	require.Equal(t, "tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy", contracts[0].String())
	t.Log("got query result", qResult)

	// when
	t.Log("Send a proposal to be returned")
	excecMsg := `{"proposal":{"title":"foo", "description":"bar", "proposal":{"demote_privileged_contract":{"contract":"tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy"}}}}`
	txResult := cli.CustomCommand("tx", "wasm", "execute", "tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy", excecMsg, fmt.Sprintf("--from=%s", myKey), "--gas=1500000")
	RequireTxSuccess(t, txResult)

	// then should not be privileged anymore
	qResult = cli.CustomQuery("q", "wasm", "callback-contracts", "gov_proposal_executor")
	contracts = gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 0, qResult)
}
