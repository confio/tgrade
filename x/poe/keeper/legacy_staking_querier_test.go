package keeper

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper/poetesting"
	"github.com/confio/tgrade/x/poe/types"
)

func TestStakingValidatorDelegations(t *testing.T) {
	var myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)

	poeKeeper := PoEKeeperMock{
		GetBondDenomFn: func(ctx sdk.Context) string { return "utgd" },
	}

	specs := map[string]struct {
		src    *stakingtypes.QueryValidatorDelegationsRequest
		mock   poetesting.StakeContractMock
		exp    *stakingtypes.QueryValidatorDelegationsResponse
		expErr bool
	}{
		"delegation": {
			src: &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
				amount := sdk.NewInt(10)
				return &amount, nil
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
			mock: poetesting.StakeContractMock{QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
				return nil, nil
			}},
			exp: &stakingtypes.QueryValidatorDelegationsResponse{},
		},
		"error": {
			src: &stakingtypes.QueryValidatorDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper.StakeContractFn = func(ctx sdk.Context) StakeContract {
				return spec.mock
			}

			// when
			q := NewLegacyStakingGRPCQuerier(poeKeeper)
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
	var (
		myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)
		myTime                        = time.Now().UTC()
		myHeight       int64          = 123
	)

	poeKeeper := PoEKeeperMock{}
	specs := map[string]struct {
		src    *stakingtypes.QueryValidatorUnbondingDelegationsRequest
		mock   poetesting.StakeContractMock
		exp    *stakingtypes.QueryValidatorUnbondingDelegationsResponse
		expErr bool
	}{
		"one delegation": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
				}, nil
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
						},
					},
				},
			},
		},
		"multiple delegations": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
					{CompletionTime: myTime.Add(time.Minute), Balance: sdk.NewInt(11), InitialBalance: sdk.NewInt(11), CreationHeight: myHeight + 1},
				}, nil
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Entries: []stakingtypes.UnbondingDelegationEntry{
							{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
							{CompletionTime: myTime.Add(time.Minute), Balance: sdk.NewInt(11), InitialBalance: sdk.NewInt(11), CreationHeight: myHeight + 1},
						},
					},
				},
			},
		},
		"none": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{}, nil
			}},
			exp: &stakingtypes.QueryValidatorUnbondingDelegationsResponse{
				UnbondingResponses: []stakingtypes.UnbondingDelegation{
					{
						DelegatorAddress: myOperatorAddr.String(),
						ValidatorAddress: myOperatorAddr.String(),
						Entries:          []stakingtypes.UnbondingDelegationEntry{},
					},
				},
			},
		},
		"error": {
			src: &stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper.StakeContractFn = func(ctx sdk.Context) StakeContract {
				return spec.mock
			}

			// when
			q := NewLegacyStakingGRPCQuerier(poeKeeper)
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
	var myStakingContract sdk.AccAddress = rand.Bytes(address.Len)

	poeKeeper := PoEKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeValset, ctype)
			return myStakingContract, nil
		},
		GetBondDenomFn:      func(ctx sdk.Context) string { return "utgd" },
		HistoricalEntriesFn: func(ctx sdk.Context) uint32 { return 1 },
		StakeContractFn: func(ctx sdk.Context) StakeContract {
			return poetesting.StakeContractMock{QueryStakingUnbondingPeriodFn: func(ctx sdk.Context) (time.Duration, error) {
				return time.Hour, nil
			}}
		},
		ValsetContractFn: func(ctx sdk.Context) ValsetContract {
			return poetesting.ValsetContractMock{QueryConfigFn: func(ctx sdk.Context) (*contract.ValsetConfigResponse, error) {
				return &contract.ValsetConfigResponse{MaxValidators: 2}, nil
			}}
		},
	}
	q := NewLegacyStakingGRPCQuerier(poeKeeper)
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
