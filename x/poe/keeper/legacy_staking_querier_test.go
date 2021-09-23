package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
	"time"
)

func TestStakingValidatorDelegations(t *testing.T) {
	var myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myOperatorAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	contractSource := PoEKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeStaking, ctype)
			return myStakingContract, nil
		},
		GetBondDenomFn: func(ctx sdk.Context) string { return "utgd" },
	}

	specs := map[string]struct {
		src     *stakingtypes.QueryValidatorDelegationsRequest
		querier types.SmartQuerier
		exp     *stakingtypes.QueryValidatorDelegationsResponse
		expErr  bool
	}{
		"delegation": {
			src: &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				var amount = 10
				return json.Marshal(contract.TG4MemberResponse{
					Weight: &amount,
				})
			}},
			exp: &stakingtypes.QueryValidatorDelegationsResponse{DelegationResponses: stakingtypes.DelegationResponses{
				{
					Delegation: stakingtypes.Delegation{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Shares:           sdk.OneDec(),
					},
					Balance: sdk.NewCoin("utgd", sdk.NewInt(10)),
				},
			}},
		},
		"empty": {
			src: &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.TG4MemberResponse{})
			}},
			exp: &stakingtypes.QueryValidatorDelegationsResponse{},
		},
		"error": {
			src: &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))

			// when
			q := NewLegacyStakingGRPCQuerier(contractSource, spec.querier)
			gotRes, gotErr := q.ValidatorDelegations(ctx, spec.src)

			// then
			if spec.expErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, gotRes)

		})
	}
}

func TestStakingValidatorUnbondingDelegations(t *testing.T) {
	var myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myOperatorAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	contractSource := PoEKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeStaking, ctype)
			return myStakingContract, nil
		},
		GetBondDenomFn: func(ctx sdk.Context) string { return "utgd" },
	}
	anyTime := time.Now().UTC()
	specs := map[string]struct {
		src     *stakingtypes.QueryValidatorUnbondingDelegationsRequest
		querier types.SmartQuerier
		exp     *stakingtypes.QueryValidatorUnbondingDelegationsResponse
		expErr  bool
	}{
		"one delegation": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {

				return json.Marshal(contract.TG4StakeClaimsResponse{
					Claims: []contract.TG4StakeClaim{{
						Amount:    sdk.NewInt(10),
						ReleaseAt: uint64(anyTime.UnixNano()),
					},
					},
				})
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{CompletionTime: anyTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10)},
						},
					},
				}},
		},
		"multiple delegations": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {

				return json.Marshal(contract.TG4StakeClaimsResponse{
					Claims: []contract.TG4StakeClaim{{
						Amount:    sdk.NewInt(10),
						ReleaseAt: uint64(anyTime.UnixNano()),
					}, {
						Amount:    sdk.NewInt(11),
						ReleaseAt: uint64(anyTime.Add(time.Minute).UnixNano()),
					},
					},
				})
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{CompletionTime: anyTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10)},
							{CompletionTime: anyTime.Add(time.Minute), Balance: sdk.NewInt(11), InitialBalance: sdk.NewInt(11)},
						},
					},
				}},
		},
		"none": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.TG4StakeClaimsResponse{})
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
					},
				}},
		},
		"error": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))

			// when
			q := NewLegacyStakingGRPCQuerier(contractSource, spec.querier)
			gotRes, gotErr := q.ValidatorUnbondingDelegations(ctx, spec.src)

			// then
			if spec.expErr {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, gotRes)

		})
	}
}

func TestStakingParams(t *testing.T) {
	var myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	keeperMock := PoEKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeValset, ctype)
			return myStakingContract, nil
		},
		GetBondDenomFn:  func(ctx sdk.Context) string { return "utgd" },
		UnbondingTimeFn: func(ctx sdk.Context) time.Duration { return time.Hour },
		HistoricalEntriesFn: func(ctx sdk.Context) uint32 {
			return 1
		},
	}
	smartQuerier := SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		return json.Marshal(contract.ValsetConfigResponse{
			MaxValidators: 2,
		})
	}}
	q := NewLegacyStakingGRPCQuerier(keeperMock, smartQuerier)
	ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
	gotRes, gotErr := q.Params(ctx, &stakingtypes.QueryParamsRequest{})
	require.NoError(t, gotErr)
	exp := &stakingtypes.QueryParamsResponse{
		Params: stakingtypes.Params{
			UnbondingTime:     time.Hour,
			MaxValidators:     2,
			MaxEntries:        0,
			HistoricalEntries: 1,
			BondDenom:         "utgd",
		},
	}
	assert.Equal(t, exp, gotRes)

}
