package contract_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryWithdrawableFunds(t *testing.T) {
	t.Skip("critical issue: https://github.com/confio/tgrade-contracts/issues/385")
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	require.NoError(t, err)
	opAddr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)

	specs := map[string]struct {
		setup      func(ctx sdk.Context) sdk.Context
		src        sdk.AccAddress
		expRewards bool
		expErr     *sdkerrors.Error
	}{
		"empty rewards": {
			src:        opAddr,
			expRewards: false,
		},
		"with rewards after epoche": {
			setup: func(ctx sdk.Context) sdk.Context {
				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultGenesisState().ValsetContractConfig.EpochLength))
				poe.EndBlocker(ctx, example.TWasmKeeper)
				return ctx
			},
			src:        opAddr,
			expRewards: true,
		},
		"unknown address": {
			src:        rand.Bytes(sdk.AddrLen),
			expRewards: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			tCtx, _ := ctx.CacheContext()
			if spec.setup != nil {
				tCtx = spec.setup(tCtx)
			}
			gotAmount, gotErr := contract.NewDistributionContractAdapter(contractAddr, example.TWasmKeeper, nil).ValidatorOutstandingReward(tCtx, spec.src)
			if spec.expErr != nil {
				assert.True(t, spec.expErr.Is(gotErr), "got %s", gotErr)
				return
			}
			require.NoError(t, gotErr)

			if spec.expRewards {
				assert.True(t, gotAmount.IsGTE(sdk.NewCoin("utgd", sdk.OneInt())))
			} else {
				assert.Equal(t, sdk.NewCoin("utgd", sdk.ZeroInt()), gotAmount)
			}
		})
	}
}
