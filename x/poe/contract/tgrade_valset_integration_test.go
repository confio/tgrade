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
	ctx, example, vals, _ := setupPoEContracts(t)
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
	// Setup contracts and seed some data. Creates three random validators
	ctx, example, expValidators, _ := setupPoEContracts(t)
	expValidators = clearTokenAmount(expValidators)
	sort.Slice(expValidators, func(i, j int) bool {
		return expValidators[i].OperatorAddress < expValidators[j].OperatorAddress
	})

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)

	specs := map[string]struct {
		pagination *types.Paginator
		expVal     []stakingtypes.Validator
		expEmpty   bool
		expError   bool
	}{
		"query no pagination": {
			pagination: nil,
			expVal:     expValidators,
		},
		"query offset 0, limit 2": {
			pagination: &types.Paginator{Limit: 2},
			expVal:     expValidators[:2],
		},
		"query offset 2, limit 2": {
			pagination: &types.Paginator{StartAfter: []byte(expValidators[1].OperatorAddress), Limit: 2},
			expVal:     expValidators[2:],
		},
		"query offset 3, limit 2": {
			pagination: &types.Paginator{StartAfter: []byte(expValidators[2].OperatorAddress), Limit: 2},
			expEmpty:   true,
		},
		"query offset invalid addr, limit 2": {
			pagination: &types.Paginator{StartAfter: []byte("invalid"), Limit: 2},
			expError:   true,
		},
		// TODO: query offset (valid) unknown addr
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotValidators, err := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil).ListValidators(ctx, spec.pagination)

			// then
			if spec.expError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if spec.expEmpty {
					assert.Equal(t, 0, len(gotValidators))
				} else {
					assert.Equal(t, spec.expVal, gotValidators)
				}
			}
		})
	}
}

func TestQueryValsetConfig(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, _, _ := setupPoEContracts(t)
	mixerContractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeMixer)
	require.NoError(t, err)
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)
	engagementAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	require.NoError(t, err)
	communityPoolAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeCommunityPool)
	require.NoError(t, err)
	distributionAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	require.NoError(t, err)

	// when
	adapter := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil)
	res, gotErr := adapter.QueryConfig(ctx)

	// then
	require.NoError(t, gotErr)

	expConfig := &contract.ValsetConfigResponse{
		Membership:      mixerContractAddr.String(),
		MinWeight:       1,
		MaxValidators:   100,
		Scaling:         1,
		EpochReward:     sdk.NewInt64Coin("utgd", 100000),
		FeePercentage:   sdk.MustNewDecFromStr("0.50"),
		RewardsContract: distributionAddr.String(),
		AutoUnjail:      false,
		DistributionContracts: []contract.DistributionContract{
			{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
			{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
		},
	}
	assert.Equal(t, expConfig, res)
}
