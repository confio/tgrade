package keeper

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

func TestSetGetPoEContractAddress(t *testing.T) {
	specs := map[string]struct {
		srcType   types.PoEContractType
		skipStore bool
		expErr    bool
	}{
		"staking": {
			srcType: types.PoEContractTypeStaking,
		},
		"valset": {
			srcType: types.PoEContractTypeValset,
		},
		"engagement": {
			srcType: types.PoEContractTypeEngagement,
		},
		"mixer": {
			srcType: types.PoEContractTypeMixer,
		},
		"undefined": {
			srcType: types.PoEContractTypeUndefined,
			expErr:  true,
		},
		"unsupported type": {
			srcType: types.PoEContractType(9999),
			expErr:  true,
		},
		"not stored": {
			srcType:   types.PoEContractType(9999),
			skipStore: true,
			expErr:    true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, _, k := createMinTestInput(t)
			var myAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
			if !spec.skipStore {
				k.SetPoEContractAddress(ctx, spec.srcType, myAddr)
			}
			gotAddr, gotErr := k.GetPoEContractAddress(ctx, spec.srcType)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, myAddr, gotAddr)
		})
	}
}

func TestIteratePoEContracts(t *testing.T) {
	ctx, _, k := createMinTestInput(t)
	storedTypes := make(map[types.PoEContractType]sdk.AccAddress)
	for c := range types.PoEContractType_name {
		src := types.PoEContractType(c)
		if src == types.PoEContractTypeUndefined {
			continue
		}
		var myAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
		k.SetPoEContractAddress(ctx, src, myAddr)
		storedTypes[src] = myAddr
	}
	readTypes := make(map[types.PoEContractType]sdk.AccAddress)
	k.IteratePoEContracts(ctx, func(c types.PoEContractType, addr sdk.AccAddress) bool {
		require.NotContains(t, readTypes, c)
		readTypes[c] = addr
		return false
	})
	assert.Equal(t, storedTypes, readTypes)
}

func TestUnbondingTime(t *testing.T) {
	ctx, _, k := createMinTestInput(t)
	var myAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	k.SetPoEContractAddress(ctx, types.PoEContractTypeStaking, myAddr)

	k.twasmKeeper = TwasmKeeperMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		require.Equal(t, myAddr, contractAddr)
		return json.Marshal(contract.UnbondingPeriodResponse{
			UnbondingPeriod: 60,
		})
	}}
	assert.Equal(t, time.Minute, k.UnbondingTime(ctx))
}
