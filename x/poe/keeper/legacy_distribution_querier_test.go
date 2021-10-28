package keeper

import (
	"context"
	"errors"
	"testing"

	"github.com/confio/tgrade/x/poe/keeper/poetesting"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
)

func TestDelegatorValidators(t *testing.T) {
	var myOperatorAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)

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
