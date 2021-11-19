package contract_test

import (
	"github.com/tendermint/tendermint/libs/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestSlashValidator(t *testing.T) {
	var systemAdmin sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t, setSystemAdminMutator(systemAdmin))

	ocProposeAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeOversightCommunityGovProposals)
	require.NoError(t, err)
	engageAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	require.NoError(t, err)
	opAddr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)

	// check initial engagement points
	points, err := contract.QueryTG4Member(ctx, example.TWasmKeeper, engageAddr, opAddr)
	require.NoError(t, err)
	require.NotNil(t, points)
	t.Logf("Initial Engagement: %d", *points)

	info := example.TWasmKeeper.GetContractInfo(ctx, engageAddr)
	require.NotNil(t, info)
	admin, err := sdk.AccAddressFromBech32(info.Admin)
	require.NoError(t, err)
	assert.Equal(t, systemAdmin.String(), admin.String())

	// slash some
	props := contract.NewOCProposalsContractAdapter(ocProposeAddr, example.TWasmKeeper, nil)
	err = props.ProposeSlash(ctx, opAddr, *contract.DecimalFromProMille(500), systemAdmin)
	require.NoError(t, err)

	// get the proposal id
	latest, err := props.LatestProposal(ctx)
	require.NoError(t, err)
	require.NotNil(t, latest)

	// execute the contract
	err = props.ExecuteProposal(ctx, latest.ID, systemAdmin)
	require.NoError(t, err)

	// check the points have gone down
	slashed, err := contract.QueryTG4Member(ctx, example.TWasmKeeper, engageAddr, opAddr)
	require.NoError(t, err)
	require.NotNil(t, slashed)
	t.Logf("Final Engagement: %d", *slashed)
	// this is not always the same as half due to rounding
	expected := *points - (*points / 2)
	assert.Equal(t, expected, *slashed)
}

func setSystemAdminMutator(admin sdk.AccAddress) func(m *types.GenesisState) {
	return func(m *types.GenesisState) {
		m.SystemAdminAddress = admin.String()
	}
}
