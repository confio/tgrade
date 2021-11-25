//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSmokeTest(t *testing.T) {
	// Scenario:
	// upload code
	// instantiate contract
	// watch for an event
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := NewTgradeCli(t, sut, verbose)
	t.Log("List keys")
	t.Log("keys", cli.Keys("keys", "list"))

	t.Log("Upload wasm code")
	txResult := cli.CustomCommand("tx", "wasm", "store", "testing/contract/hackatom.wasm.gzip", "--from=node0", "--gas=1500000")
	RequireTxSuccess(t, txResult)

	t.Log("Waiting for block")
	sut.AwaitNextBlock(t)

	t.Log("Query wasm code list")
	qResult := cli.CustomQuery("q", "wasm", "list-code")
	codes := gjson.Get(qResult, "code_infos.#.code_id").Array()
	const poeContractCount = 9
	const poeCodeCount = 8
	require.Len(t, codes, poeCodeCount+1, qResult)
	require.Equal(t, int64(poeCodeCount+1), codes[poeCodeCount].Int(), "sequential ids")
	codeID := strconv.Itoa(poeCodeCount + 1)

	t.Log("got query result", qResult)

	l := sut.NewEventListener(t)
	c, done := CaptureAllEventsConsumer(t)
	contractAddr := ContractBech32Address(poeCodeCount+1, poeContractCount+1)
	query := fmt.Sprintf(`tm.event='Tx' AND wasm._contract_address='%s'`, contractAddr)
	t.Logf("Subscribe to events: %s", query)

	cleanupFn := l.Subscribe(query, c)
	t.Cleanup(cleanupFn)

	t.Log("Instantiate wasm code")
	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	txResult = cli.CustomCommand("tx", "wasm", "instantiate", codeID, initMsg, "--label=testing", "--from=node0", "--gas=1500000")
	RequireTxSuccess(t, txResult)
	assert.Len(t, done(), 1)
	assert.Contains(t, txResult, contractAddr)
}
