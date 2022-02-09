package keeper

import (
	"context"
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper/poetesting"
)

func TestDelegatorValidators(t *testing.T) {
	var myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)

	specs := map[string]struct {
		src    *distributiontypes.QueryDelegatorValidatorsRequest
		mock   poetesting.ValsetContractMock
		exp    *distributiontypes.QueryDelegatorValidatorsResponse
		expErr bool
	}{
		"delegation": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
			mock: poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				return &stakingtypes.Validator{OperatorAddress: opAddr.String()}, nil
			}},
			exp: &distributiontypes.QueryDelegatorValidatorsResponse{Validators: []string{myOperatorAddr.String()}},
		},
		"unknown": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
			mock: poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				return nil, nil
			}},
			exp: &distributiontypes.QueryDelegatorValidatorsResponse{Validators: []string{}},
		},
		"error": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
			mock: poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper := PoEKeeperMock{ValsetContractFn: func(ctx sdk.Context) ValsetContract { return spec.mock }}

			// when
			q := NewLegacyDistributionGRPCQuerier(poeKeeper)
			gotRes, gotErr := q.DelegatorValidators(ctx, spec.src)

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

func TestDelegatorWithdrawAddress(t *testing.T) {
	// myDelegatedAddr has myWithdrawAddr as withdraw address
	var myDelegatedAddr sdk.AccAddress = rand.Bytes(address.Len)
	var myWithdrawAddr sdk.AccAddress = rand.Bytes(address.Len)
	// myUndelegatedAddr has no withdrawn address
	var myUndelegatedAddr sdk.AccAddress = rand.Bytes(address.Len)

	specs := map[string]struct {
		src    *distributiontypes.QueryDelegatorWithdrawAddressRequest
		mock   poetesting.EngagementContractMock
		exp    *distributiontypes.QueryDelegatorWithdrawAddressResponse
		expErr bool
	}{
		"with withdraw address": {
			src: &distributiontypes.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: myDelegatedAddr.String()},
			mock: poetesting.EngagementContractMock{QueryDelegatedFn: func(ctx sdk.Context, ownerAddr sdk.AccAddress) (*contract.DelegatedResponse, error) {
				return &contract.DelegatedResponse{Delegated: myWithdrawAddr.String()}, nil
			}},
			exp: &distributiontypes.QueryDelegatorWithdrawAddressResponse{WithdrawAddress: myWithdrawAddr.String()},
		},
		"without withdraw address": {
			src: &distributiontypes.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: myUndelegatedAddr.String()},
			mock: poetesting.EngagementContractMock{QueryDelegatedFn: func(ctx sdk.Context, ownerAddr sdk.AccAddress) (*contract.DelegatedResponse, error) {
				return &contract.DelegatedResponse{Delegated: ownerAddr.String()}, nil
			}},
			exp: &distributiontypes.QueryDelegatorWithdrawAddressResponse{WithdrawAddress: myUndelegatedAddr.String()},
		},
		"error": {
			src: &distributiontypes.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: "invalid address"},
			mock: poetesting.EngagementContractMock{QueryDelegatedFn: func(ctx sdk.Context, ownerAddr sdk.AccAddress) (*contract.DelegatedResponse, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper := PoEKeeperMock{EngagementContractFn: func(ctx sdk.Context) EngagementContract { return spec.mock }}

			// when
			q := NewLegacyDistributionGRPCQuerier(poeKeeper)
			gotRes, gotErr := q.DelegatorWithdrawAddress(ctx, spec.src)

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
