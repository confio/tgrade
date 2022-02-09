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

func TestSetEngagementPoints(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, _, _ := setupPoEContracts(t)

	myOperatorAddr := rand.Bytes(address.Len)
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

func TestBondDelegation(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)

	myOperatorAddr, _ := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	// fund account
	example.Faucet.Fund(ctx, myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt()))
	stakingContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)

	// when
	err = contract.BondDelegation(ctx, stakingContractAddr, myOperatorAddr, sdk.NewCoins(sdk.NewCoin("utgd", sdk.OneInt())), example.TWasmKeeper.GetContractKeeper())

	// then
	require.NoError(t, err)

	gotRes, err := contract.QueryStakedAmount(ctx, example.TWasmKeeper, stakingContractAddr, myOperatorAddr)
	require.NoError(t, err)
	expAmount := vals[0].Tokens.Add(sdk.OneInt())
	assert.Equal(t, gotRes.Stake.Amount, expAmount.String())
}

func TestUnbondDelegation(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)

	myOperatorAddr, _ := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	stakingContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)

	// when
	err = contract.UnbondDelegation(ctx, stakingContractAddr, myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt()), example.TWasmKeeper.GetContractKeeper())

	// then
	require.NoError(t, err)

	gotRes, err := contract.QueryStakedAmount(ctx, example.TWasmKeeper, stakingContractAddr, myOperatorAddr)
	require.NoError(t, err)
	expAmount := vals[0].Tokens.Sub(sdk.OneInt())
	assert.Equal(t, gotRes.Stake.Amount, expAmount.String())
}
