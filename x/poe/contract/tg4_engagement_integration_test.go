package contract_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

//go:embed tg4_engagement.wasm
var tg4Engagement []byte

func TestEngagementUpdateAdmin(t *testing.T) {
	ctx, example := keeper.CreateDefaultTestInput(t)
	var systemAdminAddr sdk.AccAddress = rand.Bytes(address.Len)

	k := example.TWasmKeeper.GetContractKeeper()
	codeID, err := k.Create(ctx, systemAdminAddr, tg4Engagement, nil)
	require.NoError(t, err)

	var newAddress sdk.AccAddress = rand.Bytes(address.Len)

	tg4EngagementInitMsg := contract.TG4EngagementInitMsg{
		Admin: systemAdminAddr.String(),
		Members: []contract.TG4Member{{
			Addr:   newAddress.String(), // test only passes with new admin address in the group
			Weight: 1,
		}},
		PreAuthsHooks:    1,
		PreAuthsSlashing: 1,
		Denom:            "alx",
		Halflife:         1,
	}
	initMsgBz, err := json.Marshal(&tg4EngagementInitMsg)
	require.NoError(t, err)
	engagementContractAddr, _, err := k.Instantiate(ctx, codeID, systemAdminAddr, systemAdminAddr, initMsgBz, "engagement", nil)
	require.NoError(t, err)

	engagementContract := contract.NewEngagementContractAdapter(engagementContractAddr, example.TWasmKeeper, nil)

	// when
	gotErr := engagementContract.UpdateAdmin(ctx, newAddress, systemAdminAddr)
	require.NoError(t, gotErr)
}

func TestQueryDelegated(t *testing.T) {
	// setup contracts and seed some data
	ctx, example, vals, _ := setupPoEContracts(t)
	vals = clearTokenAmount(vals)

	var myUnknownAddr sdk.AccAddress = rand.Bytes(address.Len)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	require.NoError(t, err)

	specs := map[string]struct {
		ownerAddr string
		expVal    contract.DelegatedResponse
		expError  bool
	}{
		"query one validator": {
			ownerAddr: vals[0].OperatorAddress,
			expVal:    contract.DelegatedResponse{Delegated: vals[0].OperatorAddress},
		},
		"query other validator": {
			ownerAddr: vals[1].OperatorAddress,
			expVal:    contract.DelegatedResponse{Delegated: vals[1].OperatorAddress},
		},
		"query with invalid address": {
			ownerAddr: "not an address",
			expError:  true,
		},
		"query with unknown address": {
			ownerAddr: myUnknownAddr.String(),
			expVal:    contract.DelegatedResponse{Delegated: myUnknownAddr.String()},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ownerAddr, _ := sdk.AccAddressFromBech32(spec.ownerAddr)

			// when
			adaptor := contract.NewEngagementContractAdapter(contractAddr, example.TWasmKeeper, nil)
			gotVal, err := adaptor.QueryDelegated(ctx, ownerAddr)

			// then
			if spec.expError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.expVal, *gotVal)
		})
	}
}
