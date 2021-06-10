package keeper

import (
	"context"
	"errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/types"
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
			q := NewGrpcQuerier(ContractSourceMock{GetPoEContractAddressFn: spec.mockFn})
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
