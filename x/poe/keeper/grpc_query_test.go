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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
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
			q := NewGrpcQuerier(ContractSourceMock{GetPoEContractAddressFn: spec.mockFn}, nil)
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
	expValidator := types.ValidatorFixtureFixture(func(m *types.Validator) {
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
		src     *types.QueryValidatorsRequest
		querier types.SmartQuerier
		exp     *types.QueryValidatorsResponse
		expErr  bool
	}{
		"all good": {
			src:     &types.QueryValidatorsRequest{},
			querier: querier,
			exp: &types.QueryValidatorsResponse{
				Validators: []types.Validator{expValidator},
			},
		},
		"nil request": {
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		//"not found": {
		//	src: &types.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
		//	querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		//		return nil, nil
		//	}},
		//	expErr: true,
		//},
		//"contract returns error": {
		//	src: &types.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
		//	querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		//		return nil, errors.New("testing")
		//	}},
		//	expErr: true,
		//},
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
	expValidator := types.ValidatorFixtureFixture(func(m *types.Validator) {
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
		src     *types.QueryValidatorRequest
		querier types.SmartQuerier
		exp     *types.QueryValidatorResponse
		expErr  bool
	}{
		"all good": {
			src:     &types.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			querier: querier,
			exp: &types.QueryValidatorResponse{
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
			src: &types.QueryValidatorRequest{},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				t.Fatalf("not expected to be called")
				return nil, nil
			}},
			expErr: true,
		},
		"not found": {
			src: &types.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
			querier: SmartQuerierMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, nil
			}},
			expErr: true,
		},
		"contract returns error": {
			src: &types.QueryValidatorRequest{ValidatorAddr: myOperator.String()},
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
