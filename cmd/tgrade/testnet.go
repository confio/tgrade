package main

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	poeclient "github.com/confio/tgrade/x/poe/client"
	poetypes "github.com/confio/tgrade/x/poe/types"
)

const stakingToken = "utgd"

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
	// custom flags
	flagCommitTimeout = "commit-timeout"
	flagSingleHost    = "single-host"
)

// get cmd to initialize all files for tendermint testnet and application
func testnetCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a tgrade testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	tgrade testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			outputDir, _ := cmd.Flags().GetString(flagOutputDir)
			keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			minGasPrices, _ := cmd.Flags().GetString(server.FlagMinGasPrices)
			nodeDirPrefix, _ := cmd.Flags().GetString(flagNodeDirPrefix)
			nodeDaemonHome, _ := cmd.Flags().GetString(flagNodeDaemonHome)
			startingIPAddress, _ := cmd.Flags().GetString(flagStartingIPAddress)
			numValidators, _ := cmd.Flags().GetInt(flagNumValidators)
			algo, _ := cmd.Flags().GetString(flags.FlagKeyAlgorithm)

			config.Consensus.TimeoutCommit, err = cmd.Flags().GetDuration(flagCommitTimeout)
			if err != nil {
				return err
			}
			singleMachine, err := cmd.Flags().GetBool(flagSingleHost)
			if err != nil {
				return err
			}

			return InitTestnet(
				clientCtx, cmd, config, mbm, genBalIterator, outputDir, chainID, minGasPrices,
				nodeDirPrefix, nodeDaemonHome, startingIPAddress, keyringBackend, algo, numValidators,
				singleMachine,
			)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./mytestnet", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "tgrade", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", stakingToken), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01tgd)")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")
	// tgrade
	cmd.Flags().Duration(flagCommitTimeout, 5*time.Second, "Time to wait after a block commit before starting on the new height")
	cmd.Flags().Bool(flagSingleHost, false, "Cluster runs on a single host machine with different ports")
	return cmd
}

const nodeDirPerm = 0755

// InitTestnet Initialize the testnet
func InitTestnet(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator banktypes.GenesisBalancesIterator,
	outputDir, chainID, minGasPrices, nodeDirPrefix, nodeDaemonHome, startingIPAddress, keyringBackend, algoStr string,
	numValidators int, singleMachine bool,
) error {

	if chainID == "" {
		chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	nodeIDs := make([]string, numValidators)
	valPubKeys := make([]cryptotypes.PubKey, numValidators)

	appConfig := srvconfig.DefaultConfig()
	appConfig.MinGasPrices = minGasPrices
	appConfig.API.Enable = true
	appConfig.Telemetry.Enabled = true
	appConfig.Telemetry.PrometheusRetentionTime = 60
	appConfig.Telemetry.EnableHostnameLabel = false
	appConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", chainID}}

	var (
		genAccounts      []authtypes.GenesisAccount
		genBalances      []banktypes.Balance
		genOCMemberAddrs []string
		genAPMemberAddrs []string
		genFiles         []string
	)
	const (
		rpcPort     = 26657
		apiPort     = 1317
		grpcPort    = 9090
		grpcWebPort = 8090
	)
	p2pPortStart := 26656

	addGenAccount := func(addr sdk.AccAddress, coins ...sdk.Coin) {
		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: sdk.Coins(coins).Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	var adminAddr sdk.AccAddress
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < numValidators; i++ {
		var portOffset int
		if singleMachine {
			portOffset = i
			p2pPortStart = 16656 // use different start point to not conflict with rpc port
			nodeConfig.P2P.AddrBookStrict = false
			nodeConfig.P2P.PexReactor = false
			nodeConfig.P2P.AllowDuplicateIP = true
		}

		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")

		nodeConfig.SetRoot(nodeDir)
		appConfig.API.Address = fmt.Sprintf("tcp://0.0.0.0:%d", apiPort+portOffset)
		appConfig.GRPC.Address = fmt.Sprintf("0.0.0.0:%d", grpcPort+portOffset)
		appConfig.GRPCWeb.Address = fmt.Sprintf("0.0.0.0:%d", grpcWebPort+portOffset)

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeConfig.Moniker = nodeDirName

		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:%d", nodeIDs[i], ip, p2pPortStart+portOffset)
		genFiles = append(genFiles, nodeConfig.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, nodeDir, inBuf)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(algoStr, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, true, algo)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
		if i == 0 { // generate new key for system admin in node0 keychain. This keychain is used by system tests
			// PoE setup
			adminAddr, _, err = server.GenerateSaveCoinKey(kb, "systemadmin", true, algo)
			if err != nil {
				_ = os.RemoveAll(outputDir)
				return err
			}
			addGenAccount(adminAddr, sdk.NewCoin(stakingToken, sdk.NewInt(100_000_000_000)))
			// add a number of OC members
			for i := 0; i < 3; i++ {
				memberAddr, _, err := server.GenerateSaveCoinKey(kb, fmt.Sprintf("oc-member-%d", i+1), true, algo)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}
				addGenAccount(memberAddr, sdk.NewCoin(stakingToken, sdk.NewInt(1_000_000_000)))
				genOCMemberAddrs = append(genOCMemberAddrs, memberAddr.String())
			}

			// add a number of AP members
			for i := 0; i < 2; i++ {
				memberAddr, _, err := server.GenerateSaveCoinKey(kb, fmt.Sprintf("ap-member-%d", i+1), true, algo)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}
				addGenAccount(memberAddr, sdk.NewCoin(stakingToken, sdk.NewInt(1_000_000_000)))
				genAPMemberAddrs = append(genAPMemberAddrs, memberAddr.String())
			}
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(fmt.Sprintf("%v.json", "key_seed"), nodeDir, cliPrint); err != nil {
			return err
		}

		accTokens := sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)
		accStakingTokens := sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction)
		addGenAccount(addr,
			sdk.NewCoin(fmt.Sprintf("%stoken", nodeDirName), accTokens),
			sdk.NewCoin(stakingToken, accStakingTokens))

		valTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
		moniker := fmt.Sprintf("moniker-%d", i)
		createValMsg, err := poetypes.NewMsgCreateValidator(
			addr,
			valPubKeys[i],
			sdk.NewCoin(stakingToken, valTokens),
			sdk.NewCoin(stakingToken, sdk.ZeroInt()),
			// moniker must be at least 3 chars. let's pad it to ensure
			stakingtypes.NewDescription(moniker, "", "", "", ""),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appConfig)
	}
	if err := initGenFiles(clientCtx, mbm, chainID, genAccounts, genBalances, genFiles, numValidators, adminAddr, genOCMemberAddrs, genAPMemberAddrs); err != nil {
		return err
	}

	err := collectGenFiles(
		clientCtx, nodeConfig, chainID, nodeIDs, valPubKeys, numValidators,
		outputDir, nodeDirPrefix, nodeDaemonHome, genBalIterator,
		rpcPort, p2pPortStart, singleMachine,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context,
	mbm module.BasicManager,
	chainID string,
	genAccounts []authtypes.GenesisAccount,
	genBalances []banktypes.Balance,
	genFiles []string,
	numValidators int,
	admin sdk.AccAddress,
	ocMemberAddrs []string,
	apMemberAddrs []string,
) error {
	appGenState := mbm.DefaultGenesis(clientCtx.Codec)

	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genBalances)
	var total sdk.Coins
	for _, v := range genBalances {
		total = total.Add(v.Coins...)
	}

	bankGenState.Supply = bankGenState.Supply.Add(total...)
	bankGenState.Balances = genBalances
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)
	poeGenesisState := poetypes.GetGenesisStateFromAppState(clientCtx.Codec, appGenState)
	for i, addr := range genAccounts {
		poeGenesisState.Engagement = append(poeGenesisState.Engagement, poetypes.TG4Member{
			Address: addr.GetAddress().String(),
			Points:  uint64(len(genAccounts) - i), // unique weight
		})
	}
	poeGenesisState.SystemAdminAddress = admin.String()
	poeGenesisState.OversightCommunityMembers = ocMemberAddrs
	poeGenesisState.ArbiterPoolMembers = apMemberAddrs
	poetypes.SetGenesisStateInAppState(clientCtx.Codec, appGenState, poeGenesisState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}
	// quick & dirty solution to set our denom instead of sdk default
	genDoc.AppState = []byte(strings.Replace(string(genDoc.AppState), "\"stake\"", fmt.Sprintf("%q", stakingToken), -1))

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context,
	nodeConfig *tmconfig.Config,
	chainID string,
	nodeIDs []string,
	valPubKeys []cryptotypes.PubKey,
	numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string,
	genBalIterator banktypes.GenesisBalancesIterator,
	rpcPortStart, p2pPortStart int,
	singleMachine bool,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()
	for i := 0; i < numValidators; i++ {
		var portOffset int
		if singleMachine {
			portOffset = i
		}
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName
		nodeConfig.RPC.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", rpcPortStart+portOffset)
		nodeConfig.P2P.ListenAddress = fmt.Sprintf("tcp://0.0.0.0:%d", p2pPortStart+portOffset)

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := poeclient.AddGenTxsToGenesisFile(clientCtx.Codec, clientCtx.TxConfig, nodeConfig, initCfg, *genDoc, genBalIterator)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := nodeConfig.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}
