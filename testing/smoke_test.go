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
		println("Work dir: ", workDir)
	}
	defaultWaitTime = *waitTime
	sut = NewSystemUnderTest(verbose)
	if *rebuild {
		sut.BuildNewBinary()
	}
	// setup single node chain and keyring
	sut.SetupChain()

	// run tests
	exitCode := m.Run()

	// postprocess
	sut.StopChain()
	if verbose || exitCode != 0 {
		sut.PrintBuffer()
	}

	os.Exit(exitCode)
}

func TestSmokeTest(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := NewTgradeCli(t, sut, verbose)
	t.Log("List keys")
	t.Log("keys", cli.Keys("keys", "list"))

	t.Log("Upload wasm code")
	txResult := cli.CustomCommand("tx", "wasm", "store", "contrib/local/hackatom.wasm.gzip", "--from=node0", "--gas=1500000")
	RequireTxSuccess(t, txResult)

	t.Log("Waiting for block")
	sut.AwaitNextBlock(t)

	t.Log("Query wasm code list")
	qResult := cli.CustomQuery("q", "wasm", "list-code")
	codes := gjson.Get(qResult, "code_infos.#.code_id").Array()
	require.Len(t, codes, 1, qResult)
	require.Equal(t, int64(1), codes[0].Int())

	t.Log("got query result", qResult)
}

func TestPrivilegedInGenesis(t *testing.T) {
	sut.ResetChain(t)
	anyAddress := "tgrade12qey0qvmkvdu5yl3x329lhrvqfgzs5vne225q7"
	commands := [][]string{
		{
			"wasm-genesis-message",
			"store",
			"testing/contracts/cw4_group.wasm",
			"--instantiate-everybody=true",
			"--builder=foo/bar:latest",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"store",
			"testing/contracts/tgrade_valset.wasm",
			"--instantiate-everybody=true",
			"--builder=foo/bar:latest",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"1",
			`{"members":[{"addr":"tgrade189s8e528jm7scum9scw2g5z8yg7csdx39fu0gm", "weight": 1}]}`,
			"--label=testing",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"2",
			`{"membership":"tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy", "min_weight": 1, "max_validators":1, "epoch_length":1, "initial_keys":[{"operator":"tgrade189s8e528jm7scum9scw2g5z8yg7csdx39fu0gm","validator_pubkey":"N30jveAajHGhSOJ+jfpz5KYQWjmXRlQN0Y0MCZfCnKc="}]}`,
			"--label=testing",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-flags",
			"set-privileged",
			"tgrade10pyejy66429refv3g35g2t7am0was7yanjs539",
		},
	}
	// contract addresses are deterministic. You can get a list of all contracts in genesis via
	// `tgrade wasm-genesis-message list-contracts --home ./testnet/node0/tgrade`
	sut.ModifyGenesis(t, commands...)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// and then should be in list of privileged contracts
	qResult := cli.CustomQuery("q", "wasm", "privileged-contracts")
	contracts := gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 1, qResult)
	require.Equal(t, "tgrade10pyejy66429refv3g35g2t7am0was7yanjs539", contracts[0].String())
	t.Log("got query result", qResult)

	// and registered for validator update
	qResult = cli.CustomQuery("q", "wasm", "callback-contracts", "validator_set_update")
	contracts = gjson.Get(qResult, "contracts").Array()

	require.Len(t, contracts, 1, qResult)
	require.Equal(t, "tgrade10pyejy66429refv3g35g2t7am0was7yanjs539", contracts[0].String())
	t.Log("got query result", qResult)

	// and smart query internals
	qResult = cli.CustomQuery("q", "wasm", "contract-state", "smart", contracts[0].String(), `{"list_active_validators":{}}`)

	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, 1, qResult)
	t.Log("got query result", qResult)
}
