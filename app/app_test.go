package app

import (
	"encoding/json"
	poetypes "github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/CosmWasm/wasmd/x/wasm"
	abci "github.com/tendermint/tendermint/abci/types"
)

var emptyWasmOpts []wasm.Option = nil

func TestTgradeExport(t *testing.T) {
	t.Skip("Alex, this is not implemented")
	db := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, EmptyBaseAppOptions{}, emptyWasmOpts)
	genesisState := NewDefaultGenesisState()

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

	// Making a new app object with the db, so that initchain hasn't been called
	newGapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, EmptyBaseAppOptions{}, emptyWasmOpts)
	_, err = newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func setupWithSingleValidatorGenTX(t *testing.T, genesisState GenesisState) {
	// a validator needs:
	// - signed genTX
	// - account object
	// - enough funds on the bank
	// - membership in engagement group
	marshaler := MakeEncodingConfig().Marshaler

	myGenTx, myAddr, _ := poetypes.RandomGenTX(t, 100)

	var authGenState authtypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenState)
	genAccounts := []authtypes.GenesisAccount{authtypes.NewBaseAccount(myAddr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts
	genesisState[authtypes.ModuleName] = marshaler.MustMarshalJSON(&authGenState)

	var bankGenState banktypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenState)

	coins := sdk.Coins{sdk.NewCoin(poetypes.DefaultBondDenom, sdk.NewInt(1000000000))}
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: myAddr.String(), Coins: coins.Sort()})
	genesisState[banktypes.ModuleName] = marshaler.MustMarshalJSON(&bankGenState)

	// add system admin to not fail poe on validation
	poeGS := poetypes.GetGenesisStateFromAppState(marshaler, genesisState)
	poeGS.BondDenom = poetypes.DefaultBondDenom
	poeGS.SystemAdminAddress = sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String()
	poeGS.GenTxs = []json.RawMessage{myGenTx}
	poeGS.Engagement = []poetypes.TG4Member{{Address: myAddr.String(), Weight: 10}}
	genesisState = poetypes.SetGenesisStateInAppState(marshaler, genesisState, poeGS)
}

// ensure that blocked addresses are properly set in bank keeper
func TestBlockedAddrs(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, EmptyBaseAppOptions{}, emptyWasmOpts)

	for acc := range maccPerms {
		t.Run(acc, func(t *testing.T) {
			require.True(t, gapp.bankKeeper.BlockedAddr(gapp.accountKeeper.GetModuleAddress(acc)),
				"ensure that blocked addresses are properly set in bank keeper",
			)
		})
	}
}

func TestGetMaccPerms(t *testing.T) {
	dup := GetMaccPerms()
	require.Equal(t, maccPerms, dup, "duplicated module account permissions differed from actual module account permissions")
}

func setGenesis(gapp *TgradeApp) error {
	genesisState := NewDefaultGenesisState()
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return err
	}

	// Initialize the chain
	gapp.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	gapp.Commit()
	return nil
}
