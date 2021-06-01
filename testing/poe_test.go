// +build system_test

package testing

import (
	"encoding/base64"
	"fmt"
	testingcontracts "github.com/confio/tgrade/testing/contracts"
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
		engagementGroupAddr = ContractBech32Address(1, 1)
		stakerGroupAddr     = ContractBech32Address(2, 2)
		mixerAddr           = ContractBech32Address(3, 3)
		valsetAddr          = ContractBech32Address(4, 4)
		anyAddress          = "tgrade12qey0qvmkvdu5yl3x329lhrvqfgzs5vne225q7"
		tg4AdminAddr        = cli.AddKey("tg4admin")
	)
	// prepare contract init messages with chain validator data
	tg4EngagementInitMsg := testingcontracts.TG4GroupInitMsg{
		Admin:    tg4AdminAddr,
		Members:  make([]testingcontracts.TG4Member, sut.nodesCount),
		Preauths: 1,
	}
	tg4StakerInitMsg := testingcontracts.TG4StakeInitMsg{
		Admin:           tg4AdminAddr,
		Denom:           testingcontracts.Denom{Native: "utgd"},
		MinBond:         "1",
		TokensPerWeight: "1",
		UnbondingPeriod: testingcontracts.UnbodingPeriod{
			TimeInSec: uint64(time.Hour.Seconds()),
		},
		Preauths: 1,
	}
	tg4MixerInitMsg := testingcontracts.TG4MixerInitMsg{
		LeftGroup:  engagementGroupAddr,
		RightGroup: stakerGroupAddr,
	}
	valsetInitMsg := testingcontracts.ValsetInitMsg{
		Membership:    mixerAddr,
		MinWeight:     1,
		MaxValidators: 100,
		EpochLength:   1,
		InitialKeys:   make([]testingcontracts.ValsetInitKey, sut.nodesCount),
	}
	stakers := make(map[string]sdk.Coin, sut.nodesCount)
	stakedAmounts := make([]int, sut.nodesCount)
	sut.withEachNodeHome(func(i int, home string) {
		k := readPubkey(t, filepath.Join(workDir, home, "config", "priv_validator_key.json"))
		pubKey := base64.StdEncoding.EncodeToString(k.Bytes())
		addr := cli.AddKey(fmt.Sprintf("node%d-owner", i))
		tg4EngagementInitMsg.Members[i] = testingcontracts.TG4Member{
			Addr:   addr,
			Weight: sut.nodesCount - i, // unique weight
		}
		valsetInitMsg.InitialKeys[i] = testingcontracts.NewValsetInitKey(addr, pubKey)
		const initialStakedTokenAmount = 10
		stakers[addr] = sdk.NewCoin("utgd", sdk.NewInt(initialStakedTokenAmount))
		stakedAmounts[i] = int(initialStakedTokenAmount)

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
			"testing/contracts/tg4_stake.wasm",
			"--instantiate-everybody=true",
			"--builder=foo/bar:latest",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"store",
			"testing/contracts/tg4_mixer.wasm",
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
		// now instantiate
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"1",
			tg4EngagementInitMsg.Json(t),
			"--label=engagement",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"2",
			tg4StakerInitMsg.Json(t),
			"--label=stakers",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"3",
			tg4MixerInitMsg.Json(t),
			"--label=poe",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		{
			"wasm-genesis-message",
			"instantiate-contract",
			"4",
			valsetInitMsg.Json(t),
			"--label=valset",
			fmt.Sprintf("--run-as=%s", anyAddress),
		},
		// and set privilege
		{
			"wasm-genesis-flags",
			"set-privileged",
			valsetAddr,
		},
	}
	// add stakers with some initial amount
	for addr, amount := range stakers {
		commands = append(commands, [][]string{
			{
				"add-genesis-account",
				addr,
				"1000utgd",
			},
			{
				"wasm-genesis-message",
				"execute",
				stakerGroupAddr,
				testingcontracts.TG4StakeExecute{Bond: &struct{}{}}.Json(t),
				fmt.Sprintf("--run-as=%s", addr),
				fmt.Sprintf("--amount=%s", amount.String()),
			}}...)
	}
	sut.ModifyGenesisCLI(t, commands...)
	sut.StartChain(t)
	cli.FundAddress(t, tg4AdminAddr, "1000utgd")
	sut.AwaitNextBlock(t)

	// and smart query internal list of validators
	qResult := cli.CustomQuery("q", "wasm", "contract-state", "smart", valsetAddr, `{"list_active_validators":{}}`)
	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, sut.nodesCount, qResult)
	t.Log("got query result", qResult)

	sortedMember := testingcontracts.SortByWeight(tg4EngagementInitMsg.Members)
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount)

	// And when **stake** increased for owner with lowest engagement points
	stakeUpdateMsg := testingcontracts.TG4StakeExecute{Bond: &struct{}{}}
	const additionalStakedAmount = 100
	stakedAmounts[sut.nodesCount-1] += additionalStakedAmount
	eResult := cli.Execute(stakerGroupAddr, stakeUpdateMsg.Json(t), fmt.Sprintf("node%d-owner", sut.nodesCount-1), sdk.NewCoin("utgd", sdk.NewInt(additionalStakedAmount)))
	RequireTxSuccess(t, eResult)
	t.Log("got execution result", eResult)
	// wait for msg execution
	sut.AwaitNextBlock(t)
	// wait for update manifests in valset (epoch has completed)
	time.Sleep(1 * time.Second)
	sut.AwaitNextBlock(t)

	// then validator set is updated
	// with lowest engaged member became the validator with highest power
	sortedMember = append([]testingcontracts.TG4Member{sortedMember[sut.nodesCount-1]}, sortedMember[0:sut.nodesCount-1]...)
	stakedAmounts = append([]int{stakedAmounts[sut.nodesCount-1]}, stakedAmounts[0:sut.nodesCount-1]...)
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount)

	// And when removed from **engagement** group
	engagementUpdateMsg := testingcontracts.TG4UpdateMembersMsg{
		Remove: []string{sortedMember[0].Addr},
	}
	eResult = cli.Execute(engagementGroupAddr, engagementUpdateMsg.Json(t), tg4AdminAddr)
	RequireTxSuccess(t, eResult)
	t.Log("got execution result", eResult)
	// wait for msg execution
	sut.AwaitNextBlock(t, sut.blockTime*5)
	// wait for update manifests in valset (epoch has completed)
	sut.AwaitNextBlock(t)

	// then validator set is updated
	// with unengaged validator removed
	sortedMember = sortedMember[1:sut.nodesCount]
	stakedAmounts = stakedAmounts[1:sut.nodesCount]
	assertValidatorsUpdated(t, sortedMember, stakedAmounts, sut.nodesCount-1)
}

func assertValidatorsUpdated(t *testing.T, sortedMember []testingcontracts.TG4Member, stakedAmounts []int, expValidators int) {
	t.Helper()
	v := sut.RPCClient(t).Validators()
	require.Len(t, v, expValidators, "got %#v", v)
	for i := 0; i < expValidators; i++ {
		// ordered by power desc
		expWeight := int64(math.Sqrt(float64(sortedMember[i].Weight * stakedAmounts[i]))) // function implemented in mixer
		assert.Equal(t, expWeight, v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}
}
