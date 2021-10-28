package contract_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryUnbondingPeriod(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, _ := setupPoEContracts(t)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)

	// when
	adaptor := contract.NewStakeContractAdapter(contractAddr, example.TWasmKeeper, nil)

	res, err := adaptor.QueryStakingUnbondingPeriod(ctx)

	// then
	require.NoError(t, err)
	const configuredTime = 21 * 24 * 60 * 60 * time.Second // in bootstrap
	assert.Equal(t, configuredTime, res)
}

func TestQueryValidatorUnboding(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)

	op1Addr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)

	// unbond some tokens for operator 1
	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now).WithBlockHeight(12)
	unbondedAmount := sdk.NewInt(10)
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	err = contract.UnbondDelegation(ctx, contractAddr, op1Addr, unbondedAmount, example.TWasmKeeper.GetContractKeeper())
	require.NoError(t, err)

	op2Addr, err := sdk.AccAddressFromBech32(vals[1].OperatorAddress)
	require.NoError(t, err)
	unbodingPeriod, err := example.PoEKeeper.StakeContract(ctx).QueryStakingUnbondingPeriod(ctx)
	require.NoError(t, err)

	specs := map[string]struct {
		srcOpAddr sdk.AccAddress
		expResult []stakingtypes.UnbondingDelegationEntry
	}{
		"unbondings": {
			srcOpAddr: op1Addr,
			expResult: []stakingtypes.UnbondingDelegationEntry{
				{
					InitialBalance: sdk.NewInt(10),
					Balance:        sdk.NewInt(10),
					CompletionTime: now.Add(unbodingPeriod).UTC(),
					CreationHeight: 12,
				},
			},
		},
		"no unbondings with existing operator": {
			srcOpAddr: op2Addr,
			expResult: []stakingtypes.UnbondingDelegationEntry{},
		},
		"unknown operator": {
			srcOpAddr: rand.Bytes(sdk.AddrLen),
			expResult: []stakingtypes.UnbondingDelegationEntry{},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			adapter := contract.NewStakeContractAdapter(contractAddr, example.TWasmKeeper, nil)
			gotRes, gotErr := adapter.QueryStakingUnbonding(ctx, spec.srcOpAddr)
			// then
			require.NoError(t, gotErr)
			require.NotNil(t, gotRes)
			assert.Equal(t, spec.expResult, gotRes)
		})
	}
}
