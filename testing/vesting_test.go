//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/app"
	poetypes "github.com/confio/tgrade/x/poe/types"
)

func TestVestingAccountCreatesPostGenesisValidatorAndUndelegates(t *testing.T) {
	// Scenario:
	//   given: a running chain with a vesting account
	//   when: a create-validator message is submitted with self delegation vesting amount > min
	//   then: the validator gets engagement points automatically
	//    and: is added to the active validator set
	//    and: the contract staking account balance is increased
	// and when: block rewards are claimed by the validator operator
	//   then: they are transferable
	// and when: the node operator undelegates all tokens
	// 	 then: the validators power is updated and removed from the active set
	//    and: the vesting account balance is restored
	//    and: the staking account balance is decreased

	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	vest1Addr := cli.AddKey("vesting1")
	myEndTimestamp := time.Now().Add(365 * 24 * time.Hour).Unix()
	sut.ModifyGenesisCLI(t,
		// delayed vesting no cash
		[]string{"add-genesis-account", vest1Addr, "1000000000utgd", "--vesting-amount=1000000000utgd", fmt.Sprintf("--vesting-end-time=%d", myEndTimestamp)},
	)
	sut.modifyGenesisJSON(t,
		SetUnbondingPeriod(t, time.Second),
		SetBlockRewards(t, sdk.NewCoin("utgd", sdk.NewInt(10_000_000))),
	)
	sut.StartChain(t)
	newNode := sut.AddFullnode(t)
	sut.AwaitNodeUp(t, fmt.Sprintf("http://127.0.0.1:%d", newNode.RPCPort))
	newPubKey := loadValidatorPubKeyForNode(t, sut, sut.nodesCount-1)
	pubKeyEncoded, err := app.MakeEncodingConfig().Codec.MarshalInterfaceJSON(newPubKey)
	require.NoError(t, err)
	stakingContractAddr := cli.GetPoEContractAddress("STAKING")

	contractInitialBalance := cli.QueryBalance(stakingContractAddr, "utgd")
	operatorInitialBalance := cli.QueryBalance(vest1Addr, "utgd")
	poolAddr := authtypes.NewModuleAddress(poetypes.BondedPoolName).String()
	poolInitialBalance := cli.QueryBalance(poolAddr, "utgd")

	// enough to enter the validator set
	const stakedVestingAmount = 500_000_000
	// when create validator
	txResult := cli.CustomCommand("tx", "poe", "create-validator", "--moniker=newMoniker", "--amount=0utgd",
		fmt.Sprintf("--vesting-amount=%dutgd", stakedVestingAmount),
		"--pubkey="+string(pubKeyEncoded), "--from=vesting1", "--gas=300000")
	RequireTxSuccess(t, txResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then it is added to the active validator set
	valResult, found := cli.IsInTendermintValset(newPubKey)
	assert.True(t, found, "not in validator set : %#v", valResult)
	// and balances updated
	contractBalanceAfter := cli.QueryBalance(stakingContractAddr, "utgd")
	assert.Equal(t, contractInitialBalance+stakedVestingAmount, contractBalanceAfter, "contract balance")
	operatortBalanceAfter := cli.QueryBalance(vest1Addr, "utgd")
	assert.Equal(t, operatorInitialBalance-stakedVestingAmount, operatortBalanceAfter, "operator balance")
	poolBalanceAfter := cli.QueryBalance(poolAddr, "utgd")
	assert.Equal(t, poolInitialBalance, poolBalanceAfter, "pool balance")
	AwaitValsetEpochCompleted(t)

	// AND when reward coins are claimed
	distrContractAddr := cli.GetPoEContractAddress("DISTRIBUTION")
	execRsp := cli.Execute(distrContractAddr, `{"withdraw_rewards":{}}`, "vesting1")
	RequireTxSuccess(t, execRsp)
	contractBalanceAfter = cli.QueryBalance(stakingContractAddr, "utgd")
	assert.Equal(t, contractInitialBalance+stakedVestingAmount, contractBalanceAfter, "contract balance")
	operatortBalanceAfter = cli.QueryBalance(vest1Addr, "utgd")
	require.Greater(t, operatortBalanceAfter, operatorInitialBalance-stakedVestingAmount, "operator balance")
	poolBalanceAfter = cli.QueryBalance(poolAddr, "utgd")
	require.Equal(t, poolInitialBalance, poolBalanceAfter, "pool balance")

	// then they are transferable
	otherAddr := cli.AddKey("other1")
	transferableAmount := operatortBalanceAfter - (operatorInitialBalance - stakedVestingAmount)
	txResult = cli.CustomCommand("tx", "bank", "send", "vesting1", otherAddr, fmt.Sprintf("%dutgd", transferableAmount))
	RequireTxSuccess(t, txResult)
	require.Equal(t, transferableAmount, cli.QueryBalance(otherAddr, "utgd"))
	operatortBalanceAfter = cli.QueryBalance(vest1Addr, "utgd")
	require.Equal(t, operatorInitialBalance-stakedVestingAmount, operatortBalanceAfter, "operator balance")

	// AND when undelegate
	txResult = cli.CustomCommand("tx", "poe", "unbond", fmt.Sprintf("%dutgd", stakedVestingAmount), "--from=vesting1")
	RequireTxSuccess(t, txResult)

	// claim unbondings
	txResult = cli.Execute(stakingContractAddr, `{"claim":{}}`, "vesting1")
	RequireTxSuccess(t, txResult)

	// wait for msg execution
	sut.AwaitNextBlock(t)
	AwaitValsetEpochCompleted(t)

	// then the validators power is updated and removed from the active set
	valResult, found = cli.IsInTendermintValset(newPubKey)
	assert.False(t, found, "still in validator set : %#v", valResult)

	contractBalanceAfter = cli.QueryBalance(stakingContractAddr, "utgd")
	assert.Equal(t, contractInitialBalance, contractBalanceAfter, "contract balance")
	operatortBalanceAfter = cli.QueryBalance(vest1Addr, "utgd")
	assert.Equal(t, operatorInitialBalance, operatortBalanceAfter, "operator balance")
	poolBalanceAfter = cli.QueryBalance(poolAddr, "utgd")
	assert.Equal(t, poolInitialBalance, poolBalanceAfter, "pool balance")
}

func TestVestingAccountExecutes(t *testing.T) {
	// Scenario:
	//   given: a running chain with a vesting account
	//   when: a message is executed with some deposits amount from a vesting account
	//   then: the TX fails
	sut.ResetChain(t)
	cli := NewTgradeCli(t, sut, verbose)
	vestAddr := cli.AddKey("vesting2")
	myEndTimestamp := time.Now().Add(time.Hour).Unix()
	sut.ModifyGenesisCLI(t,
		// delayed vesting no cash
		[]string{"add-genesis-account", vestAddr, "100000000utgd", "--vesting-amount=100000000utgd", fmt.Sprintf("--vesting-end-time=%d", myEndTimestamp)},
	)
	sut.StartChain(t)

	t.Log("Upload wasm code")
	codeID := cli.StoreWasm("testing/contract/hackatom.wasm.gzip")

	t.Log("Instantiate wasm code")
	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	txResult := cli.CustomCommand("tx", "wasm", "instantiate", strconv.Itoa(codeID), initMsg, "--label=testing", "--from=vesting2", "--gas=1500000", "--amount=1000utgd", "--no-admin")
	RequireTxFailure(t, txResult, "insufficient funds")
}
