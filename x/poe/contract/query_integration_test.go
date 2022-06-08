package contract_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryValidatorSelfDelegation(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	opAddr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)
	selfDelegation := int(sdk.TokensToConsensusPower(vals[0].Tokens, sdk.DefaultPowerReduction))
	specs := map[string]struct {
		srcOpAddr sdk.AccAddress
		expAmount *int
	}{
		"found": {
			opAddr,
			&selfDelegation,
		},
		"unknown": {
			srcOpAddr: rand.Bytes(address.Len),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			res, err := contract.QueryTG4Member(ctx, example.TWasmKeeper, contractAddr, spec.srcOpAddr)
			// then
			require.NoError(t, err)
			assert.Equal(t, spec.expAmount, res)
		})
	}
}
