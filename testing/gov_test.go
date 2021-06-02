// +build system_test

package testing

import (
	"fmt"
	"github.com/confio/tgrade/x/twasm"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
	sut.ModifyGenesisJson(t, func(genesis []byte) []byte {
		t.Helper()
		// move all contracts addresses by +1
		state, err := sjson.SetBytes(genesis, "app_state.genutil.engagement_contract_addr", []byte(twasm.ContractAddress(2, 2).String()))
		require.NoError(t, err)
		state, err = sjson.SetBytes(state, "app_state.genutil.staking_contract_addr", []byte(twasm.ContractAddress(3, 3).String()))
		require.NoError(t, err)
		state, err = sjson.SetBytes(state, "app_state.genutil.mixer_contract_addr", []byte(twasm.ContractAddress(4, 4).String()))
		require.NoError(t, err)
		state, err = sjson.SetBytes(state, "app_state.genutil.valset_contract_addr", []byte(twasm.ContractAddress(5, 5).String()))
		require.NoError(t, err)
		return state
	})
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
