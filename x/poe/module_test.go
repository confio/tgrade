package poe

import (
	"encoding/json"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestInitGenesis(t *testing.T) {
	ctx, example := keeper.CreateDefaultTestInput(t)
	ctx = ctx.WithBlockHeight(0)
	deliverTXFn := simpleDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	app := NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())
	gs := types.GenesisStateFixture(
		func(m *types.GenesisState) {
			genTx, opAddr := types.RandomGenTX(t)
			m.GenTxs = []json.RawMessage{genTx}
			m.Engagement = []types.TG4Member{{Address: opAddr.String(), Weight: 100}}
			example.AccountKeeper.NewAccountWithAddress(ctx, opAddr)
			example.BankKeeper.SetBalances(ctx, opAddr, sdk.NewCoins(
				sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(1000000000)),
			))
		})
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	valset := app.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)
	require.NotEmpty(t, valset)
}

func simpleDeliverTXFn(t *testing.T, ctx sdk.Context, k keeper.Keeper, contractKeeper wasmtypes.ContractOpsKeeper, txDecoder sdk.TxDecoder) func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
	t.Helper()
	h := NewHandler(k, contractKeeper)
	return func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
		genTx, err := txDecoder(tx.GetTx())
		require.NoError(t, err)
		msgs := genTx.GetMsgs()
		require.Len(t, msgs, 1)
		msg := msgs[0].(*types.MsgCreateValidator)
		_, err = h(ctx, msg)
		require.NoError(t, err)
		return abci.ResponseDeliverTx{}
	}
}
