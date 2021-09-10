package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestDelegatorValidators(t *testing.T) {
	var myValsetContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myOperatorAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	contractSource := StakingQuerierKeeperMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
			require.Equal(t, types.PoEContractTypeValset, ctype)
			return myValsetContract, nil
		},
	}

	specs := map[string]struct {
		src     *distributiontypes.QueryDelegatorValidatorsRequest
		querier types.SmartQuerier
		exp     *distributiontypes.QueryDelegatorValidatorsResponse
		expErr  bool
	}{
		"delegation": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				pubKey := ed25519.GenPrivKey().PubKey()
				return json.Marshal(contract.ValidatorResponse{Validator: &contract.OperatorResponse{
					Operator: myOperatorAddr.String(),
					Pubkey:   contract.ValidatorPubkey{Ed25519: pubKey.Bytes()},
					Metadata: contract.MetadataFromDescription(types.ValidatorFixture().Description),
				}})
			}},
			exp: &distributiontypes.QueryDelegatorValidatorsResponse{Validators: []string{myOperatorAddr.String()}},
		},
		"unknown": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
			querier: SmartQuerierMock{func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return json.Marshal(contract.ValidatorResponse{})
			}},
			exp: &distributiontypes.QueryDelegatorValidatorsResponse{Validators: []string{}},
		},
		"error": {
			src: &distributiontypes.QueryDelegatorValidatorsRequest{DelegatorAddress: myOperatorAddr.String()},
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
			q := NewLegacyDistributionGRPCQuerier(contractSource, spec.querier)
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
