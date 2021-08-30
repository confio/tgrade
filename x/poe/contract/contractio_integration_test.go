package contract_test

import (
	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestSetEngagementPoints(t *testing.T) {
	// setup contracts and seed some data
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, _ := withRandomValidators(t, ctx, example, 2)
	gs := types.GenesisStateFixture(mutator)
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	myOperatorAddr := rand.Bytes(sdk.AddrLen)
	engContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	require.NoError(t, err)

	// when
	err = contract.SetEngagementPoints(ctx, engContractAddr, example.TWasmKeeper, myOperatorAddr, 100)

	// then
	require.NoError(t, err)
	gotPoints, err := contract.QueryTG4Member(ctx, example.TWasmKeeper, engContractAddr, myOperatorAddr)
	require.NoError(t, err)
	require.NotNil(t, gotPoints)
	assert.Equal(t, 100, *gotPoints)
}
