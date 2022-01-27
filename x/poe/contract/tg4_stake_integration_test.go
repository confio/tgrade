package contract_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/address"

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
	ctx, example, _, _ := setupPoEContracts(t)

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

func TestQueryStakedAmount(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, _, _ := setupPoEContracts(t)
	contractKeeper := example.TWasmKeeper.GetContractKeeper()
	stakingContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	contractAdapter := contract.NewStakeContractAdapter(stakingContractAddr, example.TWasmKeeper, nil)

	// fund account
	var myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)
	example.Faucet.Fund(ctx, myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(100)))

	var oneInt = sdk.OneInt()
	specs := map[string]struct {
		addr      sdk.AccAddress
		expAmount *sdk.Int
		setup     func(ctx sdk.Context)
		expErr    bool
	}{
		"address has staked amount": {
			addr: myOperatorAddr,
			setup: func(ctx sdk.Context) {
				err := contract.BondDelegation(ctx, stakingContractAddr, myOperatorAddr, sdk.NewCoins(sdk.NewCoin("utgd", sdk.OneInt())), contractKeeper)
				require.NoError(t, err)
			},
			expAmount: &oneInt,
		},
		"address had formerly staked amount": {
			addr: myOperatorAddr,
			setup: func(ctx sdk.Context) {
				err := contract.BondDelegation(ctx, stakingContractAddr, myOperatorAddr, sdk.NewCoins(sdk.NewCoin("utgd", sdk.OneInt())), contractKeeper)
				require.NoError(t, err)
				err = contract.UnbondDelegation(ctx, stakingContractAddr, myOperatorAddr, sdk.NewCoin("utgd", sdk.OneInt()), contractKeeper)
				require.NoError(t, err)
			},
			expAmount: nil,
		},
		"unknown address": {
			addr:      rand.Bytes(address.Len),
			setup:     func(ctx sdk.Context) {},
			expAmount: nil,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			tCtx, _ := ctx.CacheContext()
			spec.setup(tCtx)
			// when
			gotAmount, gotErr := contractAdapter.QueryStakedAmount(tCtx, spec.addr)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expAmount, gotAmount, "exp %s but got %s", spec.expAmount, gotAmount)
		})
	}
}

func TestQueryValidatorUnboding(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)

	op1Addr, err := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	require.NoError(t, err)

	// unbond some tokens for operator 1
	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now).WithBlockHeight(12)
	unbondedAmount := sdk.NewInt(10)
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	err = contract.UnbondDelegation(ctx, contractAddr, op1Addr, sdk.NewCoin("utgd", unbondedAmount), example.TWasmKeeper.GetContractKeeper())
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
			srcOpAddr: rand.Bytes(address.Len),
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
