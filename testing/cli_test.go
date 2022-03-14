//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Scenario: add WASM code as part of genesis and pin it in VM cache forever
//           for faster execution.
func TestGenesisCodePin(t *testing.T) {
	cli := NewTgradeCli(t, sut, verbose)
	myKey := cli.GetKeyAddr("node0")
	require.NotEmpty(t, myKey)
	t.Logf("key: %q", myKey)
	sut.ResetChain(t)
	// WASM code 1-5 is present.
	sut.ModifyGenesisCLI(t,
		[]string{
			"wasm-genesis-message",
			"store",
			"x/poe/contract/tgrade_gov_reflect.wasm",
			fmt.Sprintf("--run-as=%s", myKey),
		},
		[]string{"wasm-genesis-flags", "set-pinned", "1"},
	)

	// At the time of writing this test, there is no public interface to
	// check if code is cached or not. Instead, we are checking the genesis
	// file only.
	raw := sut.ReadGenesisJSON(t)
	codeIDs := gjson.GetBytes([]byte(raw), "app_state.wasm.pinned_code_ids").Array()
	require.Len(t, codeIDs, 1)
	require.Equal(t, codeIDs[0].Int(), int64(1))
}

func TestUnsafeResetAll(t *testing.T) {
	// scenario:
	// 	given a non empty wasm dir exists in the node home
	//  when `unsafe-reset-all` is executed
	// 	then the dir and all files in it are removed

	wasmDir := filepath.Join(workDir, sut.nodePath(0), "wasm")
	require.NoError(t, os.MkdirAll(wasmDir, os.ModePerm))

	_, err := ioutil.TempFile(wasmDir, "testing")
	require.NoError(t, err)

	// when
	sut.ForEachNodeExecAndWait(t, []string{"unsafe-reset-all"})

	// then
	sut.withEachNodeHome(func(i int, home string) {
		if _, err := os.Stat(wasmDir); !os.IsNotExist(err) {
			t.Fatal("expected wasm dir to be removed")
		}
	})
}

func TestVestingAccounts(t *testing.T) {
	// Scenario:
	//   given: a genesis file
	//   when: a add-genesis-account with vesting flags is executed
	//   then: the vesting account data is added to the genesis
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	vest1Addr := cli.AddKey("vesting1")
	vest2Addr := cli.AddKey("vesting2")
	vest3Addr := cli.AddKey("vesting3")
	myStartTimestamp := time.Now().Add(time.Minute).Unix()
	myEndTimestamp := time.Now().Add(time.Hour).Unix()
	sut.ModifyGenesisCLI(t,
		// delayed vesting no cash
		[]string{"add-genesis-account", vest1Addr, "100000000utgd", "--vesting-amount=100000000utgd", fmt.Sprintf("--vesting-end-time=%d", myEndTimestamp)},
		// continuous vesting no cash
		[]string{"add-genesis-account", vest2Addr, "100000001utgd", "--vesting-amount=100000001utgd", fmt.Sprintf("--vesting-start-time=%d", myStartTimestamp), fmt.Sprintf("--vesting-end-time=%d", myEndTimestamp)},
		// continuous vesting with some cash
		[]string{"add-genesis-account", vest3Addr, "200000002utgd", "--vesting-amount=100000002utgd", fmt.Sprintf("--vesting-start-time=%d", myStartTimestamp), fmt.Sprintf("--vesting-end-time=%d", myEndTimestamp)},
	)
	raw := sut.ReadGenesisJSON(t)
	// delayed vesting: without a start time
	accounts := gjson.GetBytes([]byte(raw), `app_state.auth.accounts.#[@type=="/cosmos.vesting.v1beta1.DelayedVestingAccount"]#`).Array()
	require.Len(t, accounts, 1)
	gotAddr := accounts[0].Get("base_vesting_account.base_account.address").String()
	assert.Equal(t, vest1Addr, gotAddr)
	amounts := accounts[0].Get("base_vesting_account.original_vesting").Array()
	require.Len(t, amounts, 1)
	assert.Equal(t, "utgd", amounts[0].Get("denom").String())
	assert.Equal(t, "100000000", amounts[0].Get("amount").String())
	assert.Equal(t, myEndTimestamp, accounts[0].Get("base_vesting_account.end_time").Int())
	assert.Equal(t, int64(0), accounts[0].Get("start_time").Int())

	// continuous vesting: start time
	accounts = gjson.GetBytes([]byte(raw), `app_state.auth.accounts.#[@type=="/cosmos.vesting.v1beta1.ContinuousVestingAccount"]#`).Array()
	require.Len(t, accounts, 2)
	gotAddr = accounts[0].Get("base_vesting_account.base_account.address").String()
	assert.Equal(t, vest2Addr, gotAddr)
	amounts = accounts[0].Get("base_vesting_account.original_vesting").Array()
	require.Len(t, amounts, 1)
	assert.Equal(t, "utgd", amounts[0].Get("denom").String())
	assert.Equal(t, "100000001", amounts[0].Get("amount").String())
	assert.Equal(t, myEndTimestamp, accounts[0].Get("base_vesting_account.end_time").Int())
	assert.Equal(t, myStartTimestamp, accounts[0].Get("start_time").Int())
	// with some cash
	gotAddr = accounts[1].Get("base_vesting_account.base_account.address").String()
	assert.Equal(t, vest3Addr, gotAddr)
	amounts = accounts[1].Get("base_vesting_account.original_vesting").Array()
	require.Len(t, amounts, 1)
	assert.Equal(t, "utgd", amounts[0].Get("denom").String())
	assert.Equal(t, "100000002", amounts[0].Get("amount").String())
	assert.Equal(t, myEndTimestamp, accounts[0].Get("base_vesting_account.end_time").Int())
	assert.Equal(t, myStartTimestamp, accounts[0].Get("start_time").Int())

	// check accounts have some balances
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(100000000))), getGenesisBalance([]byte(raw), vest1Addr))
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(100000001))), getGenesisBalance([]byte(raw), vest2Addr))
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(200000002))), getGenesisBalance([]byte(raw), vest3Addr))
}

func getGenesisBalance(raw []byte, addr string) sdk.Coins {
	var r []sdk.Coin
	balances := gjson.GetBytes(raw, fmt.Sprintf(`app_state.bank.balances.#[address==%q]#.coins`, addr)).Array()
	for _, coins := range balances {
		for _, coin := range coins.Array() {
			r = append(r, sdk.NewCoin(coin.Get("denom").String(), sdk.NewInt(coin.Get("amount").Int())))
		}
	}
	return r
}
