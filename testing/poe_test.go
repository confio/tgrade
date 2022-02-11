//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"math"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/confio/tgrade/app"
	testingcontract "github.com/confio/tgrade/testing/contract"
	poecontracts "github.com/confio/tgrade/x/poe/contract"
	poetypes "github.com/confio/tgrade/x/poe/types"
)

func TestProofOfEngagementSetup(t *testing.T) {
	// scenario:
	//   given: PoE contracts are setup
	//   when: validator is removed from engagement group
	//         via OC gov proposal
	//   then: it is also removed from the active set in tendermint
	//    and: no rewards distributed to it

	t.Skip("Alex: there is currently no way in OC gov proposals to remove engagement points")
	sut.ResetDirtyChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	// contract addresses are deterministic. You can get a list of all contracts in genesis via
	// `tgrade wasm-genesis-message list-contracts --home ./testnet/node0/tgrade`
	var (
		tg4AdminAddr = cli.GetKeyAddr("systemadmin")
	)
	engagementGroup := make([]poecontracts.TG4Member, sut.nodesCount)
	stakedAmounts := make([]uint64, sut.nodesCount)
	sut.withEachNodeHome(func(i int, home string) {
		clix := NewTgradeCliX(t, sut.rpcAddr, sut.chainID, filepath.Join(workDir, home), verbose)
		addr := clix.GetKeyAddr(fmt.Sprintf("node%d", i))
		engagementGroup[i] = poecontracts.TG4Member{
			Addr:   addr,
			Points: uint64(sut.nodesCount - i), // unique weight
		}
		initialStakedTokenAmount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction) //set via testnet command
		stakedAmounts[i] = initialStakedTokenAmount.Uint64()
	})

	sut.StartChain(t)
	sut.AwaitNextBlock(t)

	engagementGroupAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "ENGAGEMENT"), "address").String()
	valsetAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "VALSET"), "address").String()
	distributionAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "DISTRIBUTION"), "address").String()
	assert.NotEmpty(t, distributionAddr)

	// and smart query internal list of validators
	qResult := cli.CustomQuery("q", "wasm", "contract-state", "smart", valsetAddr, `{"list_active_validators":{}}`)
	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, sut.nodesCount, qResult)
	t.Log("got query result", qResult)

	sortedMember := poecontracts.SortByWeightDesc(engagementGroup)
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount)

	if sut.nodesCount < 4 {
		t.Skip("4 nodes min required for > 2/3 alive")
	}
	initialValBalances := make(map[string]int64, len(sortedMember))
	for _, v := range sortedMember {
		initialValBalances[v.Addr] = cli.QueryBalance(v.Addr, "utgd")
	}
	initialSupply := cli.QueryTotalSupply("utgd")

	// And when removed from **engagement** group
	engagementUpdateMsg := poecontracts.UpdateMembersMsg{
		Remove: []string{sortedMember[0].Addr},
	}
	eResult := cli.Execute(engagementGroupAddr, engagementUpdateMsg.Json(t), tg4AdminAddr)
	RequireTxSuccess(t, eResult)
	t.Log("got execution result", eResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then validator set is updated
	// with unengaged validator removed
	sortedMember = sortedMember[1:sut.nodesCount]
	stakedAmounts = stakedAmounts[1:sut.nodesCount]
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount-1)

	// and new tokens were minted
	assert.Greater(t, cli.QueryTotalSupply("utgd"), initialSupply)

	// check rewards distributed
	for _, v := range sortedMember {
		reward := cli.QueryValidatorRewards(v.Addr)
		assert.True(t, reward.IsGTE(sdk.NewDecCoinFromDec("utgd", sdk.OneDec())), "got %s for addr: %s", reward, v.Addr)
	}

	// And when moniker updated
	myAddr := cli.GetKeyAddr("node0")
	txResult := cli.CustomCommand("tx", "poe", "edit-validator", "--moniker=newMoniker", "--from=node0")
	RequireTxSuccess(t, txResult)
	qResult = cli.QueryValidator(myAddr)
	assert.Equal(t, "newMoniker", gjson.Get(qResult, "description.moniker").String())
}

func TestPoEAddPostGenesisValidatorWithAutoEngagementPoints(t *testing.T) {
	// Scenario:
	//   given: a running chain
	//   when: a create-validator message is submitted with self delegation amount > min
	//   then: the validator gets engagement points automatically
	//    and: is added to the active validator set
	cli := NewTgradeCli(t, sut, verbose)
	sut.ModifyGenesisJSON(t,
		SetPoEParamsMutator(t, poetypes.NewParams(100, 10, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(5))))),
	)
	sut.StartChain(t)
	newNode := sut.AddFullnode(t)
	sut.AwaitNodeUp(t, fmt.Sprintf("http://127.0.0.1:%d", newNode.RPCPort))
	opAddr := cli.AddKey("newOperator")
	cli.FundAddress(opAddr, "1000utgd")
	newPubKey := loadValidatorPubKeyForNode(t, sut, sut.nodesCount-1)
	pubKeyEncoded, err := app.MakeEncodingConfig().Codec.MarshalInterfaceJSON(newPubKey)
	require.NoError(t, err)
	// when
	txResult := cli.CustomCommand("tx", "poe", "create-validator", "--moniker=newMoniker", "--amount=10utgd",
		"--pubkey="+string(pubKeyEncoded), "--from=newOperator", "--gas=250000")
	RequireTxSuccess(t, txResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then
	valResult, found := cli.IsInTendermintValset(newPubKey)
	assert.True(t, found, "not in validator set : %#v", valResult)
}

func TestPoEAddPostGenesisValidatorWithGovProposalEngagementPoints(t *testing.T) {
	// Scenario:
	//   given: a running chain
	//   when: a create-validator message is submitted but no EP distributed automatically
	//   then: the validator is not in the active set
	//    and
	//   when: an OC gov proposal adds EP
	//   then: is added to the active validator set
	cli := NewTgradeCli(t, sut, verbose)
	sut.ModifyGenesisJSON(t,
		SetPoEParamsMutator(t, poetypes.NewParams(100, 0, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(5))))),
	)
	sut.StartChain(t)
	systemAdminAddr := cli.GetKeyAddr("systemadmin")
	engagementGroupAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "ENGAGEMENT"), "address").String()

	newNode := sut.AddFullnode(t)
	sut.AwaitNodeUp(t, fmt.Sprintf("http://127.0.0.1:%d", newNode.RPCPort))
	opAddr := cli.AddKey("newOperator")
	cli.FundAddress(opAddr, "1000utgd")
	newPubKey := loadValidatorPubKeyForNode(t, sut, sut.nodesCount-1)
	t.Logf("new operator address %s", opAddr)
	pubKeyEncoded, err := app.MakeEncodingConfig().Codec.MarshalInterfaceJSON(newPubKey)
	require.NoError(t, err)

	// when
	txResult := cli.CustomCommand("tx", "poe", "create-validator", "--moniker=newMoniker", "--amount=10utgd",
		"--pubkey="+string(pubKeyEncoded), "--from=newOperator")
	RequireTxSuccess(t, txResult)
	// wait for msg execution
	sut.AwaitNextBlock(t, defaultWaitTime*2)
	AwaitValsetEpochCompleted(t)

	// then
	valResult, found := cli.IsInTendermintValset(newPubKey)
	assert.False(t, found, "in validator set : %#v", valResult)

	// and new operator should not be in engagement group
	query := poecontracts.TG4Query{Member: &poecontracts.MemberQuery{Addr: opAddr}}
	qResult := cli.CustomQuery("q", "wasm", "contract-state", "smart", engagementGroupAddr, toJson(t, query))
	assert.Empty(t, gjson.Get(qResult, "data.points").String(), qResult)

	// and when
	// val operator added to engagement group via gov
	ocGovPropContractAddr := cli.GetPoEContractAddress("OVERSIGHT_COMMUNITY_PROPOSALS")
	msgBz := toJson(t, testingcontract.OCGovProposalMsg{
		Propose: &testingcontract.OCGovProposalSubmit{
			Title:       "Add my validator",
			Description: "testing",
			Proposal: testingcontract.Proposal{
				GrantEngagement: testingcontract.EngagementMember{
					Addr:   opAddr,
					Points: 10,
				},
			},
		},
	})
	execRsp := cli.Execute(ocGovPropContractAddr, msgBz, systemAdminAddr)
	RequireTxSuccess(t, execRsp)
	AwaitValsetEpochCompleted(t)

	msgBz = toJson(t, testingcontract.OCGovProposalMsg{
		Execute: &testingcontract.OCGovProposalExecute{ProposalID: 1},
	})
	RequireTxSuccess(t, cli.Execute(ocGovPropContractAddr, msgBz, systemAdminAddr))
	AwaitValsetEpochCompleted(t)

	// then new operator should be in engagement group
	qResult = cli.CustomQuery("q", "wasm", "contract-state", "smart", engagementGroupAddr, toJson(t, query))
	assert.Equal(t, int64(10), gjson.Get(qResult, "data.points").Int(), qResult)
	AwaitValsetEpochCompleted(t)

	// and in new validator set
	valResult, found = cli.IsInTendermintValset(newPubKey)
	assert.True(t, found, "not in validator set : %#v", valResult)
}

func TestPoESelfDelegate(t *testing.T) {
	// Scenario:
	// given a running chain
	// when a validator adds stake
	// then their staked amount increases by that amount
	// and the total power increases
	sut.ResetDirtyChain(t)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	qRes := cli.CustomQuery("q", "poe", "self-delegation", cli.GetKeyAddr("node0"))
	amountBefore := gjson.Get(qRes, "balance.amount").Int()
	powerBefore := queryTendermintValidatorPower(t, sut, 0)

	// when
	txResult := cli.CustomCommand("tx", "poe", "self-delegate", "100000utgd", "--from=node0")
	RequireTxSuccess(t, txResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then
	qRes = cli.CustomQuery("q", "poe", "self-delegation", cli.GetKeyAddr("node0"))
	amountAfter := gjson.Get(qRes, "balance.amount").Int()
	assert.Equal(t, int64(100000), amountAfter-amountBefore)

	powerAfter := queryTendermintValidatorPower(t, sut, 0)
	assert.Greater(t, powerAfter, powerBefore)
}

func TestPoEUndelegate(t *testing.T) {
	// Scenario:
	// given a running chain
	// when a validator unbonds stake
	// then their staked amount decreases by that amount
	// and the total power decreases
	// and unbonded amount still locked until auto unbonding happens
	// when unboding time expired
	// then claims got executed automatically

	unbodingPeriod := 10 * time.Second // not too short so that claims not get auto unbonded
	sut.ModifyGenesisJSON(t, SetUnbodingPeriod(t, unbodingPeriod), SetBlockRewards(t, sdk.NewCoin("utgd", sdk.ZeroInt())))
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	qRes := cli.CustomQuery("q", "poe", "self-delegation", cli.GetKeyAddr("node0"))
	delegatedAmountBefore := gjson.Get(qRes, "balance.amount").Int()
	powerBefore := queryTendermintValidatorPower(t, sut, 0)
	balanceBefore := cli.QueryBalance(cli.GetKeyAddr("node0"), "utgd")

	// when
	txResult := cli.CustomCommand("tx", "poe", "unbond", "100000utgd", "--from=node0")
	RequireTxSuccess(t, txResult)
	txResult = cli.CustomCommand("tx", "poe", "unbond", "200000utgd", "--from=node0")
	RequireTxSuccess(t, txResult)
	// wait for msg executions
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then
	qRes = cli.CustomQuery("q", "poe", "self-delegation", cli.GetKeyAddr("node0"))
	delegatedAmountAfter := gjson.Get(qRes, "balance.amount").Int()
	assert.Equal(t, int64(-300000), delegatedAmountAfter-delegatedAmountBefore)

	// the total power decreases
	powerAfter := queryTendermintValidatorPower(t, sut, 0)
	assert.Less(t, powerAfter, powerBefore)

	// account balance not increased, yet
	balanceAfter := cli.QueryBalance(cli.GetKeyAddr("node0"), "utgd")
	require.Equal(t, balanceBefore, balanceAfter)

	// but unbonding delegations pending
	qRes = cli.CustomQuery("q", "poe", "unbonding-delegations", cli.GetKeyAddr("node0"))
	entries := gjson.Get(qRes, "entries").Array()
	assert.Len(t, entries, 2, qRes)

	amounts := gjson.Get(qRes, "entries.#.initial_balance").Array()
	require.Len(t, amounts, 2, qRes)
	assert.Equal(t, int64(100000), amounts[0].Int())
	assert.Equal(t, int64(200000), amounts[1].Int())

	// and when unboding time expired
	expBalance := balanceBefore + 100000 + 200000
	balanceAfter = 0
	for i := 0; i < int(unbodingPeriod/sut.blockTime); i++ {
		balanceAfter = cli.QueryBalance(cli.GetKeyAddr("node0"), "utgd")
		if balanceAfter == expBalance {
			break
		}
		sut.AwaitNextBlock(t)
	}
	// then auto claimed
	require.Equal(t, expBalance, balanceAfter)
	qRes = cli.CustomQuery("q", "poe", "unbonding-delegations", cli.GetKeyAddr("node0"))
	entries = gjson.Get(qRes, "entries").Array()
	assert.Len(t, entries, 0, qRes)
}

func TestPoEQueries(t *testing.T) {
	sut.ResetDirtyChain(t)
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
				assert.Equal(t, "1814400s", gotTime, qResult)
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
		"validators limited": {
			query: []string{"q", "poe", "validators", "--limit=100"},
			assert: func(t *testing.T, qResult string) {
				gotValidators := gjson.Get(qResult, "validators").Array()
				assert.Greater(t, len(gotValidators), 0, gotValidators)
				assert.NotEmpty(t, gjson.Get(gotValidators[0].String(), "description.moniker"), "moniker")
			},
		},
		"validator": {
			query: []string{"q", "poe", "validator", cli.GetKeyAddr("node0")},
			assert: func(t *testing.T, qResult string) {
				assert.NotEmpty(t, gjson.Get(qResult, "description.moniker"), "moniker")
			},
		},
		"historical info": {
			query: []string{"q", "poe", "historical-info", "1"},
			assert: func(t *testing.T, qResult string) {
				gotHeight := gjson.Get(qResult, "header.height").String()
				assert.Equal(t, "1", gotHeight)
			},
		},
		"self delegation": {
			query: []string{"q", "poe", "self-delegation", cli.GetKeyAddr("node0")},
			assert: func(t *testing.T, qResult string) {
				delegatedAmount := gjson.Get(qResult, "balance.amount").Int()
				assert.Equal(t, int64(100000000), delegatedAmount)
			},
		},
		"unbonding delegations": {
			query: []string{"q", "poe", "unbonding-delegations", cli.GetKeyAddr("node0")},
			assert: func(t *testing.T, qResult string) {
				delegatedAmount := gjson.Get(qResult, "entries").Array()
				assert.Len(t, delegatedAmount, 0)
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

func assertValidatorsUpdated(t *testing.T, sortedMember []poecontracts.TG4Member, stakedAmounts []uint64, expValidators int) {
	t.Helper()
	v := sut.RPCClient(t).Validators()
	require.Len(t, v, expValidators, "got %#v", v)
	for i := 0; i < expValidators; i++ {
		// ordered by power desc
		expWeight := int64(math.Sqrt(float64(sortedMember[i].Points * stakedAmounts[i]))) // function implemented in mixer
		assert.Equal(t, expWeight, v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}
}
