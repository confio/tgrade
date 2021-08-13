// +build system_test

package testing

import (
	"fmt"
	testingcontracts "github.com/confio/tgrade/x/poe/contract"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
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
	// wait for update manifests in valset (epoch has completed)
	time.Sleep(time.Second)
	sut.AwaitNextBlock(t)

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
