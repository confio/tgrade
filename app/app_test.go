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
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
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
	poeGS := poetypes.GetGenesisStateFromAppState(marshaler, genesisState)
	if poeGS.GetSeedContracts() == nil {
		panic("not in seed mode")
	}

	bootstrapAccountAddr := sdk.AccAddress(rand.Bytes(address.Len))
	myGenTx, myAddr, _ := poetypes.RandomGenTX(t, 100)
	var authGenState authtypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenState)
	genAccounts := []authtypes.GenesisAccount{
		authtypes.NewBaseAccount(myAddr, nil, 0, 0),
		authtypes.NewBaseAccount(bootstrapAccountAddr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts
	genesisState[authtypes.ModuleName] = marshaler.MustMarshalJSON(&authGenState)

	var bankGenState banktypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenState)

	coins := sdk.Coins{sdk.NewCoin(poetypes.DefaultBondDenom, sdk.NewInt(1000000000))}
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: myAddr.String(), Coins: coins})
	bankGenState.Supply = bankGenState.Supply.Add(coins...)
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: bootstrapAccountAddr.String(), Coins: coins})
	bankGenState.Supply = bankGenState.Supply.Add(coins...)

	genAddrAndUpdateBalance := func(numAddr int, balance sdk.Coins) []string {
		genAddr := make([]string, numAddr)
		for i := 0; i < numAddr; i++ {
			addr := poetypes.RandomAccAddress().String()
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: addr, Coins: balance})
			genAddr[i] = addr
			bankGenState.Supply = bankGenState.Supply.Add(balance...)
		}
		return genAddr
	}
	// add 3 oc members
	ocMembers := genAddrAndUpdateBalance(3, coins)

	// add 2 ap members
	apMembers := genAddrAndUpdateBalance(2, coins)

	genesisState[banktypes.ModuleName] = marshaler.MustMarshalJSON(&bankGenState)

	// add system admin to not fail poe on validation
	poeGS.GetSeedContracts().BondDenom = poetypes.DefaultBondDenom
	poeGS.GetSeedContracts().GenTxs = []json.RawMessage{myGenTx}
	poeGS.GetSeedContracts().Engagement = []poetypes.TG4Member{{Address: myAddr.String(), Points: 10}}
	poeGS.GetSeedContracts().BootstrapAccountAddress = bootstrapAccountAddr.String()
	poeGS.GetSeedContracts().OversightCommunityMembers = ocMembers
	poeGS.GetSeedContracts().ArbiterPoolMembers = apMembers
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
	state, err := gapp.ibcKeeper.ClientKeeper.GetSelfConsensusState(ctx, height)
	// then
	require.NoError(t, err)
	assert.Equal(t, []byte("myAppHash"), state.GetRoot().GetHash())
	assert.Equal(t, uint64(now.UnixNano()), state.GetTimestamp())
}
