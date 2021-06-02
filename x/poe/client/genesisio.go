package client

import (
	"encoding/json"
	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetGenTxsInAppGenesisState - sets the genesis transactions in the app genesis state
func SetGenTxsInAppGenesisState(
	cdc codec.JSONMarshaler, txJSONEncoder sdk.TxEncoder, appGenesisState map[string]json.RawMessage, genTxs []sdk.Tx,
) (map[string]json.RawMessage, error) {

	genesisState := GetGenesisStateFromAppState(cdc, appGenesisState)
	genTxsBz := make([]json.RawMessage, 0, len(genTxs))

	for _, genTx := range genTxs {
		txBz, err := txJSONEncoder(genTx)
		if err != nil {
			return appGenesisState, err
		}

		genTxsBz = append(genTxsBz, txBz)
	}

	genesisState.GenTxs = genTxsBz
	return SetGenesisStateInAppState(cdc, appGenesisState, genesisState), nil
}

// GetGenesisStateFromAppState gets the genutil genesis state from the expected app state
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *types.GenesisState {
	var genesisState types.GenesisState
	if appState[poe.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[poe.ModuleName], &genesisState)
	}
	return &genesisState
}

// SetGenesisStateInAppState sets the genutil genesis state within the expected app state
func SetGenesisStateInAppState(
	cdc codec.JSONMarshaler, appState map[string]json.RawMessage, genesisState *types.GenesisState,
) map[string]json.RawMessage {

	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[poe.ModuleName] = genesisStateBz
	return appState
}
