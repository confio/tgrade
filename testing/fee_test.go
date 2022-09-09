//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	poetypes "github.com/confio/tgrade/x/poe/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGlobalFee(t *testing.T) {
	sut.ModifyGenesisJSON(t, SetGlobalMinFee(t,
		sdk.NewDecCoinFromDec("utgd", sdk.NewDecWithPrec(1, 3)),
		sdk.NewDecCoinFromDec("node0token", sdk.NewDecWithPrec(1, 4))),
	)
	sut.StartChain(t)

	cli := NewTgradeCli(t, sut, verbose)
	rsp := cli.CustomQuery("q", "globalfee", "minimum-gas-prices")
	exp := `[{"denom":"node0token","amount":"0.000100000000000000"},{"denom":"utgd","amount":"0.001000000000000000"}]`
	require.Equal(t, exp, gjson.Get(rsp, "minimum_gas_prices").String())

	const anyContract = "testing/contract/hackatom.wasm.gzip"
	t.Log("Any transaction without enough fees should fail")
	txResult := cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1600000", "--fees=1utg")
	RequireTxFailure(t, txResult, "insufficient fee")

	t.Log("Any transaction with enough fees should pass")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1600000", "--fees=1600utgd")
	RequireTxSuccess(t, txResult)

	t.Log("Any transaction with enough alternative fee token amount should pass")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1600000", "--fees=160node0token")
	RequireTxSuccess(t, txResult)

	t.Log("Transactions with too high fees should fail (fees)")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--fees=101tgd")
	RequireTxFailure(t, txResult)

	t.Log("Transactions with too high fees should fail (gas)")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=101", "--gas-prices=1tgd")
	RequireTxFailure(t, txResult)
}

func TestFeeDistribution(t *testing.T) {
	// scenario:
	// when a transaction with high fees is submitted
	// then the fees are distributed to the validators
	sut.ModifyGenesisJSON(t, SetAllEngagementPoints(t, 1))
	sut.StartChain(t)

	cli := NewTgradeCli(t, sut, verbose)
	cli.FundAddress(cli.AddKey("myFatFingerKey"), "200000000utgd")
	oldBalances := make([]int64, sut.nodesCount)
	for i := 0; i < sut.nodesCount; i++ {
		oldBalances[i] = cli.QueryBalance(cli.GetKeyAddr(fmt.Sprintf("node%d", i)), "utgd")
	}

	// when
	const anyContract = "testing/contract/hackatom.wasm.gzip"
	txResult := cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=myFatFingerKey", "--gas=1600000", "--fees=200000000utgd")
	RequireTxSuccess(t, txResult)
	AwaitValsetEpochCompleted(t) // so that fees are distributed

	// and rewards claimed
	distAddr := cli.GetPoEContractAddress(poetypes.PoEContractTypeDistribution.String())
	for i := 0; i < sut.nodesCount; i++ {
		rsp := cli.Execute(distAddr, `{"withdraw_rewards":{}}`, fmt.Sprintf("node%d", i))
		RequireTxSuccess(t, rsp)
	}
	engAddr := cli.GetPoEContractAddress(poetypes.PoEContractTypeEngagement.String())
	for i := 0; i < sut.nodesCount; i++ {
		rsp := cli.Execute(engAddr, `{"withdraw_rewards":{}}`, fmt.Sprintf("node%d", i))
		RequireTxSuccess(t, rsp)
	}
	// then balance contains rewards
	// 200000000 * 47.5% *(1/ 4 + 1/10) = 33250000 # 1/4 is reserved for all vals, 1/10 is reserved for all EPs
	const expMinRevenue int64 = 33250000
	for i := 0; i < sut.nodesCount; i++ {
		newBalance := cli.QueryBalance(cli.GetKeyAddr(fmt.Sprintf("node%d", i)), "utgd")
		diff := newBalance - oldBalances[i]
		assert.LessOrEqualf(t, expMinRevenue, diff, "node %d got diff: %d (before %d after %d)", i, diff, oldBalances[i], newBalance)
	}
}
