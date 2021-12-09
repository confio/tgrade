package contract_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

//go:embed tg4_stake.wasm
var randomContract []byte

func TestValidatorsGovProposal(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)
	require.Len(t, vals, 3)

	op1Addr, _ := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	anyAddress := types.RandomAccAddress()

	contractKeeper := example.TWasmKeeper.GetContractKeeper()

	valVotingAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValidatorVoting)
	require.NoError(t, err)
	distrAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	require.NoError(t, err)
	// ensure members set
	members, err := contract.QueryTG4Members(ctx, example.TWasmKeeper, distrAddr)
	require.NoError(t, err)
	require.Len(t, members, 3)
	for _, m := range members {
		t.Logf("%s : %d\n", m.Addr, m.Weight)
	}
	// upload any contract that is not pinned
	codeID, err := contractKeeper.Create(ctx, anyAddress, randomContract, nil)
	require.NoError(t, err)
	require.False(t, example.TWasmKeeper.IsPinnedCode(ctx, codeID), "pinned")
	specs := map[string]struct {
		src       contract.ValidatorProposal
		assertExp func(t *testing.T, ctx sdk.Context)
	}{
		"pin code": {
			src: contract.ValidatorProposal{
				PinCodes: []uint64{codeID},
			},
			assertExp: func(t *testing.T, ctx sdk.Context) {
				assert.True(t, example.TWasmKeeper.IsPinnedCode(ctx, codeID), "pinned")
			},
		},
		"chain upgrade": {
			src: contract.ValidatorProposal{
				RegisterUpgrade: &contract.ChainUpgrade{
					Name:   "v2",
					Info:   "v2-info",
					Height: 7654321,
				},
			},
			assertExp: func(t *testing.T, ctx sdk.Context) {
				gotPlan, exists := example.UpgradeKeeper.GetUpgradePlan(ctx)
				assert.True(t, exists, "exists")
				exp := upgradetypes.Plan{
					Name:   "v2",
					Info:   "v2-info",
					Height: 7654321,
				}
				assert.Equal(t, exp, gotPlan)
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
			// when submit proposal
			proposalMsg := contract.ValidatorVotingExecuteMsg{
				Propose: &contract.ValidatorVotingPropose{
					Title:       "My proposal",
					Description: "My description",
					Proposal:    spec.src,
				},
			}
			msgBz, err := json.Marshal(proposalMsg)
			require.NoError(t, err)
			_, err = contractKeeper.Execute(ctx, valVotingAddr, op1Addr, msgBz, nil)
			require.NoError(t, err, "exec: %s", string(msgBz))

			// then it is persisted
			adapter := contract.NewVotingContractAdapter(valVotingAddr, example.TWasmKeeper, nil)
			rsp, err := adapter.LatestProposal(ctx)
			require.NoError(t, err)
			require.Equal(t, contract.ProposalStatusOpen, rsp.Status)
			myProposalID := rsp.ID
			t.Logf("%d %s- voting power: %s\n", 0, op1Addr.String(), vals[0].Tokens)

			// and when all validators vote
			// first val has auto YES due to submission, let another one vote
			for i, val := range vals[1:] {
				t.Logf("%d %s - voting power: %s\n", i+1, val.OperatorAddress, val.Tokens)
				opAddr, _ := sdk.AccAddressFromBech32(val.OperatorAddress)
				require.NoError(t, adapter.VoteProposal(ctx, myProposalID, contract.YES_VOTE, opAddr), "voter: %d", i)
			}
			// then
			rsp, err = adapter.QueryProposal(ctx, myProposalID)
			require.NoError(t, err)
			require.Equal(t, contract.ProposalStatusPassed, rsp.Status)

			// and when execute proposal
			require.NoError(t, adapter.ExecuteProposal(ctx, myProposalID, op1Addr))

			// then
			rsp, err = adapter.QueryProposal(ctx, myProposalID)
			require.NoError(t, err)
			require.Equal(t, contract.ProposalStatusExecuted, rsp.Status)
			// and verify action state
			spec.assertExp(t, ctx)
		})
	}

}
