package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

// Get flags every time the simulator is run
func init() {
	simapp.GetSimulatorFlags()
}

type StoreKeysPrefixes struct {
	A        sdk.StoreKey
	B        sdk.StoreKey
	Prefixes [][]byte
}

// SetupSimulation wraps simapp.SetupSimulation in order to create any export directory if they do not exist yet
func SetupSimulation(dirPrefix, dbName string) (simtypes.Config, dbm.DB, string, log.Logger, bool, error) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation(dirPrefix, dbName)
	if err != nil {
		return simtypes.Config{}, nil, "", nil, false, err
	}

	paths := []string{config.ExportParamsPath, config.ExportStatePath, config.ExportStatsPath}
	for _, path := range paths {
		if len(path) == 0 {
			continue
		}

		path = filepath.Dir(path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				panic(err)
			}
		}
	}

	return config, db, dir, logger, skip, err
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, sdr sdk.StoreDecoderRegistry, kvAs, kvBs []kv.Pair) (log string) {
	for i := 0; i < len(kvAs); i++ {
		if len(kvAs[i].Value) == 0 && len(kvBs[i].Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		decoder, ok := sdr[storeName]
		if ok {
			log += decoder(kvAs[i], kvBs[i])
		} else {
			log += fmt.Sprintf("store A %q => %q\nstore B %q => %q\n", kvAs[i].Key, kvAs[i].Value, kvBs[i].Key, kvBs[i].Value)
		}
	}

	return log
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

func TestAppImportExport(t *testing.T) {
	now := time.Now().UTC()
	simapp.FlagGenesisTimeValue = now.Unix() // overwrite genesis time to something that is serializable and does not overflow
	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	encConf := MakeEncodingConfig()
	app := NewTgradeApp(logger, db, nil, true, map[int64]bool{}, dir, simapp.FlagPeriodValue, encConf, EmptyBaseAppOptions{}, nil, fauxMerkleModeOpt)
	require.Equal(t, appName, app.Name())
	app.sm.Modules = append(app.sm.Modules, nonSimModuleSetup{})
	// Run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simapp.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts,
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	t.Log("exporting genesis...")

	exported, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)

	t.Log("importing genesis...")

	_, newDB, newDir, _, _, err := SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()
	newApp := NewTgradeApp(logger, newDB, nil, true, map[int64]bool{}, newDir, simapp.FlagPeriodValue, encConf, EmptyBaseAppOptions{}, nil, fauxMerkleModeOpt)
	require.Equal(t, appName, newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContext(true, tmproto.Header{Height: app.LastBlockHeight(), Time: now})
	newApp.mm.InitGenesis(ctxB, app.AppCodec(), genesisState)
	newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)

	app.poeKeeper.IteratePoEContracts(ctxA, func(contractType types.PoEContractType, address sdk.AccAddress) bool {
		t.Logf("poe contract %s: %s\n", contractType, address.String())
		return false
	})
	t.Log("comparing stores...")

	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keys[authtypes.StoreKey], newApp.keys[authtypes.StoreKey], [][]byte{}},
		{app.keys[banktypes.StoreKey], newApp.keys[banktypes.StoreKey], [][]byte{banktypes.BalancesPrefix}},
		{app.keys[paramstypes.StoreKey], newApp.keys[paramstypes.StoreKey], [][]byte{}},
		{app.keys[evidencetypes.StoreKey], newApp.keys[evidencetypes.StoreKey], [][]byte{}},
		{app.keys[capabilitytypes.StoreKey], newApp.keys[capabilitytypes.StoreKey], [][]byte{}},
		{app.keys[ibchost.StoreKey], newApp.keys[ibchost.StoreKey], [][]byte{}},
		{app.keys[ibctransfertypes.StoreKey], newApp.keys[ibctransfertypes.StoreKey], [][]byte{}},
		{app.keys[authzkeeper.StoreKey], newApp.keys[authzkeeper.StoreKey], [][]byte{}},
		{app.keys[feegrant.StoreKey], newApp.keys[feegrant.StoreKey], [][]byte{}},
		{app.keys[twasm.StoreKey], newApp.keys[twasm.StoreKey], [][]byte{}},
	}

	prepareWasmStorage(ctxA, app, ctxB, newApp)

	// diff both stores
	for i, skp := range storeKeysPrefixes {
		t.Logf("pos: %d\n", i)
		storeA := ctxA.KVStore(skp.A)
		storeB := ctxB.KVStore(skp.B)

		failedKVAs, failedKVBs := sdk.DiffKVStores(storeA, storeB, skp.Prefixes)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")

		t.Logf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), skp.A, skp.B)
		simulationLog := GetSimulationLog(skp.A.Name(), app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs)
		max := 500
		if n := len(simulationLog); n < max {
			max = n
		}
		require.Len(t, failedKVAs, 0, simulationLog[0:max])
	}
}

// prepare both storages to be comparable
func prepareWasmStorage(ctxA sdk.Context, app *TgradeApp, ctxB sdk.Context, newApp *TgradeApp) {
	// delete persistent tx counter value
	ctxA.KVStore(app.keys[twasm.StoreKey]).Delete(wasmtypes.TXCounterPrefix)

	// reset contract code index in source DB for comparison with dest DB
	dropContractHistory := func(s store.KVStore, keys ...[]byte) {
		for _, key := range keys {
			prefixStore := prefix.NewStore(s, key)
			iter := prefixStore.Iterator(nil, nil)
			for ; iter.Valid(); iter.Next() {
				prefixStore.Delete(iter.Key())
			}
			iter.Close()
		}
	}
	prefixes := [][]byte{wasmtypes.ContractCodeHistoryElementPrefix, wasmtypes.ContractByCodeIDAndCreatedSecondaryIndexPrefix}
	dropContractHistory(ctxA.KVStore(app.keys[twasm.StoreKey]), prefixes...)
	dropContractHistory(ctxB.KVStore(newApp.keys[twasm.StoreKey]), prefixes...)

	normalizeContractInfo := func(ctx sdk.Context, app *TgradeApp) {
		var index uint64
		app.twasmKeeper.IterateContractInfo(ctx, func(address sdk.AccAddress, info wasmtypes.ContractInfo) bool {
			created := &wasmtypes.AbsoluteTxPosition{
				BlockHeight: uint64(0),
				TxIndex:     index,
			}
			info.Created = created
			store := ctx.KVStore(app.keys[twasm.StoreKey])
			store.Set(wasmtypes.GetContractAddressKey(address), app.appCodec.MustMarshal(&info))
			index++
			return false
		})
	}
	normalizeContractInfo(ctxA, app)
	normalizeContractInfo(ctxB, newApp)

	dropExportImportPrivilegedState := func(ctx sdk.Context, xapp *TgradeApp) {
		xapp.twasmKeeper.IteratePrivilegedContractsByType(ctx, twasmtypes.PrivilegeStateExporterImporter, func(pos uint8, contractAddr sdk.AccAddress) bool {
			prefixStore := prefix.NewStore(ctx.KVStore(xapp.keys[twasm.StoreKey]), wasmtypes.GetContractStorePrefix(contractAddr))
			iter := prefixStore.Iterator(nil, nil)
			for ; iter.Valid(); iter.Next() {
				if string(iter.Key()) == "admin" {
					prefixStore.Delete(iter.Key())
				}
			}
			iter.Close()
			return false
		})
	}
	// delete admin until fixed: https://github.com/confio/poe-contracts/issues/139
	dropExportImportPrivilegedState(ctxA, app)
}

func TestFullAppSimulation(t *testing.T) {
	simapp.FlagGenesisTimeValue = time.Now().Unix() // overwrite genesis time to something that is serializable and does not overflow

	config, db, dir, logger, skip, err := SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()
	encConf := MakeEncodingConfig()
	app := NewTgradeApp(logger, db, nil, true, map[int64]bool{}, t.TempDir(), simapp.FlagPeriodValue,
		encConf, simapp.EmptyAppOptions{}, nil, fauxMerkleModeOpt)
	require.Equal(t, "tgrade", app.Name())
	app.sm.Modules = append(app.sm.Modules, nonSimModuleSetup{})

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simapp.AppStateFn(app.appCodec, app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}

var _ module.AppModuleSimulation = &nonSimModuleSetup{}

// quick hack to register modules that do not provide default genesis for simulations
type nonSimModuleSetup struct{}

func (x nonSimModuleSetup) GenerateGenesisState(simstate *module.SimulationState) {
	for k, v := range map[string]proto.Message{
		icatypes.ModuleName: icatypes.DefaultGenesis(),
	} {
		simstate.GenState[k] = simstate.Cdc.MustMarshalJSON(v)
	}
}

func (x nonSimModuleSetup) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

func (x nonSimModuleSetup) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

func (x nonSimModuleSetup) RegisterStoreDecoder(registry sdk.StoreDecoderRegistry) {
}

func (x nonSimModuleSetup) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
