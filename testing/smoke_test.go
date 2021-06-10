// +build system_test

package testing

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"strconv"
	"testing"
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
	require.Len(t, codes, 1, qResult)
	require.Equal(t, int64(1), codes[0].Int())
	codeID := strconv.Itoa(1)
	t.Log("got query result", qResult)

	l := sut.NewEventListener(t)
	c, done := CaptureAllEventsConsumer(t)
	query := fmt.Sprintf(`tm.event='Tx' AND wasm.contract_address='%s'`, ContractBech32Address(1, 1))
	t.Logf("Subscribe to events: %s", query)

	cleanupFn := l.Subscribe(query, c)
	t.Cleanup(cleanupFn)

	t.Log("Instantiate wasm code")
	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	txResult = cli.CustomCommand("tx", "wasm", "instantiate", codeID, initMsg, "--label=testing", "--from=node0", "--gas=1500000")
	RequireTxSuccess(t, txResult)
	assert.Len(t, done(), 1)
	assert.Contains(t, txResult, ContractBech32Address(1, 1))
}
