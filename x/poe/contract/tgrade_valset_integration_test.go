package contract_test

import (
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/address"

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
			operatorAddr: sdk.AccAddress(rand.Bytes(address.Len)).String(),
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
		pagination *contract.Paginator
		expVal     []stakingtypes.Validator
		expCursor  bool
		expError   bool
	}{
		"query no pagination passed": {
			pagination: nil,
			expVal:     expValidators,
			expCursor:  true,
		},
		"query offset 0, limit 2": {
			pagination: &contract.Paginator{Limit: 2},
			expVal:     expValidators[:2],
			expCursor:  true,
		},
		"query offset 2, limit 2": {
			pagination: &contract.Paginator{StartAfter: []byte(expValidators[1].OperatorAddress), Limit: 2},
			expVal:     expValidators[2:],
			expCursor:  true,
		},
		"query offset 3, limit 2": {
			pagination: &contract.Paginator{StartAfter: []byte(expValidators[2].OperatorAddress), Limit: 2},
			expVal:     []stakingtypes.Validator{},
		},
		"query offset invalid addr, limit 2": {
			pagination: &contract.Paginator{StartAfter: []byte("invalid"), Limit: 2},
			expError:   true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotValidators, gotCursor, err := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil).ListValidators(ctx, spec.pagination)

			// then
			if spec.expError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.expVal, gotValidators)
			assert.Equal(t, spec.expCursor, len(gotCursor) > 0)
		})
	}
}

func TestListAllValidatorsViaCursor(t *testing.T) {
	// Setup contracts and seed some data. Creates three random validators
	ctx, example, expValidators, _ := setupPoEContracts(t)
	require.Len(t, expValidators, 3)

	expValidators = clearTokenAmount(expValidators)
	sort.Slice(expValidators, func(i, j int) bool {
		return expValidators[i].OperatorAddress < expValidators[j].OperatorAddress
	})

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)
	adapter := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil)
	specs := map[string]struct {
		limit uint64
	}{
		"limit 0": {
			limit: 0,
		},
		"limit 1": {
			limit: 1,
		},
		"limit 2": {
			limit: 2,
		},
		"limit 3": {
			limit: 3,
		},
		"limit 4": {
			limit: 4,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			fetchAll := func(limit uint64) (r []stakingtypes.Validator) {
				var cursor contract.PaginationCursor
				for {
					var gotValidatorChunk []stakingtypes.Validator
					gotValidatorChunk, cursor, err = adapter.ListValidators(ctx, &contract.Paginator{StartAfter: cursor, Limit: limit})
					require.NoError(t, err)
					r = append(r, gotValidatorChunk...)
					if len(cursor) == 0 {
						return
					}
				}
			}
			gotValidators := fetchAll(spec.limit)
			require.Equal(t, len(expValidators), len(gotValidators))
			assert.Equal(t, expValidators, gotValidators)
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

func TestJailUnjail(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)
	require.Len(t, vals, 3)

	ocProposeAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeOversightCommunityGovProposals)
	require.NoError(t, err)

	op1Addr, _ := sdk.AccAddressFromBech32(vals[1].OperatorAddress)
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)

	ctx, _ = ctx.CacheContext()
	nextBlockDuration := 1 * time.Second
	specs := map[string]struct {
		jailDuration     time.Duration
		jailForever      bool
		actor            sdk.AccAddress
		expErrUnjail     bool
		expJailingPeriod contract.JailingPeriod
	}{
		"jail lock expired": {
			jailDuration:     nextBlockDuration,
			actor:            op1Addr,
			expJailingPeriod: contract.JailingPeriod{Until: ctx.BlockTime().Add(nextBlockDuration).UTC()},
			expErrUnjail:     true,
		},
		"forever": {
			jailForever:      true,
			actor:            op1Addr,
			expJailingPeriod: contract.JailingPeriod{Forever: true},
			expErrUnjail:     false,
		},
	}
	for name, spec := range specs {
		ctx, _ = ctx.CacheContext()
		t.Run(name, func(t *testing.T) {
			t.Logf("block time: %s", ctx.BlockTime())
			// when
			adapter := contract.NewValsetContractAdapter(contractAddr, example.TWasmKeeper, nil)
			err = adapter.JailValidator(ctx, op1Addr, spec.jailDuration, spec.jailForever, ocProposeAddr)
			require.NoError(t, err)

			// then
			res, err := adapter.QueryRawValidator(ctx, op1Addr)
			require.NoError(t, err)
			require.Equal(t, &spec.expJailingPeriod, res.Validator.JailedUntil)

			// and when
			qCtx := ctx.WithBlockTime(ctx.BlockTime().Add(nextBlockDuration).UTC())
			gotErr := adapter.UnjailValidator(qCtx, spec.actor)
			if !spec.expErrUnjail {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			// then
			res, err = adapter.QueryRawValidator(qCtx, op1Addr)
			require.NoError(t, err)
			assert.Empty(t, res.Validator.JailedUntil)
		})
	}

}
