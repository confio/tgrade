// +build system_test

package testing

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"testing"
)

var sut *SystemUnderTest
var verbose bool

func TestMain(m *testing.M) {
	rebuild := flag.Bool("rebuild", false, "rebuild artifacts")
	waitTime := flag.Duration("wait-time", defaultWaitTime, "time to wait for chain events")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workDir = filepath.Join(dir, "../")
	if verbose {
		println("+++> Work dir: ", workDir)
	}
	defaultWaitTime = *waitTime
	sut = NewSystemUnderTest(verbose)
	if *rebuild {
		sut.BuildNewContainer()
	}
	// setup single node chain and keyring
	sut.SetupChain()

	// run tests
	var exitCode int
	func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Recovering panic: %#+v\n", err)
				exitCode = 1
			}
		}()
		exitCode = m.Run()
	}()

	// postprocess
	sut.StopChain()
	if verbose || exitCode != 0 {
		sut.PrintBuffer()
	}
	os.Exit(exitCode)
}

func TestSmokeTest(t *testing.T) {
	sut.Restart()
	sut.StartChain()
	t.Cleanup(sut.StopChain)

	cli := NewTgradeCli(t, sut, verbose)
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
