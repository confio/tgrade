package app

import (
	"encoding/json"
	"github.com/confio/tgrade/x/poe"
	poetypes "github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cfg "github.com/tendermint/tendermint/config"
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

const stakingToken = "utgd"

func TestTgradeExport(t *testing.T) {
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

	myGenTx, myAddr := randomGenTX(t)

	var authGenState authtypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenState)
	genAccounts := []authtypes.GenesisAccount{authtypes.NewBaseAccount(myAddr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	authGenState.Accounts = accounts
	genesisState[authtypes.ModuleName] = marshaler.MustMarshalJSON(&authGenState)

	var bankGenState banktypes.GenesisState
	marshaler.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenState)

	coins := sdk.Coins{sdk.NewCoin(stakingToken, sdk.NewInt(1000000000))}
	bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: myAddr.String(), Coins: coins.Sort()})
	genesisState[banktypes.ModuleName] = marshaler.MustMarshalJSON(&bankGenState)

	// add system admin to not fail poe on validation
	var poeGS poetypes.GenesisState
	err = marshaler.UnmarshalJSON(genesisState[poe.ModuleName], &poeGS)
	require.NoError(t, err)
	poeGS.SystemAdminAddress = sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String()
	poeGS.GenTxs = []json.RawMessage{myGenTx}
	poeGS.Engagement = []poetypes.TG4Member{{Address: myAddr.String(), Weight: 10}}
	genesisState[poe.ModuleName], err = marshaler.MarshalJSON(&poeGS)
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

// randomGenTX returns a signed genesis tx
func randomGenTX(t *testing.T) (json.RawMessage, sdk.AccAddress) {
	t.Helper()
	nodeConfig := cfg.TestConfig()
	nodeConfig.RootDir = t.TempDir()
	nodeConfig.NodeKey = "key.json"
	_, valPubKey, err := genutil.InitializeNodeValidatorFiles(nodeConfig)
	require.NoError(t, err)
	//setup keyring
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, t.TempDir(), nil)
	require.NoError(t, err)
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(t, err)
	const myKey = "myKey"
	addr, _, err := server.GenerateSaveCoinKey(kb, myKey, true, algo)
	require.NoError(t, err)

	// prepare genesis tx
	valTokens := sdk.TokensFromConsensusPower(100)
	createValMsg, err := poetypes.NewMsgCreateValidator(
		sdk.ValAddress(addr),
		valPubKey,
		sdk.NewCoin(stakingToken, valTokens),
		stakingtypes.NewDescription("testing", "", "", "", ""),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
	)
	require.NoError(t, err)
	txConfig := MakeEncodingConfig().TxConfig
	txBuilder := txConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(createValMsg)
	require.NoError(t, err)

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID("").
		WithKeybase(kb).
		WithTxConfig(txConfig)

	err = tx.Sign(txFactory, myKey, txBuilder, true)
	require.NoError(t, err)

	txBz, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	return txBz, addr
}
