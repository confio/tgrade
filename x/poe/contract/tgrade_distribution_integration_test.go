package contract_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryWithdrawableFunds(t *testing.T) {
	// setup contracts and seed some data
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, vals := withRandomValidators(t, ctx, example, 1)
	gs := types.GenesisStateFixture(mutator)
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	require.NoError(t, err)
	opAddr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)

	// empty rewards
	gotAmount, gotErr := contract.QueryWithdrawableFunds(ctx, example.TWasmKeeper, contractAddr, opAddr)
	require.NoError(t, gotErr)
	assert.Equal(t, sdk.NewDecCoin("utgd", sdk.ZeroInt()), gotAmount)

	// and when one epoch has passed
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultGenesisState().ValsetContractConfig.EpochLength))
	module.EndBlock(ctx, abci.RequestEndBlock{})
	// then
	gotAmount, gotErr = contract.QueryWithdrawableFunds(ctx, example.TWasmKeeper, contractAddr, opAddr)
	require.NoError(t, gotErr)
	assert.True(t, gotAmount.IsGTE(sdk.NewDecCoin("utgd", sdk.NewInt(1))), gotAmount)
}
