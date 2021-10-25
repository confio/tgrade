package contract_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"

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

	specs := map[string]struct {
		setup  func(ctx sdk.Context) sdk.Context
		src    sdk.AccAddress
		exp    sdk.DecCoin
		expErr *sdkerrors.Error
	}{
		"empty rewards": {
			src: opAddr,
			exp: sdk.NewDecCoin("utgd", sdk.ZeroInt()),
		},
		"with rewards after epoche": {
			setup: func(ctx sdk.Context) sdk.Context {
				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultGenesisState().ValsetContractConfig.EpochLength))
				module.EndBlock(ctx, abci.RequestEndBlock{})
				return ctx
			},
			src: opAddr,
			exp: sdk.NewDecCoin("utgd", sdk.NewInt(49999)),
		},
		"unknown address": {
			src:    rand.Bytes(sdk.AddrLen),
			exp:    sdk.NewDecCoin("utgd", sdk.ZeroInt()),
			expErr: types.ErrNotFound,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			tCtx, _ := ctx.CacheContext()
			if spec.setup != nil {
				tCtx = spec.setup(tCtx)
			}
			gotAmount, gotErr := contract.QueryWithdrawableFunds(tCtx, example.TWasmKeeper, contractAddr, spec.src)
			if spec.expErr != nil {
				assert.True(t, spec.expErr.Is(gotErr), "got %s", gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, gotAmount)
		})
	}
}
