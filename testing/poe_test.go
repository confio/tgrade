//go:build system_test
// +build system_test

package testing

import (
	"encoding/json"
	"fmt"
	testingcontracts "github.com/confio/tgrade/x/poe/contract"
	poetypes "github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"math"
	"path/filepath"
	"testing"
	"time"
)

func TestProofOfEngagementSetup(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// contract addresses are deterministic. You can get a list of all contracts in genesis via
	// `tgrade wasm-genesis-message list-contracts --home ./testnet/node0/tgrade`
	var (
		tg4AdminAddr = cli.GetKeyAddr("systemadmin")
	)
	engagementGroup := make([]testingcontracts.TG4Member, sut.nodesCount)
	stakedAmounts := make([]uint64, sut.nodesCount)
	sut.withEachNodeHome(func(i int, home string) {
		clix := NewTgradeCliX(t, sut.rpcAddr, sut.chainID, filepath.Join(workDir, home), verbose)
		addr := clix.GetKeyAddr(fmt.Sprintf("node%d", i))
		engagementGroup[i] = testingcontracts.TG4Member{
			Addr:   addr,
			Weight: uint64(sut.nodesCount - i), // unique weight
		}
		initialStakedTokenAmount := sdk.TokensFromConsensusPower(100) //set via testnet command
		stakedAmounts[i] = initialStakedTokenAmount.Uint64()
	})

	sut.StartChain(t)
	sut.AwaitNextBlock(t)

	engagementGroupAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "ENGAGEMENT"), "address").String()
	valsetAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "VALSET"), "address").String()

	// and smart query internal list of validators
	qResult := cli.CustomQuery("q", "wasm", "contract-state", "smart", valsetAddr, `{"list_active_validators":{}}`)
	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, sut.nodesCount, qResult)
	t.Log("got query result", qResult)

	sortedMember := testingcontracts.SortByWeightDesc(engagementGroup)
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount)

	if sut.nodesCount < 4 {
		t.Skip("4 nodes required")
	}
	initialValBalances := make(map[string]int64, len(sortedMember))
	for _, v := range sortedMember {
		initialValBalances[v.Addr] = cli.QueryBalance(v.Addr, "utgd")
	}
	initialSupply := cli.QueryTotalSupply("utgd")

	// And when removed from **engagement** group
	engagementUpdateMsg := testingcontracts.TG4UpdateMembersMsg{
		Remove: []string{sortedMember[0].Addr},
	}
	eResult := cli.Execute(engagementGroupAddr, engagementUpdateMsg.Json(t), tg4AdminAddr)
	RequireTxSuccess(t, eResult)
	t.Log("got execution result", eResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	awaitValsetEpochCompleted(t)

	// then validator set is updated
	// with unengaged validator removed
	sortedMember = sortedMember[1:sut.nodesCount]
	stakedAmounts = stakedAmounts[1:sut.nodesCount]
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount-1)

	// and new tokens were minted
	assert.Greater(t, cli.QueryTotalSupply("utgd"), initialSupply)

	// and distributed to the validators
	var distributed bool
	for _, v := range sortedMember {
		if initialValBalances[v.Addr] < cli.QueryBalance(v.Addr, "utgd") {
			distributed = true
		}
	}
	assert.True(t, distributed, "no tokens distributed")

	// And when moniker updated
	myAddr := cli.GetKeyAddr("node0")
	txResult := cli.CustomCommand("tx", "poe", "edit-validator", "--moniker=newMoniker", "--from=node0")
	RequireTxSuccess(t, txResult)
	qResult = cli.QueryValidator(myAddr)
	assert.Equal(t, "newMoniker", gjson.Get(qResult, "description.moniker").String())
}

func TestPoEAddPostGenesisValidator(t *testing.T) {
	// Scenario:
	// given a running chain
	// when a create-validator message is submitted with self delegation amount > min
	// then the validator gets engagement points automatically
	// and is added to the active validator set
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	sut.ModifyGenesisJson(t,
		SetPoEParamsMutator(t, poetypes.NewParams(100, 10, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(5))))),
	)
	sut.StartChain(t)

	newNode := sut.AddFullnode(t)
	sut.AwaitNodeUp(t, fmt.Sprintf("http://127.0.0.1:%d", newNode.RPCPort))
	opAddr := cli.AddKey("newOperator")
	cli.FundAddress(opAddr, "1000utgd")
	newPubKey, pubKeyAddr := loadValidatorPubKey(t, filepath.Join(workDir, sut.nodePath(sut.nodesCount-1), "config", "priv_validator_key.json"))
	// when
	txResult := cli.CustomCommand("tx", "poe", "create-validator", "--moniker=newMoniker", "--amount=10utgd",
		"--pubkey="+pubKeyAddr, "--from=newOperator")
	RequireTxSuccess(t, txResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	awaitValsetEpochCompleted(t)

	// then
	valResult := cli.GetTendermintValidatorSet()
	var found bool
	for _, v := range valResult.Validators {
		if v.PubKey.Equals(newPubKey) {
			found = true
			break
		}
	}
	assert.True(t, found, "not in validator set : %#v", valResult)
}

func TestPoEQueries(t *testing.T) {
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	sut.StartChain(t)
	specs := map[string]struct {
		query  []string
		assert func(t *testing.T, qResult string)
	}{
		"unbonding period": {
			query: []string{"q", "poe", "unbonding-period"},
			assert: func(t *testing.T, qResult string) {
				gotTime := gjson.Get(qResult, "time").String()
				assert.Equal(t, "1814400s", gotTime)
			},
		},
		"validators": {
			query: []string{"q", "poe", "validators"},
			assert: func(t *testing.T, qResult string) {
				gotValidators := gjson.Get(qResult, "validators").Array()
				assert.Greater(t, len(gotValidators), 0, gotValidators)
				assert.NotEmpty(t, gjson.Get(gotValidators[0].String(), "description.moniker"), "moniker")
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			qResult := cli.CustomQuery(spec.query...)
			spec.assert(t, qResult)
			t.Logf(qResult)
		})
	}
}

func assertValidatorsUpdated(t *testing.T, sortedMember []testingcontracts.TG4Member, stakedAmounts []uint64, expValidators int) {
	t.Helper()
	v := sut.RPCClient(t).Validators()
	require.Len(t, v, expValidators, "got %#v", v)
	for i := 0; i < expValidators; i++ {
		// ordered by power desc
		expWeight := int64(math.Sqrt(float64(sortedMember[i].Weight * stakedAmounts[i]))) // function implemented in mixer
		assert.Equal(t, expWeight, v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}
}

// SetPoEParams set in genesis
func SetPoEParamsMutator(t *testing.T, params poetypes.Params) GenesisMutator {
	return func(genesis []byte) []byte {
		t.Helper()
		val, err := json.Marshal(params)
		require.NoError(t, err)
		state, err := sjson.SetRawBytes(genesis, "app_state.poe.params", val)
		require.NoError(t, err)
		return state
	}
}

func awaitValsetEpochCompleted(t *testing.T) {
	// wait for update manifests in valset (epoch has completed)
	time.Sleep(time.Second)
	sut.AwaitNextBlock(t)
}
