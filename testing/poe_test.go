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
		cw4AdminAddr        = cli.AddKey("cw4admin")
	)
	// prepare contract init messages with chain validator data
	tg4EngagementInitMsg := testingcontracts.CW4InitMsg{
		Admin:    cw4AdminAddr,
		Members:  make([]testingcontracts.CW4Member, sut.nodesCount),
		Preauths: 1,
	}
	tg4StakerInitMsg := testingcontracts.TG4StakeInitMsg{
		Admin:           cw4AdminAddr,
		Denom:           testingcontracts.Denom{Native: "utgd"},
		MinBond:         "1",
		TokensPerWeight: "1",
		UnbondingPeriod: testingcontracts.UnbodingPeriod{
			TimeInSec: uint64(time.Hour.Seconds()),
		},
		Preauths: 1,
	}
	tg4MixerInitMsg := testingcontracts.CW4MixerInitMsg{
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
	sut.withEachNodeHome(func(i int, home string) {
		k := readPubkey(t, filepath.Join(workDir, home, "config", "priv_validator_key.json"))
		pubKey := base64.StdEncoding.EncodeToString(k.Bytes())
		addr := randomBech32Addr()
		tg4EngagementInitMsg.Members[i] = testingcontracts.CW4Member{
			Addr:   addr,
			Weight: sut.nodesCount - i, // unique weight
		}
		valsetInitMsg.InitialKeys[i] = testingcontracts.NewValsetInitKey(addr, pubKey)
		stakers[addr] = sdk.NewCoin("utgd", sdk.OneInt())
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
			"--label=testing",
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

	// and smart query internals
	qResult := cli.CustomQuery("q", "wasm", "contract-state", "smart", valsetAddr, `{"list_active_validators":{}}`)
	validators := gjson.Get(qResult, "data.validators").Array()
	require.Len(t, validators, sut.nodesCount, qResult)
	t.Log("got query result", qResult)

	// and ensure validator count via rpc call
	sut.AwaitNextBlock(t)
	v := sut.RPCClient(t).Validators()
	require.Len(t, v, sut.nodesCount, "got %#v", v)
	sortedMember := testingcontracts.SortByWeight(tg4EngagementInitMsg.Members)
	for i := 0; i < sut.nodesCount; i++ {
		// ordered by power desc
		expWeight := int64(math.Sqrt(float64(sortedMember[i].Weight * 1))) // function implemented in mixer
		assert.Equal(t, expWeight, v[i].VotingPower, "address: %s", encodeBech32Addr(v[i].Address.Bytes()))
	}

	// And when weight updated
	// todo...
}
