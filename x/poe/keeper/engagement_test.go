package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetValidatorInitialEngagementPoints(t *testing.T) {
	var myOpAddr = RandomAddress(t)
	ctx, _, k := createMinTestInput(t)
	const initialPointsToGrant = 2
	k.setParams(ctx, types.NewParams(0, initialPointsToGrant, sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(10)))))
	engagementContractAddr := RandomAddress(t)
	k.SetPoEContractAddress(ctx, types.PoEContractTypeEngagement, engagementContractAddr)

	var capturedUpdateMsg []byte
	specs := map[string]struct {
		selfDelegation sdk.Coin
		queryFn        func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
		SudoFn         func(ctx sdk.Context, contractAddr sdk.AccAddress, msg []byte) ([]byte, error)
		expErr         bool
		expUpdateMsg   string
	}{
		"self delegation equal min; new account - default points": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(10)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				return json.Marshal(contract.TG4MemberResponse{})
			},
			SudoFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, msg []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				capturedUpdateMsg = msg
				return nil, nil
			},
			expUpdateMsg: fmt.Sprintf(`{"update_member":{"addr": %q, "weight":%d}}`, myOpAddr.String(), initialPointsToGrant),
		},
		"self delegation below min - no points": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(1)),
		},
		"self delegation with diff token - no points": {
			selfDelegation: sdk.NewCoin("XLA", sdk.NewInt(11)),
		},
		"operator has engagement points < initial - default points": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(11)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				var current = 1
				return json.Marshal(contract.TG4MemberResponse{Weight: &current})
			},
			SudoFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, msg []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				capturedUpdateMsg = msg
				return nil, nil
			},
			expUpdateMsg: fmt.Sprintf(`{"update_member":{"addr": %q, "weight":%d}}`, myOpAddr.String(), initialPointsToGrant),
		},
		"operator has engagement points = initial - no update": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(11)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				var current = initialPointsToGrant
				return json.Marshal(contract.TG4MemberResponse{Weight: &current})
			},
		},
		"operator has engagement points > initial - no update": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(11)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				var current = initialPointsToGrant + 1
				return json.Marshal(contract.TG4MemberResponse{Weight: &current})
			},
		},
		"engagement status query fails": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(11)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
		"engagement update command fails": {
			selfDelegation: sdk.NewCoin("ALX", sdk.NewInt(11)),
			queryFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
				require.Equal(t, engagementContractAddr, contractAddr)
				return json.Marshal(contract.TG4MemberResponse{})
			},
			SudoFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, msg []byte) ([]byte, error) {
				return nil, errors.New("testing")
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedUpdateMsg = nil
			k.twasmKeeper = TwasmKeeperMock{
				QuerySmartFn: spec.queryFn,
				SudoFn:       spec.SudoFn,
			}
			// when
			gotErr := k.SetValidatorInitialEngagementPoints(ctx, myOpAddr, spec.selfDelegation)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			if spec.expUpdateMsg == "" {
				assert.Nil(t, capturedUpdateMsg)
			} else {
				assert.JSONEq(t, spec.expUpdateMsg, string(capturedUpdateMsg))
			}
		})
	}
}
