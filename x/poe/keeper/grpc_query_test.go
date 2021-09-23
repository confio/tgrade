package keeper

import (
	"context"
	"encoding/json"
	"errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestQueryContractAddress(t *testing.T) {
	var myContractAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
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
			q := NewGrpcQuerier(PoEKeeperMock{GetPoEContractAddressFn: spec.mockFn}, nil)
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
	var myValsetContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	contractSource := newContractSourceMock(t, myValsetContract, nil)

	pubKey := ed25519.GenPrivKey().PubKey()
	expValidator := types.ValidatorFixture(func(m *stakingtypes.Validator) {
		pkAny, _ := codectypes.NewAnyWithValue(pubKey)
		m.ConsensusPubkey = pkAny
	})

	querier := SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		return json.Marshal(contract.ListValidatorsResponse{Validators: []contract.OperatorResponse{
			{
				Operator: expValidator.OperatorAddress,
				Pubkey:   contract.ValidatorPubkey{Ed25519: pubKey.Bytes()},
				Metadata: contract.MetadataFromDescription(expValidator.Description),
			},
		}})
	}}
	specs := map[string]struct {
		src     *stakingtypes.QueryValidatorsRequest
		querier types.SmartQuerier
		exp     *stakingtypes.QueryValidatorsResponse
		expErr  bool
	}{
		"all good": {
			src:     &stakingtypes.QueryValidatorsRequest{},
			querier: querier,
			exp: &stakingtypes.QueryValidatorsResponse{
				Validators: []stakingtypes.Validator{expValidator},
			},
		},
		"nil request": {
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		"empty result": {
			src: &stakingtypes.QueryValidatorsRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				r := contract.ListValidatorsResponse{}
				return json.Marshal(r)
			}},
			exp: &stakingtypes.QueryValidatorsResponse{
				Validators: []stakingtypes.Validator{},
			},
		},
		"nil result": {
			src: &stakingtypes.QueryValidatorsRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns error": {
			src: &stakingtypes.QueryValidatorsRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))

			// when
			s := NewGrpcQuerier(contractSource, spec.querier)
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
	var myValsetContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myOperator sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	contractSource := newContractSourceMock(t, myValsetContract, nil)

	pubKey := ed25519.GenPrivKey().PubKey()
	expValidator := types.ValidatorFixture(func(m *stakingtypes.Validator) {
		pkAny, _ := codectypes.NewAnyWithValue(pubKey)
		m.ConsensusPubkey = pkAny
	})

	querier := SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		return json.Marshal(contract.ValidatorResponse{Validator: &contract.OperatorResponse{
			Operator: expValidator.OperatorAddress,
			Pubkey:   contract.ValidatorPubkey{Ed25519: pubKey.Bytes()},
			Metadata: contract.MetadataFromDescription(expValidator.Description),
		}})
	}}
	specs := map[string]struct {
		src     *stakingtypes.QueryValidatorRequest
		querier types.SmartQuerier
		exp     *stakingtypes.QueryValidatorResponse
		expErr  bool
	}{
		"all good": {
			src:     &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			querier: querier,
			exp: &stakingtypes.QueryValidatorResponse{
				Validator: expValidator,
			},
		},
		"nil request": {
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		"empty address": {
			src: &stakingtypes.QueryValidatorRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		"not found": {
			src: &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns error": {
			src: &stakingtypes.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))

			// when
			s := NewGrpcQuerier(contractSource, spec.querier)
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
	var myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	contractSource := newContractSourceMock(t, nil, myStakingContract)

	specs := map[string]struct {
		src     *types.QueryUnbondingPeriodRequest
		querier types.SmartQuerier
		exp     *types.QueryUnbondingPeriodResponse
		expErr  bool
	}{
		"all good": {
			src: &types.QueryUnbondingPeriodRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.UnbondingPeriodResponse{
					UnbondingPeriod: 1,
				})
			}},
			exp: &types.QueryUnbondingPeriodResponse{
				Time: time.Second,
			},
		},
		"nil request": {
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns nil": {
			src: &types.QueryUnbondingPeriodRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns error": {
			src: &types.QueryUnbondingPeriodRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()))

			// when
			s := NewGrpcQuerier(contractSource, spec.querier)
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
		src     *types.QueryValidatorDelegationRequest
		querier types.SmartQuerier
		exp     *types.QueryValidatorDelegationResponse
		expErr  bool
	}{
		"delegation": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				var amount = 10
				return json.Marshal(contract.TG4MemberResponse{
					Weight: &amount,
				})
			}},
			exp: &types.QueryValidatorDelegationResponse{
				Balance: sdk.NewCoin("utgd", sdk.NewInt(10)),
			},
		},
		"empty": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.TG4MemberResponse{})
			}},
			expErr: true,
		},
		"error": {
			src: &types.QueryValidatorDelegationRequest{ValidatorAddr: myOperatorAddr.String()},
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
			q := NewGrpcQuerier(contractSource, spec.querier)
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
		myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
		myOperatorAddr    sdk.AccAddress = rand.Bytes(sdk.AddrLen)
		myTime                           = time.Now().UTC()
		myHeight          int64          = 123
	)

	contractSource := PoEKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeStaking, ctype)
			return myStakingContract, nil
		},
		GetBondDenomFn: func(ctx sdk.Context) string { return "utgd" },
	}
	specs := map[string]struct {
		src     *types.QueryValidatorUnbondingDelegationsRequest
		querier types.SmartQuerier
		exp     *types.QueryValidatorUnbondingDelegationsResponse
		expErr  bool
	}{
		"one delegation": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {

				return json.Marshal(contract.TG4StakeClaimsResponse{
					Claims: []contract.TG4StakeClaim{{
						Amount:         sdk.NewInt(10),
						ReleaseAt:      uint64(myTime.UnixNano()),
						CreationHeight: uint64(myHeight),
					},
					},
				})
			}},
			exp: &types.QueryValidatorUnbondingDelegationsResponse{
				Entries: []stakingtypes.UnbondingDelegationEntry{
					{CompletionTime: myTime, Balance: sdk.NewInt(10), InitialBalance: sdk.NewInt(10), CreationHeight: myHeight},
				},
			},
		},
		"multiple delegations": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {

				return json.Marshal(contract.TG4StakeClaimsResponse{
					Claims: []contract.TG4StakeClaim{{
						Amount:         sdk.NewInt(10),
						ReleaseAt:      uint64(myTime.UnixNano()),
						CreationHeight: uint64(myHeight),
					}, {
						Amount:         sdk.NewInt(11),
						ReleaseAt:      uint64(myTime.Add(time.Minute).UnixNano()),
						CreationHeight: uint64(myHeight + 1),
					},
					},
				})
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
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.TG4StakeClaimsResponse{})
			}},
			exp: &types.QueryValidatorUnbondingDelegationsResponse{},
		},
		"error": {
			src: &types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myOperatorAddr.String()},
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
			q := NewGrpcQuerier(contractSource, spec.querier)
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
