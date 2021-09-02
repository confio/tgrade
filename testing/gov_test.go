//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

// Scenario: add reflect contract to genesis and set privileged
// 			 trigger gov proposal to unset privileges
//			 then verify that callback permission was removed
func TestGovProposal(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	myKey := cli.GetKeyAddr("node0")
	require.NotEmpty(t, myKey)
	t.Logf("key: %q", myKey)
	myContractAddr := ContractBech32Address(1, 1)
	commands := [][]string{
		{
			"wasm-genesis-message",
			"store",
			"x/poe/contract/tgrade_gov_reflect.wasm",
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
			myContractAddr,
		},
	}
	sut.ModifyGenesisCLI(t, commands...)
	sut.StartChain(t)

	qResult := cli.CustomQuery("q", "wasm", "list-privileged-by-type", "gov_proposal_executor")
	contracts := gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 1, qResult)
	require.Equal(t, myContractAddr, contracts[0].String())
	t.Log("got query result", qResult)

	// when
	t.Log("Send a proposal to be returned")
	excecMsg := fmt.Sprintf(`{"proposal":{"title":"foo", "description":"bar", "proposal":{"demote_privileged_contract":{"contract":%q}}}}`, myContractAddr)
	txResult := cli.CustomCommand("tx", "wasm", "execute", myContractAddr, excecMsg, fmt.Sprintf("--from=%s", myKey), "--gas=1500000")
	RequireTxSuccess(t, txResult)

	// then should not be privileged anymore
	qResult = cli.CustomQuery("q", "wasm", "list-privileged-by-type", "gov_proposal_executor")
	contracts = gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 0, qResult)
}
