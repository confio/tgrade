// +build system_test

package testing

import (
	"encoding/base64"
	"fmt"
	testingcontracts "github.com/confio/tgrade/testing/contracts"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tidwall/gjson"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	RunTestMain(m)
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

// TestPrivilegedInGenesis instantiates the tgrade valset contract and setup cluster to run n-1 validators.
func TestPrivilegedInGenesis(t *testing.T) {
	sut.ResetChain(t)
	// contract addresses are deterministic. You can get a list of all contracts in genesis via
	// `tgrade wasm-genesis-message list-contracts --home ./testnet/node0/tgrade`
	const (
		cw4ContractAddr    = "tgrade18vd8fpwxzck93qlwghaj6arh4p7c5n89hzs8hy"
		valsetContractAddr = "tgrade10pyejy66429refv3g35g2t7am0was7yanjs539"
		anyAddress         = "tgrade12qey0qvmkvdu5yl3x329lhrvqfgzs5vne225q7"
	)
	// prepare contract init messages with chain validator data
	cw4initMsg := testingcontracts.CW4InitMsg{Members: make([]testingcontracts.CW4Member, sut.nodesCount)}
	valsetInitMsg := testingcontracts.ValsetInitMsg{
		Membership:    cw4ContractAddr,
		MinWeight:     1,
		MaxValidators: 100,
		EpochLength:   1,
		InitialKeys:   make([]testingcontracts.ValsetInitKey, sut.nodesCount),
	}

	sut.withEachNodeHome(func(i int, home string) {
		k := readPubkey(t, filepath.Join(workDir, home, "config", "priv_validator_key.json"))
		pubKey := base64.StdEncoding.EncodeToString(k.Bytes())
		addr := randomBech32Addr()
		cw4initMsg.Members[i] = testingcontracts.CW4Member{
			Addr:   addr,
			Weight: sut.nodesCount - i, // unique weight
		}
		valsetInitMsg.InitialKeys[i] = testingcontracts.ValsetInitKey{
			Operator:        addr,
			ValidatorPubkey: pubKey,
		}
	})

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
			cw4initMsg.Json(t),
			"--label=testing",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"2",
			valsetInitMsg.Json(t),
			"--label=testing",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-flags",
			"set-privileged",
			valsetContractAddr,
		},
	}
	sut.ModifyGenesis(t, commands...)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// and then should be in list of privileged contracts
	qResult := cli.CustomQuery("q", "wasm", "privileged-contracts")
	contracts := gjson.Get(qResult, "contracts").Array()
	require.Len(t, contracts, 1, qResult)
	require.Equal(t, valsetContractAddr, contracts[0].String())
	t.Log("got query result", qResult)

	// and registered for validator update
	qResult = cli.CustomQuery("q", "wasm", "callback-contracts", "validator_set_update")
	contracts = gjson.Get(qResult, "contracts").Array()

	require.Len(t, contracts, 1, qResult)
	require.Equal(t, valsetContractAddr, contracts[0].String())
	t.Log("got query result", qResult)

	// and smart query internals
	qResult = cli.CustomQuery("q", "wasm", "contract-state", "smart", contracts[0].String(), `{"list_active_validators":{}}`)

	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, sut.nodesCount, qResult)
	t.Log("got query result", qResult)

	// and ensure validator count via rpc call
	sut.AwaitNextBlock(t)
	v := sut.RPCClient(t).Validators()
	require.Len(t, v, sut.nodesCount, "got %#v", v)
	for i := 0; i < sut.nodesCount; i++ {
		// ordered by power desc
		assert.Equal(t, int64(sut.nodesCount-i), v[i].VotingPower, "address: %s", v[i].Address.String())
	}
}

func randomBech32Addr() string {
	bech32Addr, _ := bech32.ConvertAndEncode("tgrade", rand.Bytes(sdk.AddrLen))
	return bech32Addr
}

func readPubkey(t *testing.T, filePath string) crypto.PubKey {
	key, err := p2p.LoadOrGenNodeKey(filePath)
	require.NoError(t, err)
	return key.PubKey()
}
