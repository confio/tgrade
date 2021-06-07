package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/client/cli"
	"github.com/confio/tgrade/x/poe/types"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/stretchr/testify/require"
	cfg "github.com/tendermint/tendermint/config"
	tmtypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	myChainID = "testing"
	myKey     = "myKey"
	bondDenum = "ustake"
)

func TestGenTxCmd(t *testing.T) {
	dir := t.TempDir()
	encodingConfig := twasmkeeper.MakeEncodingConfig(t)
	addr, clientCtx, moduleManager := setupSystem(t, dir, encodingConfig)

	genBalIterator := banktypes.GenesisBalancesIterator{}
	cmd := cli.GenTxCmd(
		moduleManager,
		encodingConfig.TxConfig, genBalIterator, dir)

	_, out := testutil.ApplyMockIO(cmd)
	clientCtx.WithOutput(out)

	ctx := context.Background()
	ctx = context.WithValue(ctx, sdkclient.ClientContextKey, &clientCtx)

	amount := sdk.NewCoin("ustake", sdk.NewInt(12))
	genTxFile := filepath.Join(dir, "myTx")
	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=%s", flags.FlagChainID, myChainID),
		fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile),
		myKey,
		amount.String(),
	})

	err := cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	// Validate generated transaction.
	open, err := os.Open(genTxFile)
	require.NoError(t, err)

	all, err := ioutil.ReadAll(open)
	require.NoError(t, err)

	tx, err := encodingConfig.TxConfig.TxJSONDecoder()(all)
	require.NoError(t, err)

	msgs := tx.GetMsgs()
	require.Len(t, msgs, 1)

	require.Equal(t, types.TypeMsgCreateValidator, msgs[0].Type())
	require.Equal(t, []sdk.AccAddress{addr}, msgs[0].GetSigners())
	require.Equal(t, amount, msgs[0].(*types.MsgCreateValidator).Value)
	err = tx.ValidateBasic()
	require.NoError(t, err)
}

func setupSystem(t *testing.T, workDir string, encodingConfig params.EncodingConfig) (sdk.AccAddress, sdkclient.Context, module.BasicManager) {
	// init config dir
	nodeConfig := cfg.TestConfig()
	nodeConfig.RootDir = t.TempDir()
	nodeConfig.NodeKey = "key.json"
	require.NoError(t, os.MkdirAll(filepath.Join(workDir, "config"), 0755))

	// create node key file
	_, _, err := genutil.InitializeNodeValidatorFiles(nodeConfig)
	require.NoError(t, err)

	// create operator address and key
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, t.TempDir(), nil)
	require.NoError(t, err)
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(t, err)
	addr, _, err := server.GenerateSaveCoinKey(kb, myKey, true, algo)
	require.NoError(t, err)

	// create genesis
	moduleManager := module.NewBasicManager(poe.AppModuleBasic{}, auth.AppModuleBasic{}, bank.AppModuleBasic{})
	moduleManager.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	gs := moduleManager.DefaultGenesis(encodingConfig.Marshaler)
	// with PoE setup
	state := types.GetGenesisStateFromAppState(encodingConfig.Marshaler, gs)
	state.BondDenom = bondDenum
	state.Engagement = append(state.Engagement, types.TG4Member{
		Address: addr.String(),
		Weight:  1,
	})
	types.SetGenesisStateInAppState(encodingConfig.Marshaler, gs, state)
	// with bank setup
	bs := banktypes.GetGenesisStateFromAppState(encodingConfig.Marshaler, gs)
	bs.Balances = append(bs.Balances, banktypes.Balance{
		Address: addr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(bondDenum, sdk.NewInt(1000))),
	})
	genesisStateBz := encodingConfig.Marshaler.MustMarshalJSON(bs)
	gs[banktypes.ModuleName] = genesisStateBz
	// with account setup
	var as authtypes.GenesisState
	encodingConfig.Marshaler.MustUnmarshalJSON(gs[authtypes.ModuleName], &as)
	genAccounts := []authtypes.GenesisAccount{authtypes.NewBaseAccount(addr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	as.Accounts = accounts
	gs[authtypes.ModuleName] = encodingConfig.Marshaler.MustMarshalJSON(&as)

	appGenStateJSON, err := json.MarshalIndent(gs, "", "  ")
	require.NoError(t, err)
	t.Log(string(appGenStateJSON))

	genDoc := tmtypes.GenesisDoc{
		ChainID:    myChainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}
	require.NoError(t, genDoc.SaveAs(filepath.Join(workDir, "config", "genesis.json")))
	clientCtx := sdkclient.Context{}.
		WithKeyringDir(workDir).
		WithKeyring(kb).
		WithHomeDir(workDir).
		WithChainID(myChainID).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithLegacyAmino(encodingConfig.Amino).
		WithTxConfig(encodingConfig.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{})
	return addr, clientCtx, moduleManager
}
