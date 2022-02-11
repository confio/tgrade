package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
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

	appparams "github.com/confio/tgrade/app/params"
	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/client/cli"
	"github.com/confio/tgrade/x/poe/types"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
)

const (
	myChainID      = "testing"
	myKey          = "myKey"
	bondDenum      = "utgd"
	initialBalance = 100
)

func TestGenTxCmd(t *testing.T) {
	specs := map[string]struct {
		stakingAmount sdk.Coin
		expErr        bool
	}{
		"all good": {
			stakingAmount: sdk.NewCoin(bondDenum, sdk.NewInt(1)),
		},
		"staked more than balance": {
			stakingAmount: sdk.NewCoin(bondDenum, sdk.NewInt(initialBalance+1)),
			expErr:        true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {

			dir := t.TempDir()
			encodingConfig := twasmkeeper.MakeEncodingConfig(t)
			addr, clientCtx, moduleManager := setupSystem(t, dir, encodingConfig)

			genBalIterator := banktypes.GenesisBalancesIterator{}
			cmd := cli.GenTxCmd(
				moduleManager,
				encodingConfig.TxConfig, genBalIterator, dir)

			_, out := testutil.ApplyMockIO(cmd)
			clientCtx.WithOutput(out)
			ctx := context.WithValue(context.Background(), sdkclient.ClientContextKey, &clientCtx)

			// when
			genTxFile := filepath.Join(dir, "myTx")
			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", flags.FlagChainID, myChainID),
				fmt.Sprintf("--%s=%s", flags.FlagOutputDocument, genTxFile),
				myKey,
				spec.stakingAmount.String(),
			})

			gotErr := cmd.ExecuteContext(ctx)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)

			// Validate generated transaction.
			open, err := os.Open(genTxFile)
			require.NoError(t, err)

			all, err := ioutil.ReadAll(open)
			require.NoError(t, err)

			tx, err := encodingConfig.TxConfig.TxJSONDecoder()(all)
			require.NoError(t, err)

			msgs := tx.GetMsgs()
			require.Len(t, msgs, 1)

			require.IsType(t, &types.MsgCreateValidator{}, msgs[0])
			require.Equal(t, []sdk.AccAddress{addr}, msgs[0].GetSigners())
			require.Equal(t, spec.stakingAmount, msgs[0].(*types.MsgCreateValidator).Value)
			require.NoError(t, tx.ValidateBasic())
			require.NoError(t, msgs[0].ValidateBasic())
		})
	}
}

func setupSystem(t *testing.T, workDir string, encodingConfig appparams.EncodingConfig) (sdk.AccAddress, sdkclient.Context, module.BasicManager) {
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
	gs := moduleManager.DefaultGenesis(encodingConfig.Codec)
	// with PoE setup
	state := types.GetGenesisStateFromAppState(encodingConfig.Codec, gs)
	state.BondDenom = bondDenum
	state.Engagement = append(state.Engagement, types.TG4Member{
		Address: addr.String(),
		Points:  1,
	})
	types.SetGenesisStateInAppState(encodingConfig.Codec, gs, state)
	// with bank setup
	bs := banktypes.GetGenesisStateFromAppState(encodingConfig.Codec, gs)
	bs.Balances = append(bs.Balances, banktypes.Balance{
		Address: addr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(bondDenum, sdk.NewInt(initialBalance))),
	})
	genesisStateBz := encodingConfig.Codec.MustMarshalJSON(bs)
	gs[banktypes.ModuleName] = genesisStateBz
	// with account setup
	var as authtypes.GenesisState
	encodingConfig.Codec.MustUnmarshalJSON(gs[authtypes.ModuleName], &as)
	genAccounts := []authtypes.GenesisAccount{authtypes.NewBaseAccount(addr, nil, 0, 0)}
	accounts, err := authtypes.PackAccounts(genAccounts)
	require.NoError(t, err)
	as.Accounts = accounts
	gs[authtypes.ModuleName] = encodingConfig.Codec.MustMarshalJSON(&as)

	appGenStateJSON, err := json.MarshalIndent(gs, "", "  ")
	require.NoError(t, err)

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
		WithCodec(encodingConfig.Codec).
		WithLegacyAmino(encodingConfig.Amino).
		WithTxConfig(encodingConfig.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{})
	return addr, clientCtx, moduleManager
}
