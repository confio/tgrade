package contract_test

import (
	"github.com/confio/tgrade/x/poe/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"testing"
	"time"

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
	example.Faucet.Fund(ctx, myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt()))

	convertToFullVestingAccount(t, example, ctx, myOperatorAddr)
	// and one liquid token
	example.Faucet.Fund(ctx, myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt()))

	stakingContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	specs := map[string]struct {
		liquid  sdk.Coins
		vesting *sdk.Coin
		expErr  bool
	}{
		"liquid only": {
			liquid: sdk.NewCoins(sdk.NewCoin("utgd", sdk.OneInt())),
		},
		"insufficient liquid tokens": {
			liquid: sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(2))),
			expErr: true,
		},
		// todo: "vesting only": { // rejected by contract currently
		//	vesting: &sdk.Coin{Denom: "utgd", Amount: sdk.OneInt()},
		//},
		// todo: insufficient vesting tokens
		"both": {
			liquid:  sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(1))),
			vesting: &sdk.Coin{Denom: "utgd", Amount: sdk.OneInt()},
		},
	}
	parentCtx := ctx
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ = parentCtx.CacheContext()
			// when
			gotErr := contract.BondDelegation(ctx, stakingContractAddr, myOperatorAddr, spec.liquid, spec.vesting, example.TWasmKeeper.GetContractKeeper())

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			gotRes, err := contract.QueryStakedAmount(ctx, example.TWasmKeeper, stakingContractAddr, myOperatorAddr)
			require.NoError(t, err)
			expAmount := vals[0].Tokens.Add(sdk.OneInt())
			assert.Equal(t, gotRes.Liquid.Amount, expAmount.String())
		})
	}

}

func TestUnbondDelegation(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)
	stakingContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)

	myOperatorAddr, _ := sdk.AccAddressFromBech32(vals[0].OperatorAddress)
	vestingAmount := sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt())
	example.Faucet.Fund(ctx, myOperatorAddr, vestingAmount)
	convertToFullVestingAccount(t, example, ctx, myOperatorAddr)

	liquidAmount := sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt())
	example.Faucet.Fund(ctx, myOperatorAddr, liquidAmount)
	someLiquidAmount := sdk.NewCoins(liquidAmount) // todo: should be empty when fixed in PoE contract
	// and bond from vesting amount
	err = contract.BondDelegation(ctx, stakingContractAddr, myOperatorAddr, someLiquidAmount, &sdk.Coin{Denom: types.DefaultBondDenom, Amount: sdk.OneInt()}, example.TWasmKeeper.GetContractKeeper())
	require.NoError(t, err)
	unbodingPeriod, err := example.PoEKeeper.StakeContract(ctx).QueryStakingUnbondingPeriod(ctx)
	require.NoError(t, err)

	parentCtx := ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	specs := map[string]struct {
		unbondAmount   sdk.Coin
		expLiquidBond  sdk.Int
		expVestingBond sdk.Int
		expErr         bool
	}{
		"liquid": {
			unbondAmount:   sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt()),
			expLiquidBond:  sdk.OneInt(),
			expVestingBond: sdk.OneInt(),
		},
		"liquid + vesting": {
			unbondAmount:   sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(3)),
			expLiquidBond:  sdk.ZeroInt(),
			expVestingBond: sdk.ZeroInt(),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _ = parentCtx.CacheContext()
			// when
			completionTime, err := contract.UnbondDelegation(ctx, stakingContractAddr, myOperatorAddr, spec.unbondAmount, example.TWasmKeeper.GetContractKeeper())

			// then
			require.NoError(t, err)
			assert.Equal(t, ctx.BlockTime().Add(unbodingPeriod).UTC(), *completionTime)

			gotRes, err := contract.QueryStakedAmount(ctx, example.TWasmKeeper, stakingContractAddr, myOperatorAddr)
			require.NoError(t, err)
			// amount staked during genTx + bond liquid + bond vesting - unbond amount
			expAmount := vals[0].Tokens.Add(liquidAmount.Amount).Sub(spec.expLiquidBond)
			assert.Equal(t, expAmount.String(), gotRes.Liquid.Amount)
			assert.Equal(t, spec.expVestingBond.String(), gotRes.Vesting.Amount)
			// todo: claim and check if added to balance
			//a:=contract.NewStakeContractAdapter(stakingContractAddr, example.TWasmKeeper, nil)
			//a.Claim()
		})
	}
}

// make vesting account with current balance as vested amount
func convertToFullVestingAccount(t *testing.T, example keeper.TestKeepers, ctx sdk.Context, addr sdk.AccAddress) {
	vestingtypes.RegisterInterfaces(example.EncodingConfig.InterfaceRegistry)
	acc := example.AccountKeeper.GetAccount(ctx, addr)
	require.NotNil(t, acc)
	bAcc, ok := acc.(*authtypes.BaseAccount)
	require.True(t, ok)
	balance := example.BankKeeper.GetBalance(ctx, addr, types.DefaultBondDenom)
	// setup vesting account with old balance into vesting
	vAcct := vestingtypes.NewDelayedVestingAccount(bAcc, sdk.NewCoins(balance), time.Now().Add(time.Hour).UnixNano())
	example.AccountKeeper.SetAccount(ctx, vAcct)
}
