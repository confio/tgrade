package contract_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/keeper"

	"github.com/confio/tgrade/x/poe/contract"
)

//go:embed tg4_engagement.wasm
var tg4Engagement []byte

func TestEngagementUpdateAdmin(t *testing.T) {
	ctx, example := keeper.CreateDefaultTestInput(t)
	var systemAdminAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	k := example.TWasmKeeper.GetContractKeeper()
	codeID, err := k.Create(ctx, systemAdminAddr, tg4Engagement, nil)
	require.NoError(t, err)

	var newAddress sdk.AccAddress = rand.Bytes(sdk.AddrLen)

	tg4EngagementInitMsg := contract.TG4EngagementInitMsg{
		Admin: systemAdminAddr.String(),
		Members: []contract.TG4Member{{
			Addr:   newAddress.String(), // test only passes with new admin address in the group
			Weight: 1,
		}},
		PreAuthsHooks:    1,
		PreAuthsSlashing: 1,
		Token:            "alx",
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
