//go:build system_test
// +build system_test

package testing

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

// SetGlobalMinFee set the passed coins to the global minimum fee
func SetGlobalMinFee(t *testing.T, fees ...sdk.DecCoin) GenesisMutator {
	return func(genesis []byte) []byte {
		t.Helper()
		coins := sdk.NewDecCoins(fees...)
		require.NoError(t, coins.Validate())
		val, err := json.Marshal(coins)
		require.NoError(t, err)
		state, err := sjson.SetRawBytes(genesis, "app_state.globalfee.params.minimum_gas_prices", val)
		require.NoError(t, err)
		return state
	}
}
