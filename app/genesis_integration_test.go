package app

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	poetypes "github.com/confio/tgrade/x/poe/types"

	"github.com/stretchr/testify/assert"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/confio/tgrade/x/twasm"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

func TestTgradeGenesisExportImport(t *testing.T) {
	doInitWithGenesis := func(gapp *TgradeApp, genesisState GenesisState) {
		setupWithSingleValidatorGenTX(t, genesisState)

		stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
		require.NoError(t, err)

		// Initialize the chain
		gapp.InitChain(
			abci.RequestInitChain{
				Time:          time.Now().UTC(),
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: stateBytes,
			},
		)
		gapp.Commit()
	}
	memDB := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), memDB, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)
	initGenesis := NewDefaultGenesisState()

	doInitWithGenesis(gapp, initGenesis)
	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), memDB, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)

	exported, err := newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")

	// then ensure that valset contract state is exported
	var gs GenesisState
	require.NoError(t, tmjson.Unmarshal(exported.AppState, &gs))
	require.Contains(t, gs, twasm.ModuleName)
	var twasmGs twasmtypes.GenesisState
	require.NoError(t, newGapp.appCodec.UnmarshalJSON(gs[twasm.ModuleName], &twasmGs))
	// todo (Alex): enable require.NoError(t, twasmGs.ValidateBasic())

	var customModelContractCount int
	for _, v := range twasmGs.Contracts {
		var ext twasmtypes.TgradeContractDetails
		require.NoError(t, v.ContractInfo.ReadExtension(&ext))
		if ext.HasRegisteredPrivilege(twasmtypes.PrivilegeStateExporterImporter) {
			m := v.GetCustomModel()
			require.NotNil(t, m)
			assert.Nil(t, v.GetKvModel())
			assert.NoError(t, m.Msg.ValidateBasic())
			customModelContractCount++
		} else {
			m := v.GetKvModel()
			require.NotNil(t, m)
			assert.Nil(t, v.GetCustomModel())
		}
	}
	assert.Equal(t, 1, customModelContractCount)

	// ensure poe is correct
	var poeGs poetypes.GenesisState
	require.NoError(t, newGapp.appCodec.UnmarshalJSON(gs[poetypes.ModuleName], &poeGs))
	require.NoError(t, poetypes.ValidateGenesis(poeGs, MakeEncodingConfig().TxConfig.TxJSONDecoder()))
	t.Log(string(gs[twasm.ModuleName]))
	// now import the state on a fresh DB
	memDB = db.NewMemDB()
	newApp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), memDB, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)
	doInitWithGenesis(newApp, gs)
}
