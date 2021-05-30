// +build system_test

package testing

import (
	"encoding/base64"
	"fmt"
	testingcontracts "github.com/confio/tgrade/testing/contracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tidwall/gjson"
	"path/filepath"
	"testing"
)

// TestProofOfAuthoritySetup instantiates the tgrade valset contract and setup cluster to run n validators.
// Then the validator powers are modified via an tg4 contract member update.
func TestProofOfAuthoritySetup(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// contract addresses are deterministic. You can get a list of all contracts in genesis via
	// `tgrade wasm-genesis-message list-contracts --home ./testnet/node0/tgrade`
	var (
		tg4ContractAddr    = ContractBech32Address(1, 1)
		valsetContractAddr = ContractBech32Address(2, 2)
		anyAddress         = "tgrade12qey0qvmkvdu5yl3x329lhrvqfgzs5vne225q7"
		tg4AdminAddr       = cli.AddKey("tg4admin")
	)
	// prepare contract init messages with chain validator data
	tg4initMsg := testingcontracts.TG4GroupInitMsg{
		Admin:   tg4AdminAddr,
		Members: make([]testingcontracts.TG4Member, sut.nodesCount),
	}
	valsetInitMsg := testingcontracts.ValsetInitMsg{
		Membership:    tg4ContractAddr,
		MinWeight:     1,
		MaxValidators: 100,
		EpochLength:   1,
		InitialKeys:   make([]testingcontracts.ValsetInitKey, sut.nodesCount),
	}

	sut.withEachNodeHome(func(i int, home string) {
		k := readPubkey(t, filepath.Join(workDir, home, "config", "priv_validator_key.json"))
		pubKey := base64.StdEncoding.EncodeToString(k.Bytes())
		addr := randomBech32Addr()
		tg4initMsg.Members[i] = testingcontracts.TG4Member{
			Addr:   addr,
			Weight: sut.nodesCount - i, // unique weight
		}
		valsetInitMsg.InitialKeys[i] = testingcontracts.NewValsetInitKey(addr, pubKey)
	})

	commands := [][]string{
		{
			"wasm-genesis-message",
			"store",
			"testing/contracts/tg4_group.wasm",
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
			tg4initMsg.Json(t),
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
	sut.ModifyGenesisCLI(t, commands...)
	sut.StartChain(t)

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
	sortedMember := testingcontracts.SortByWeight(tg4initMsg.Members)
	for i := 0; i < sut.nodesCount; i++ {
		// ordered by power desc
		exp := sortedMember[i]
		assert.Equal(t, int64(exp.Weight), v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}

	// And when weight updated
	cli.FundAddress(t, tg4AdminAddr, "1000utgd")
	tg4UpdateMsg := testingcontracts.TG4UpdateMembersMsg{
		Add: make([]testingcontracts.TG4Member, len(tg4initMsg.Members)),
	}
	// update weights for all members
	for i, v := range tg4initMsg.Members {
		tg4UpdateMsg.Add[i] = testingcontracts.TG4Member{
			Addr:   v.Addr,
			Weight: 10 + i,
		}
	}
	eResult := cli.Execute(tg4ContractAddr, tg4UpdateMsg.Json(t), "tg4admin")
	RequireTxSuccess(t, eResult)
	t.Log("got execution result", eResult)
	// wait for msg execution
	sut.AwaitNextBlock(t, sut.blockTime*5)
	// wait for update manifests in valset (epoch has completed)
	sut.AwaitNextBlock(t)

	// then validator set is updated
	v = sut.RPCClient(t).Validators()
	require.Len(t, v, sut.nodesCount, "got %#v", v)
	sortedMember = testingcontracts.SortByWeight(tg4UpdateMsg.Add)
	for i := 0; i < sut.nodesCount; i++ {
		// ordered by power desc
		exp := sortedMember[i]
		assert.Equal(t, int64(exp.Weight), v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}
}

func readPubkey(t *testing.T, filePath string) crypto.PubKey {
	key, err := p2p.LoadOrGenNodeKey(filePath)
	require.NoError(t, err)
	return key.PubKey()
}
