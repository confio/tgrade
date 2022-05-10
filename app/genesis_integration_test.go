package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

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
	srcApp := NewTgradeApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		memDB,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		MakeEncodingConfig(),
		EmptyBaseAppOptions{},
		emptyWasmOpts,
	)

	init := NewDefaultGenesisState()
	setupWithSingleValidatorGenTX(t, init)
	doInitWithGenesis(srcApp, init)

	now := time.Now().UTC()
	for i := 0; i < 3; i++ { // add some blocks
		header := tmproto.Header{
			ChainID: "testing-1",
			Height:  int64(2 + i),
			Time:    now.Add(time.Duration(i+1) * time.Hour), // big step > epoch
			AppHash: []byte(fmt.Sprintf("myAppHash%d", i)),
		}
		srcApp.BeginBlock(abci.RequestBeginBlock{Header: header})
		srcApp.Commit()
	}

	// create a new instance to read state only from db. initchain was not called on this
	srcApp = NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), memDB, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)

	exported, err := srcApp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")

	// then ensure that valset contract state is exported
	var gs GenesisState
	require.NoError(t, tmjson.Unmarshal(exported.AppState, &gs))
	require.Contains(t, gs, twasm.ModuleName)
	var twasmGs twasmtypes.GenesisState
	require.NoError(t, srcApp.appCodec.UnmarshalJSON(gs[twasm.ModuleName], &twasmGs))
	require.NoError(t, twasmGs.ValidateBasic())

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
	require.NoError(t, srcApp.appCodec.UnmarshalJSON(gs[poetypes.ModuleName], &poeGs))
	require.NoError(t, poetypes.ValidateGenesis(poeGs, MakeEncodingConfig().TxConfig.TxJSONDecoder()))
	// now import the state on a fresh DB
	memDB = db.NewMemDB()
	newApp := NewTgradeApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)),
		memDB,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		MakeEncodingConfig(),
		EmptyBaseAppOptions{},
		emptyWasmOpts,
	)
	doInitWithGenesis(newApp, gs)
}
