//go:build system_test

package testing

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	sut.MarkDirty()

	cli := NewTgradeCli(t, sut, verbose)
	t.Log("List keys")
	t.Log("keys", cli.Keys("keys", "list"))

	t.Log("Upload wasm code")
	txRsp := cli.CustomCommand("tx", "wasm", "store", "testing/contract/hackatom.wasm.gzip", "--from=node0", "--gas=1600000")
	RequireTxSuccess(t, txRsp)
	codeChecksum := gjson.Get(txRsp, "logs.#.events.#.attributes.#(key=code_checksum).value").Array()[0].Array()[0].String()
	require.NotEmpty(t, codeChecksum)
	codeChecksumBz, err := hex.DecodeString(codeChecksum)
	require.NoError(t, err)
	t.Log("Waiting for block")
	sut.AwaitNextBlock(t)

	t.Log("Query wasm code list")
	qResult := cli.CustomQuery("q", "wasm", "list-code")
	codes := gjson.Get(qResult, "code_infos.#.code_id").Array()
	t.Log("got query result", qResult)

	const poeContractCount = 11
	const poeCodeCount = 9
	require.Len(t, codes, poeCodeCount+1, qResult)
	require.Equal(t, int64(poeCodeCount+1), codes[poeCodeCount].Int(), "sequential ids")
	codeID := poeCodeCount + 1

	l := sut.NewEventListener(t)
	c, done := CaptureAllEventsConsumer(t)
	expContractAddr := ContractBech32Address(poeCodeCount+1, poeContractCount+1)
	query := fmt.Sprintf(`tm.event='Tx' AND wasm._contract_address='%s'`, expContractAddr)
	t.Logf("Subscribe to events: %s", query)
	cleanupFn := l.Subscribe(query, c)
	t.Cleanup(cleanupFn)

	t.Log("Instantiate wasm code - classic")
	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	newContractAddr := cli.InstantiateWasm(codeID, initMsg)
	assert.Equal(t, expContractAddr, newContractAddr)

	t.Log("Instantiate wasm code - predictable address")
	args := []string{"--label=testing", "--from=" + defaultSrcAddr, "--no-admin", "--fix-msg"}
	salt := hex.EncodeToString([]byte("my-salt"))
	rsp := cli.run(cli.withTXFlags(append([]string{"tx", "wasm", "instantiate2", strconv.Itoa(codeID), initMsg, salt}, args...)...))
	RequireTxSuccess(t, rsp)

	gotContractAddr := gjson.Get(rsp, "logs.#.events.#.attributes.#(key=_contract_address).value").Array()[0].Array()[0].String()
	require.NotEmpty(t, gotContractAddr)
	expContractAddr = wasmkeeper.BuildContractAddressPredictable(codeChecksumBz, sdk.MustAccAddressFromBech32(cli.GetDefaultKeyAddr()), []byte("my-salt"), []byte(initMsg)).String()
	assert.Equal(t, expContractAddr, gotContractAddr)

	assert.Len(t, done(), 1)
}
