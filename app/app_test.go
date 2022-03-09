package app

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"

	poetypes "github.com/confio/tgrade/x/poe/types"
)

var emptyWasmOpts []wasm.Option = nil

func TestTgradeExport(t *testing.T) {
	t.Skip("Alex, export is not implemented")

	db := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)
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
	newGapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)

	_, err = newGapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func setupWithSingleValidatorGenTX(t *testing.T, genesisState GenesisState) {
	// a validator needs:
	// - signed genTX
	// - account object
	// - enough funds on the bank
	// - membership in engagement group
	marshaler := MakeEncodingConfig().Codec

	systemAdminAddr := sdk.AccAddress(rand.Bytes(address.Len))
	myGenTx, myAddr, _ := poetypes.RandomGenTX(t, 100)
	var authGenState authtypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenState)
	genAccounts := []authtypes.GenesisAccount{
		authtypes.NewBaseAccount(myAddr, nil, 0, 0),
		authtypes.NewBaseAccount(systemAdminAddr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts
	genesisState[authtypes.ModuleName] = marshaler.MustMarshalJSON(&authGenState)

	var bankGenState banktypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenState)

	coins := sdk.Coins{sdk.NewCoin(poetypes.DefaultBondDenom, sdk.NewInt(1000000000))}
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: myAddr.String(), Coins: coins.Sort()})
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: systemAdminAddr.String(), Coins: coins.Sort()})

	// add 3 oc members
	ocMembers := make([]string, 3)
	for i := 0; i < 3; i++ {
		addr := poetypes.RandomAccAddress().String()
		bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: addr, Coins: coins.Sort()})
		ocMembers[i] = addr
	}

	genesisState[banktypes.ModuleName] = marshaler.MustMarshalJSON(&bankGenState)

	// add system admin to not fail poe on validation
	poeGS := poetypes.GetGenesisStateFromAppState(marshaler, genesisState)
	poeGS.BondDenom = poetypes.DefaultBondDenom
	poeGS.GenTxs = []json.RawMessage{myGenTx}
	poeGS.Engagement = []poetypes.TG4Member{{Address: myAddr.String(), Points: 10}}
	poeGS.SystemAdminAddress = systemAdminAddr.String()
	poeGS.OversightCommunityMembers = ocMembers
	genesisState = poetypes.SetGenesisStateInAppState(marshaler, genesisState, poeGS)
}

// ensure that blocked addresses are properly set in bank keeper
func TestBlockedAddrs(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)

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

func TestIBCKeeperLazyInitialization(t *testing.T) {
	db := db.NewMemDB()
	gapp := NewTgradeApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, MakeEncodingConfig(), EmptyBaseAppOptions{}, emptyWasmOpts)
	genesisState := NewDefaultGenesisState()
	setupWithSingleValidatorGenTX(t, genesisState)

	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	now := time.Now().UTC()
	gapp.InitChain(
		abci.RequestInitChain{
			Time:          now,
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	gapp.Commit()
	// store some historic information
	header := tmproto.Header{ChainID: "testing-1", Height: 2, Time: now, AppHash: []byte("myAppHash")}
	gapp.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	gapp.Commit()

	ctx := gapp.BaseApp.NewContext(true, header)
	height := ibcclienttypes.Height{RevisionNumber: 1, RevisionHeight: 2}

	// when
	// https://github.com/cosmos/cosmos-sdk/blob/v0.42.9/x/ibc/core/02-client/keeper/keeper.go#L252
	state, found := gapp.ibcKeeper.ClientKeeper.GetSelfConsensusState(ctx, height)
	// then
	require.True(t, found)
	assert.Equal(t, []byte("myAppHash"), state.GetRoot().GetHash())
	assert.Equal(t, uint64(now.UnixNano()), state.GetTimestamp())
}
