package contract_test

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryValidator(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals := setupPoEContracts(t)
	vals = clearTokenAmount(vals)

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
	expValidators = clearTokenAmount(expValidators)

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
		DistributionContract:  "cosmos14hj2tavq8fpesdwxxcu44rty3hh90vhuc53mp6",
		RewardsContract:       "cosmos1cnuw3f076wgdyahssdkd0g3nr96ckq8caf5mdm",
		AutoUnjail:            false,
	}
	assert.Equal(t, expConfig, res)
}
