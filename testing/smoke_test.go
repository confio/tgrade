// +build system_test

package testing

import (
	"flag"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"os"
	"testing"
)

var sut *SystemUnderTest

func TestMain(m *testing.M) {
	rebuild := flag.Bool("rebuild", false, "rebuild artifacts")
	waitTime := flag.Duration("wait-time", defaultWaitTime, "time to wait for chain events")
	flag.Parse()

	defaultWaitTime = *waitTime
	sut = NewSystemUnderTest()
	if *rebuild {
		// make install docker-build
		sut.CompileBinaries()
	}
	// setup single node chain and keyring
	sut.SetupChain()
	os.Exit(m.Run())
	sut.StopChain()
}

func TestSmokeTest(t *testing.T) {
	sut.Restart()
	sut.StartChain()
	t.Cleanup(sut.StopChain)

	cli := NewTgradeCli(t, sut)
	t.Log("List keys")
	t.Log("keys", cli.Keys("keys", "list"))

	t.Log("Upload wasm code")
	txResult := cli.CustomCommand("tx", "wasm", "store", "contrib/local/hackatom.wasm.gzip", "--from=node0", "--gas=1500000")
	RequireTxSuccess(t, txResult)

	t.Log("Waiting for block")
	sut.AwaitNextBlock()

	t.Log("Query wasm code list")
	qResult := cli.CustomQuery("q", "wasm", "list-code")
	codes := gjson.Get(qResult, "code_infos.#.id").Array() // see query syntax https://github.com/tidwall/gjson#path-syntax
	require.Len(t, codes, 1, qResult)
	require.Equal(t, int64(1), codes[0].Int())

	t.Log("got query result", qResult)
}
