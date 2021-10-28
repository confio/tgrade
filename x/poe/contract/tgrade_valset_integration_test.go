package contract_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryValidator(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)
	vals = resetTokenAmount(vals)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)

	specs := map[string]struct {
		operatorAddr string
		expVal       stakingtypes.Validator
		expEmpty     bool
	}{
		"query one validator": {
			operatorAddr: vals[0].OperatorAddress,
			expVal:       vals[0],
		},
		"query other validator": {
			operatorAddr: vals[1].OperatorAddress,
			expVal:       vals[1],
		},
		"query with unknown address": {
			operatorAddr: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
			expEmpty:     true,
		},
		"query with invalid address": {
			operatorAddr: "not an address",
			expEmpty:     true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			opAddr, _ := sdk.AccAddressFromBech32(spec.operatorAddr)

			// when
			adaptor := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil)
			gotVal, err := adaptor.QueryValidator(ctx, opAddr)

			// then
			if spec.expEmpty {
				assert.Nil(t, gotVal)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.expVal, *gotVal)
		})
	}
}

func TestListValidators(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, expValidators := setupPoEContracts(t)
	expValidators = resetTokenAmount(expValidators)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)

	// when
	gotValidators, err := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil).ListValidators(ctx)

	// then
	require.NoError(t, err)
	sort.Slice(expValidators, func(i, j int) bool {
		return expValidators[i].OperatorAddress < expValidators[j].OperatorAddress
	})
	assert.Equal(t, expValidators, gotValidators)
}

func TestQueryValsetConfig(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, _ := setupPoEContracts(t)
	mixerContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeMixer)
	require.NoError(t, err)
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)

	// when
	adapter := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil)
	res, gotErr := adapter.QueryConfig(ctx)

	// then
	require.NoError(t, gotErr)

	expConfig := &contract.ValsetConfigResponse{
		Membership:            mixerContractAddr.String(),
		MinWeight:             1,
		MaxValidators:         100,
		Scaling:               1,
		EpochReward:           sdk.NewInt64Coin("utgd", 100000),
		FeePercentage:         sdk.MustNewDecFromStr("0.50"),
		ValidatorsRewardRatio: sdk.MustNewDecFromStr("0.50"),
		DistributionContract:  "cosmos1nc5tatafv6eyq7llkr2gv50ff9e22mnfapsq9f",
		RewardsContract:       "cosmos18v47nqmhvejx3vc498pantg8vr435xa0ln6420",
		AutoUnjail:            false,
	}
	assert.Equal(t, expConfig, res)
}

func setupPoEContracts(t *testing.T) (sdk.Context, keeper.TestKeepers, []stakingtypes.Validator) {
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, expValidators := withRandomValidators(t, ctx, example, 3)
	gs := types.GenesisStateFixture(mutator)

	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)
	return ctx, example, expValidators
}