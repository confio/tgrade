package contract_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/contract"
)

//go:embed tgrade_trusted_circle.wasm
var tgTrustedCircles []byte

func TestInitTrustedCircle(t *testing.T) {
	// setup contracts and seed some data
	var creator sdk.AccAddress = rand.Bytes(address.Len)
	ctx, example, _, _ := setupPoEContracts(t)
	contractKeeper := example.TWasmKeeper.GetContractKeeper()

	depositAmount := sdk.NewCoin("utgd", sdk.NewInt(10_000_000))
	example.Faucet.Fund(ctx, creator, depositAmount)

	init := contract.TrustedCircleInitMsg{
		Name:                      "foo",
		Denom:                     depositAmount.Denom,
		EscrowAmount:              sdk.NewInt(10_000_000),
		VotingPeriod:              1,
		Quorum:                    sdk.NewDecWithPrec(1, 1),
		Threshold:                 sdk.NewDecWithPrec(5, 1),
		AllowEndEarly:             true,
		InitialMembers:            []string{creator.String()},
		DenyList:                  "",
		EditTrustedCircleDisabled: false,
		RewardDenom:               "utgd",
	}
	initBz, err := json.Marshal(init)
	require.NoError(t, err)
	t.Log(string(initBz))
	codeID, _, err := contractKeeper.Create(ctx, creator, tgTrustedCircles, nil)
	require.NoError(t, err)
	// when
	contractAddr, _, err := contractKeeper.Instantiate(ctx, codeID, creator, nil, initBz, "poe", sdk.NewCoins(depositAmount))
	// then
	require.NoError(t, err)
	require.NotEmpty(t, contractAddr)
}
