//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestGlobalFee(t *testing.T) {
	sut.ResetChain(t)
	sut.ModifyGenesisJson(t, SetGlobalMinFee(t,
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
	txResult := cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1500000", "--fees=1utg")
	RequireTxFailure(t, txResult, "insufficient fee")

	t.Log("Any transaction with enough fees should pass")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1500000", "--fees=1500utgd")
	RequireTxSuccess(t, txResult)

	t.Log("Any transaction with enough alternative fee token amount should pass")
	txResult = cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=node0", "--gas=1500000", "--fees=150node0token")
	RequireTxSuccess(t, txResult)
}

func TestFeeDistribution(t *testing.T) {
	// scenario:
	// when a transaction with high fees is submitted
	// then the fees are distributed to the validators
	sut.ResetChain(t)
	sut.ModifyGenesisJson(t, SetAllEngagementPoints(t, 1))
	sut.StartChain(t)

	cli := NewTgradeCli(t, sut, verbose)
	cli.FundAddress(cli.AddKey("myFatFingerKey"), "20000000utgd")
	oldBalances := make([]int64, sut.nodesCount)
	for i := 0; i < sut.nodesCount; i++ {
		oldBalances[i] = cli.QueryBalance(cli.GetKeyAddr(fmt.Sprintf("node%d", i)), "utgd")
	}

	// when
	const anyContract = "testing/contract/hackatom.wasm.gzip"
	txResult := cli.CustomCommand("tx", "wasm", "store", anyContract, "--from=myFatFingerKey", "--gas=1500000", "--fees=20000000utgd")
	RequireTxSuccess(t, txResult)
	AwaitValsetEpochCompleted(t) // so that fees are distributed

	// TODO: no more fees are distributed. Rather they are held in a new contract to be withdrawn.
	// DO proper fix in issue #156, so we can query the pending stake. For now I will disable
	// then
	//for i := 0; i < sut.nodesCount; i++ {
	//	newBalance := cli.QueryBalance(cli.GetKeyAddr(fmt.Sprintf("node%d", i)), "utgd")
	//	diff := newBalance - oldBalances[i]
	//	t.Logf("Block rewards: %d\n", diff)
	//	assert.LessOrEqualf(t, int64(20000000/sut.nodesCount), diff, "got diff: %d (before %d after %d)", diff, oldBalances[i], newBalance)
	//}
}
