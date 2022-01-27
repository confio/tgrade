package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetGenesisStateFromAppState gets the genesis state from the expected app state
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return &genesisState
}

// SetGenTxsInAppGenesisState - sets the genesis transactions in the app genesis state
func SetGenTxsInAppGenesisState(
	cdc codec.JSONCodec, txJSONEncoder sdk.TxEncoder, appGenesisState map[string]json.RawMessage, genTxs []sdk.Tx,
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

// SetGenesisStateInAppState sets the genesis state within the expected app state
func SetGenesisStateInAppState(
	cdc codec.JSONCodec, appState map[string]json.RawMessage, genesisState *GenesisState,
) map[string]json.RawMessage {

	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}
