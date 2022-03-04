//go:build system_test
// +build system_test

package testing

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Scenario: add a contract without migrator address
// 			 trigger gov proposal to upgrade chain to v07
//			 then verify that contract has validator-voting contract as migrator set
func TestUpgradev07(t *testing.T) {
	cli := NewTgradeCli(t, sut, verbose)
	myKey := cli.GetKeyAddr("node0")
	require.NotEmpty(t, myKey)

	sut.ResetChain(t)
	sut.StartChain(t)

	codeID := cli.StoreWasm("x/poe/contract/tgrade_gov_reflect.wasm")
	myContractAddr := cli.InstantiateWasm(codeID, "{}")

	// assert no migrator set
	qResult := cli.CustomQuery("q", "wasm", "contract", myContractAddr)
	oldMigrator := gjson.Get(qResult, "contract_info.admin").String()
	require.Empty(t, oldMigrator, qResult)
	valVotingAddr := gjson.Get(cli.CustomQuery("q", "poe", "contract-address", "VALIDATOR_VOTING"), "address").String()
	require.NotEmpty(t, valVotingAddr, qResult)

	// when
	t.Log("Send a proposal start the upgrade")
	plannedHeight := sut.currentHeight + 4 // height needs to be in the future when proposal is executed
	proposalMsg := fmt.Sprintf(`{"propose": {"title": "Upgrade", "description": "Testing", "proposal": {"register_upgrade": {"name":"v07", "height": %d, "info": "%s"}} }}`, plannedHeight, myContractAddr)
	txResult := cli.CustomCommand("tx", "wasm", "execute", valVotingAddr, proposalMsg, fmt.Sprintf("--from=%s", myKey))
	RequireTxSuccess(t, txResult)
	proposalID := gjson.Get(txResult, "logs.#.events.#.attributes.#(key=proposal_id).value").Array()[0].Array()[0].Int()
	require.NotEmpty(t, codeID)

	// and have all voting
	t.Log("Send votes")
	voteMsg := fmt.Sprintf(`{"vote": {"proposal_id": %d, "vote": "yes"}}`, proposalID)
	var wg sync.WaitGroup
	wg.Add(sut.nodesCount - 1)
	for i := 1; i < sut.nodesCount; i++ {
		go func(i int) {
			// exec may fail due to "end-early" settings
			_ = cli.CustomCommand("tx", "wasm", "execute", valVotingAddr, voteMsg, fmt.Sprintf("--from=node%d", i))
			wg.Done()
		}(i)
	}
	wg.Wait()

	// and execute
	t.Log("Send an execute to trigger the upgrade")
	execMsg := fmt.Sprintf(`{"execute": {"proposal_id": %d}}`, proposalID)
	txResult = cli.CustomCommand("tx", "wasm", "execute", valVotingAddr, execMsg, fmt.Sprintf("--from=%s", myKey))
	RequireTxSuccess(t, txResult)

	sut.AwaitBlockHeight(t, plannedHeight)
	// then the migrator address should be set
	qResult = cli.CustomQuery("q", "wasm", "contract", myContractAddr)
	newMigrator := gjson.Get(qResult, "contract_info.admin").String()
	require.Equal(t, valVotingAddr, newMigrator, qResult)
}
