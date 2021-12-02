package contract_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

//go:embed tg4_stake.wasm
var randomContract []byte

func TestValidatorsGovProposal(t *testing.T) {
	anyAddress := types.RandomAccAddress()
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)
	require.Len(t, vals, 3)
	contractKeeper := example.TWasmKeeper.GetContractKeeper()
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	valVotingAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValidatorVoting)
	require.NoError(t, err)

	// upload any contract that is not pinned
	codeID, err := contractKeeper.Create(ctx, anyAddress, randomContract, nil)
	require.NoError(t, err)
	require.False(t, example.TWasmKeeper.IsPinnedCode(ctx, codeID), "pinned")

	// when submit proposal to pin
	proposalMsg := contract.ValidatorVotingExecuteMsg{
		Propose: &contract.ValidatorVotingPropose{
			Title:       "My proposal",
			Description: "My description",
			Proposal: contract.ValidatorProposal{
				PinCodes: &contract.CodeIDsWrapper{
					CodeIDs: []uint64{codeID},
				},
			},
		},
	}
	op1Addr, _ := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	msgBz, err := json.Marshal(proposalMsg)
	require.NoError(t, err)
	_, err = contractKeeper.Execute(ctx, valVotingAddr, op1Addr, msgBz, nil)
	require.NoError(t, err)

	// then is persisted
	adapter := contract.NewVotingContractAdapter(valVotingAddr, example.TWasmKeeper, nil)
	rsp, err := adapter.LatestProposal(ctx)
	require.NoError(t, err)
	require.Equal(t, contract.ProposalStatusOpen, rsp.Status)
	myProposalID := rsp.ID

	// and when all validators vote
	// first val has auto YES due to submission, let another one vote
	for _, val := range vals[1:] {
		opAddr, _ := sdk.AccAddressFromBech32(val.OperatorAddress)
		_ = adapter.VoteProposal(ctx, myProposalID, contract.YES_VOTE, opAddr) // todo: fix power  so that consensus is not reached randomly
		//require.NoError(t, voteErr, "voter: %d", i)
		rsp, err = adapter.LatestProposal(ctx)
		require.NoError(t, err)
		t.Logf("proposal status: %s\n", rsp.Status)
	}
	// then
	rsp, err = adapter.LatestProposal(ctx)
	require.NoError(t, err)
	require.Equal(t, contract.ProposalStatusPassed, rsp.Status)

	// and when execute proposal
	require.NoError(t, adapter.ExecuteProposal(ctx, myProposalID, op1Addr))

	// then
	assert.True(t, example.TWasmKeeper.IsPinnedCode(ctx, codeID), "pinned")
	rsp, err = adapter.LatestProposal(ctx)
	require.NoError(t, err)
	require.Equal(t, contract.ProposalStatusExecuted, rsp.Status)
}
