package keeper

import (
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
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

func TestSetGetPoESystemAdmin(t *testing.T) {
	ctx, _, k := createMinTestInput(t)
	require.Empty(t, k.GetPoESystemAdminAddress(ctx))
	var myAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	k.setPoESystemAdminAddress(ctx, myAddr)
	assert.Equal(t, myAddr, k.GetPoESystemAdminAddress(ctx))
}
func TestIteratePoEContracts(t *testing.T) {
	ctx, _, k := createMinTestInput(t)
	storedTypes := make(map[types.PoEContractType]sdk.AccAddress)
	for c, _ := range types.PoEContractType_name {
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
