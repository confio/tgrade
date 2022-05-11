package keeper

import (
	"context"
	"errors"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper/poetesting"
	"github.com/confio/tgrade/x/poe/types"
)

func TestQueryContractAddress(t *testing.T) {
	var myContractAddr sdk.AccAddress = rand.Bytes(address.Len)
	specs := map[string]struct {
		srcMsg     types.QueryContractAddressRequest
		mockFn     func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
		expResult  *types.QueryContractAddressResponse
		expErrCode codes.Code
	}{
		"return address": {
			srcMsg: types.QueryContractAddressRequest{ContractType: types.PoEContractTypeMixer},
			mockFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
				return myContractAddr, nil
			},
			expResult: &types.QueryContractAddressResponse{
				Address: myContractAddr.String(),
			},
		},
		"not found": {
			srcMsg: types.QueryContractAddressRequest{ContractType: types.PoEContractTypeMixer},
			mockFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
				return nil, wasmtypes.ErrNotFound
			},
			expErrCode: codes.NotFound,
		},
		"other error": {
			srcMsg: types.QueryContractAddressRequest{ContractType: types.PoEContractTypeMixer},
			mockFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
				return nil, errors.New("testing")
			},
			expErrCode: codes.Internal,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			q := NewGrpcQuerier(PoEKeeperMock{GetPoEContractAddressFn: spec.mockFn})
			ctx := sdk.Context{}.WithContext(context.Background())
			gotRes, gotErr := q.ContractAddress(sdk.WrapSDKContext(ctx), &spec.srcMsg)
			if spec.expErrCode != 0 {
				require.Error(t, gotErr)
				assert.Equal(t, spec.expErrCode, status.Code(gotErr))
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expResult, gotRes)
		})
	}
}

func TestQueryValidators(t *testing.T) {
	var myValsetContract sdk.AccAddress = rand.Bytes(address.Len)
	poeKeeper := newContractSourceMock(t, myValsetContract, nil)

	pubKey := ed25519.GenPrivKey().PubKey()
	expValidator := types.ValidatorFixture(func(m *stakingtypes.Validator) {
		pkAny, _ := codectypes.NewAnyWithValue(pubKey)
		m.ConsensusPubkey = pkAny
	})

	specs := map[string]struct {
		src    *stakingtypes.QueryValidatorsRequest
		mock   poetesting.ValsetContractMock
		exp    *stakingtypes.QueryValidatorsResponse
		expErr bool
	}{
		"all good": {
			src: &stakingtypes.QueryValidatorsRequest{},
			mock: poetesting.ValsetContractMock{
				ListValidatorsFn: func(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error) {
					return []stakingtypes.Validator{expValidator}, []byte("my_next_key"), nil
				},
			},
			exp: &stakingtypes.QueryValidatorsResponse{
				Validators: []stakingtypes.Validator{expValidator},
				Pagination: &query.PageResponse{
					NextKey: []byte("my_next_key"),
				},
			},
		},
		"all good - without pagination cursor": {
			src: &stakingtypes.QueryValidatorsRequest{},
			mock: poetesting.ValsetContractMock{
				ListValidatorsFn: func(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error) {
					return []stakingtypes.Validator{expValidator}, []byte{}, nil
				},
			},
			exp: &stakingtypes.QueryValidatorsResponse{
				Validators: []stakingtypes.Validator{expValidator},
			},
		},
		"nil request": {
			mock:   poetesting.ValsetContractMock{},
			expErr: true,
		},
		"empty result": {
			src: &stakingtypes.QueryValidatorsRequest{},
			mock: poetesting.ValsetContractMock{
				ListValidatorsFn: func(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error) {
					return []stakingtypes.Validator{}, nil, nil
				},
			},
			exp: &stakingtypes.QueryValidatorsResponse{
				Validators: []stakingtypes.Validator{},
			},
		},
		"contract returns error": {
			src: &stakingtypes.QueryValidatorsRequest{},
			mock: poetesting.ValsetContractMock{
				ListValidatorsFn: func(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error) {
					return nil, nil, errors.New("testing")
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper.ValsetContractFn = func(ctx sdk.Context) ValsetContract { return spec.mock }
			// when
			s := NewGrpcQuerier(poeKeeper)
			gotRes, gotErr := s.Validators(ctx, spec.src)

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

func TestQueryValidator(t *testing.T) {
	var myOperator sdk.AccAddress = rand.Bytes(address.Len)

	pubKey := ed25519.GenPrivKey().PubKey()
	expValidator := types.ValidatorFixture(func(m *stakingtypes.Validator) {
		pkAny, _ := codectypes.NewAnyWithValue(pubKey)
		m.ConsensusPubkey = pkAny
	})

	specs := map[string]struct {
		src    *stakingtypes.QueryValidatorRequest
		mock   poetesting.ValsetContractMock
		exp    *stakingtypes.QueryValidatorResponse
		expErr bool
	}{
		"all good": {
			src: &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			mock: poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				return &expValidator, nil
			}},
			exp: &stakingtypes.QueryValidatorResponse{
				Validator: expValidator,
			},
		},
		"nil request": {
			mock:   poetesting.ValsetContractMock{},
			expErr: true,
		},
		"empty address": {
			src:    &stakingtypes.QueryValidatorRequest{},
			mock:   poetesting.ValsetContractMock{},
			expErr: true,
		},
		"not found": {
			src: &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			mock: poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns error": {
			src: &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
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
			s := NewGrpcQuerier(poeKeeper)
			gotRes, gotErr := s.Validator(ctx, spec.src)

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

func TestQueryUnbondingPeriod(t *testing.T) {
	specs := map[string]struct {
		src    *types.QueryUnbondingPeriodRequest
		mock   poetesting.StakeContractMock
		exp    *types.QueryUnbondingPeriodResponse
		expErr bool
	}{
		"all good": {
			src: &types.QueryUnbondingPeriodRequest{},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingPeriodFn: func(ctx sdk.Context) (time.Duration, error) {
				return time.Second, nil
			}},
			exp: &types.QueryUnbondingPeriodResponse{
				Time: time.Second,
			},
		},
		"nil request": {
			expErr: true,
		},
		"contract returns error": {
			src: &types.QueryUnbondingPeriodRequest{},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingPeriodFn: func(ctx sdk.Context) (time.Duration, error) {
				return 0, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			poeKeeper := PoEKeeperMock{StakeContractFn: func(ctx sdk.Context) StakeContract {
				return spec.mock
			}}

			// when
			s := NewGrpcQuerier(poeKeeper)
			gotRes, gotErr := s.UnbondingPeriod(ctx, spec.src)

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

func TestValidatorDelegation(t *testing.T) {
	var myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)

	poeKeeper := PoEKeeperMock{
		GetBondDenomFn: func(ctx sdk.Context) string { return "utgd" },
	}

	specs := map[string]struct {
		src    *types.QueryValidatorDelegationRequest
		mock   poetesting.StakeContractMock
		exp    *types.QueryValidatorDelegationResponse
		expErr bool
	}{
		"delegation": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
				amount := sdk.NewInt(10)
				return &amount, nil
			}},
			exp: &types.QueryValidatorDelegationResponse{
				Balance: sdk.NewCoin("utgd", sdk.NewInt(10)),
			},
		},
		"empty": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"error": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
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
			q := NewGrpcQuerier(poeKeeper)
			gotRes, gotErr := q.ValidatorDelegation(ctx, spec.src)

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

func TestValidatorUnbondingDelegations(t *testing.T) {
	var (
		myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)
		myTime                        = time.Now().UTC()
		myHeight       int64          = 123
	)

	poeKeeper := PoEKeeperMock{}
	specs := map[string]struct {
		src    *types.QueryValidatorUnbondingDelegationsRequest
		mock   poetesting.StakeContractMock
		exp    *types.QueryValidatorUnbondingDelegationsResponse
		expErr bool
	}{
		"one delegation": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
				}, nil
			}},
			exp: &types.QueryValidatorUnbondingDelegationsResponse{
				Entries: []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
				},
			},
		},
		"multiple delegations": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
					{CompletionTime: myTime.Add(time.Minute), Balance: sdk.NewInt(11), InitialBalance: sdk.NewInt(11), CreationHeight: myHeight + 1},
				}, nil
			}},
			exp: &types.QueryValidatorUnbondingDelegationsResponse{
				Entries: []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
					{CompletionTime: myTime.Add(time.Minute), Balance: sdk.NewInt(11), InitialBalance: sdk.NewInt(11), CreationHeight: myHeight + 1},
				},
			},
		},
		"none": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			mock: poetesting.StakeContractMock{QueryStakingUnbondingFn: func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
				return []stakingtypes.UnbondingDelegationEntry{}, nil
			}},
			exp: &types.QueryValidatorUnbondingDelegationsResponse{Entries: []stakingtypes.UnbondingDelegationEntry{}},
		},
		"error": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
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
			q := NewGrpcQuerier(poeKeeper)
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

func TestValidatorOutstandingReward(t *testing.T) {
	var anyAddr sdk.AccAddress = rand.Bytes(address.Len)

	specs := map[string]struct {
		src    *types.QueryValidatorOutstandingRewardRequest
		mock   DistributionContract
		exp    *types.QueryValidatorOutstandingRewardResponse
		expErr error
	}{
		"reward": {
			src: &types.QueryValidatorOutstandingRewardRequest{ValidatorAddress: anyAddr.String()},
			mock: poetesting.DistributionContractMock{ValidatorOutstandingRewardFn: func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
				require.Equal(t, anyAddr, addr)
				return sdk.NewCoin("utgd", sdk.OneInt()), nil
			}},
			exp: &types.QueryValidatorOutstandingRewardResponse{
				Reward: sdk.NewDecCoin("utgd", sdk.OneInt()),
			},
		},
		"not found": {
			src: &types.QueryValidatorOutstandingRewardRequest{ValidatorAddress: anyAddr.String()},
			mock: poetesting.DistributionContractMock{ValidatorOutstandingRewardFn: func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
				return sdk.Coin{}, types.ErrNotFound
			}},
			expErr: status.Error(codes.NotFound, "address"),
		},
		"any error": {
			src: &types.QueryValidatorOutstandingRewardRequest{ValidatorAddress: anyAddr.String()},
			mock: poetesting.DistributionContractMock{ValidatorOutstandingRewardFn: func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
				return sdk.Coin{}, errors.New("testing")
			}},
			expErr: status.Error(codes.Internal, "testing"),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			keeperMock := PoEKeeperMock{
				DistributionContractFn: func(ctx sdk.Context) DistributionContract { return spec.mock },
			}

			c := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))
			// when
			s := NewGrpcQuerier(keeperMock)
			gotResp, gotErr := s.ValidatorOutstandingReward(c, spec.src)
			// then
			if spec.expErr != nil {
				assert.ErrorIs(t, gotErr, spec.expErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.exp, gotResp)
		})
	}
}
