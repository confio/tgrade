package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/confio/tgrade/x/poe/types"
)

type GenesisBalancesIterator = genutiltypes.GenesisBalancesIterator

// GenTxCmd builds the application's gentx command.
func GenTxCmd(mbm module.BasicManager, txEncCfg client.TxEncodingConfig, genBalIterator GenesisBalancesIterator, defaultNodeHome string) *cobra.Command {
	ipDefault, _ := server.ExternalIP()
	fsCreateValidator, defaultsDesc := CreateValidatorMsgFlagSet(ipDefault)

	cmd := &cobra.Command{
		Use:   "gentx [key_name] [liquid-amount] [vesting-amount]",
		Short: "Generate a genesis tx carrying a self delegation",
		Args:  cobra.ExactArgs(3),
		Long: fmt.Sprintf(`Generate a genesis transaction that creates a validator with a self-delegation,
that is signed by the key in the Keyring referenced by a given name. A node ID and Bech32 consensus
pubkey may optionally be provided. If they are omitted, they will be retrieved from the priv_validator.json
file. The following default parameters are included:
    %s

Example:
$ %s gentx my-key-name 1000000utgd --home=/path/to/home/dir --keyring-backend=os --chain-id=test-chain-1 \
    --moniker="myvalidator" \
    --details="..." \
    --security-contact="..." \
    --website="..."
`, defaultsDesc, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			cdc := clientCtx.Codec

			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}

			// read --nodeID, if empty take it from priv_validator.json
			if nodeIDString, _ := cmd.Flags().GetString(FlagNodeID); nodeIDString != "" {
				nodeID = nodeIDString
			}

			// read --pubkey, if empty take it from priv_validator.json
			if pkStr, _ := cmd.Flags().GetString(FlagPubKey); pkStr != "" {
				if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &valPubKey); err != nil {
					return errors.Wrap(err, "failed to unmarshal validator public key")
				}
			}

			genDoc, err := tmtypes.GenesisDocFromFile(config.GenesisFile())
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis doc file %s", config.GenesisFile())
			}

			var genesisState map[string]json.RawMessage
			if err = json.Unmarshal(genDoc.AppState, &genesisState); err != nil {
				return errors.Wrap(err, "failed to unmarshal genesis state")
			}

			if err = mbm.ValidateGenesis(cdc, txEncCfg, genesisState); err != nil {
				return errors.Wrap(err, "failed to validate genesis state")
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			name := args[0]
			key, err := clientCtx.Keyring.Key(name)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch '%s' from the keyring", name)
			}

			moniker := config.Moniker
			if m, _ := cmd.Flags().GetString(FlagMoniker); m != "" {
				moniker = m
			}

			// set flags for creating a gentx
			createValCfg, err := PrepareConfigForTxCreateValidator(cmd.Flags(), moniker, nodeID, genDoc.ChainID, valPubKey)
			if err != nil {
				return errors.Wrap(err, "error creating configuration to create validator msg")
			}

			liquidAmount := args[1]
			liquidCoins, err := sdk.ParseCoinsNormalized(liquidAmount)
			if err != nil {
				return errors.Wrap(err, "failed to parse coins for liquid amount")
			}
			vestedAmount := args[2]
			vestedCoins, err := sdk.ParseCoinsNormalized(vestedAmount)
			if err != nil {
				return errors.Wrap(err, "failed to parse coins for vested amount")
			}

			err = ValidateAccountInGenesis(genesisState, genBalIterator, key.GetAddress(), liquidCoins.Add(vestedCoins...), cdc)
			if err != nil {
				return errors.Wrap(err, "failed to validate account in genesis")
			}

			txFactory := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "error creating tx builder")
			}

			clientCtx = clientCtx.WithInput(inBuf).WithFromAddress(key.GetAddress())

			// The following line comes from a discrepancy between the `gentx`
			// and `create-validator` commands:
			// - `gentx` expects amount as an arg,
			// - `create-validator` expects amount as a required flag.
			// ref: https://github.com/cosmos/cosmos-sdk/issues/8251
			// Since gentx doesn't set the amount flag (which `create-validator`
			// reads from), we copy the amount arg into the valCfg directly.
			//
			// Ideally, the `create-validator` command should take a validator
			// config file instead of so many flags.
			// ref: https://github.com/cosmos/cosmos-sdk/issues/8177
			if createValCfg.LiquidAmount != "" && createValCfg.LiquidAmount != liquidAmount {
				return errors.New("liquid amount param and argument missmatch")
			}
			if createValCfg.VestingAmount != "" && createValCfg.VestingAmount != vestedAmount {
				return errors.New("vesting amount param and argument missmatch")
			}
			createValCfg.LiquidAmount = liquidAmount
			createValCfg.VestingAmount = vestedAmount

			// create a 'create-validator' message
			txBldr, msg, err := BuildCreateValidatorMsg(clientCtx, createValCfg, txFactory, true)
			if err != nil {
				return errors.Wrap(err, "failed to build create-validator message")
			}

			if key.GetType() == keyring.TypeOffline || key.GetType() == keyring.TypeMulti {
				cmd.PrintErrln("Offline key passed in. Use `tx sign` command to sign.")
				return authclient.PrintUnsignedStdTx(txBldr, clientCtx, []sdk.Msg{msg})
			}

			// write the unsigned transaction to the buffer
			w := bytes.NewBuffer([]byte{})
			clientCtx = clientCtx.WithOutput(w)

			if err = authclient.PrintUnsignedStdTx(txBldr, clientCtx, []sdk.Msg{msg}); err != nil {
				return errors.Wrap(err, "failed to print unsigned std tx")
			}

			// read the transaction
			stdTx, err := readUnsignedGenTxFile(clientCtx, w)
			if err != nil {
				return errors.Wrap(err, "failed to read unsigned gen tx file")
			}

			// sign the transaction and write it to the output file
			txBuilder, err := clientCtx.TxConfig.WrapTxBuilder(stdTx)
			if err != nil {
				return fmt.Errorf("error creating tx builder: %w", err)
			}

			err = authclient.SignTx(txFactory, clientCtx, name, txBuilder, true, true)
			if err != nil {
				return errors.Wrap(err, "failed to sign std tx")
			}

			outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			if outputDocument == "" {
				outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
				if err != nil {
					return errors.Wrap(err, "failed to create output file path")
				}
			}

			if err := writeSignedGenTx(clientCtx, outputDocument, stdTx); err != nil {
				return errors.Wrap(err, "failed to write signed gen tx")
			}

			cmd.PrintErrf("Genesis transaction written to %q\n", outputDocument)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagOutputDocument, "", "Write the genesis transaction JSON document to the given file instead of the default location")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.Flags().AddFlagSet(fsCreateValidator)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
	writePath := filepath.Join(rootDir, "config", "gentx")
	if err := tmos.EnsureDir(writePath, 0700); err != nil {
		return "", err
	}

	return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(clientCtx client.Context, r io.Reader) (sdk.Tx, error) {
	bz, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	aTx, err := clientCtx.TxConfig.TxJSONDecoder()(bz)
	if err != nil {
		return nil, err
	}

	return aTx, err
}

func writeSignedGenTx(clientCtx client.Context, outputDocument string, tx sdk.Tx) error {
	outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	json, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(outputFile, "%s\n", json)

	return err
}

// ValidateAccountInGenesis checks that the provided account has a sufficient
// balance in the set of genesis accounts.
func ValidateAccountInGenesis(
	appGenesisState map[string]json.RawMessage, genBalIterator GenesisBalancesIterator,
	addr sdk.Address, coins sdk.Coins, cdc codec.JSONCodec,
) error {
	gs := types.GetGenesisStateFromAppState(cdc, appGenesisState)
	bondDenom := gs.BondDenom

	var err error
	accountIsInGenesis := false

	genBalIterator.IterateGenesisBalances(cdc, appGenesisState,
		func(bal bankexported.GenesisBalance) (stop bool) {
			accAddress := bal.GetAddress()
			accCoins := bal.GetCoins()

			// ensure that account is in genesis
			if accAddress.Equals(addr) {
				// ensure account contains enough funds of default bond denom
				if coins.AmountOf(bondDenom).GT(accCoins.AmountOf(bondDenom)) {
					err = fmt.Errorf(
						"account %s has a balance in genesis, but it only has %v%s available to stake, not %v%s",
						addr, accCoins.AmountOf(bondDenom), bondDenom, coins.AmountOf(bondDenom), bondDenom,
					)

					return true
				}

				accountIsInGenesis = true
				return true
			}

			return false
		},
	)

	if err != nil {
		return err
	}

	if !accountIsInGenesis {
		return fmt.Errorf("account %s does not have a balance in the genesis state", addr)
	}

	return nil
}
