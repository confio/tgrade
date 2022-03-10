package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	poecontracts "github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

// default values
var (
	DefaultTokens = sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	defaultAmount = DefaultTokens.String() + types.DefaultBondDenom
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd() *cobra.Command {
	poeTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "PoE transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poeTxCmd.AddCommand(
		NewCreateValidatorCmd(),
		NewEditValidatorCmd(),
		NewDelegateCmd(),
		NewUnbondCmd(),
		NewUnjailTxCmd(),
	)

	return poeTxCmd
}

func NewCreateValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator",
		Short: "create new validator initialized with a self-delegation to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).
				WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			txf, msg, err := NewBuildCreateValidatorMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
		Long: fmt.Sprintln(`Create validator gentx with PoE parameters. Considering this is run pre-genesis
--generate-only flag should be set. Otherwise client will try to connect to non-existent node. Also pass in
address instead of keyname to --from flag.

Example:
$ tgrade tx poe create-validator \
	--amount 1000utgd
	--from tgrade1n4kjhlrpapnpv0n0e3048ydftrjs9m6mm473jf \
	--pubkey tgradevalconspub1zcjduepqu7xf85mmfyv5p9m8mc6wk0u0pcjwcpr9p8wsv4h96dhpxqyxs4uqv06vlq \
	--home $APP_HOME \
	--chain-id=tgrade-int \
    --moniker="myvalidator" \
    --details="..." \
    --security-contact="..." \
    --website="..." \
	--generate-only`),
	}

	cmd.Flags().AddFlagSet(FlagSetPublicKey())
	cmd.Flags().AddFlagSet(FlagSetAmount())
	cmd.Flags().AddFlagSet(flagSetDescriptionCreate())

	cmd.Flags().String(FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	_ = cmd.MarkFlagRequired(FlagAmount)
	_ = cmd.MarkFlagRequired(FlagPubKey)
	_ = cmd.MarkFlagRequired(FlagMoniker)

	return cmd
}

func NewBuildCreateValidatorMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	fAmount, _ := fs.GetString(FlagAmount)
	amount, err := sdk.ParseCoinNormalized(fAmount)
	if err != nil {
		return txf, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	pkStr, err := fs.GetString(FlagPubKey)
	if err != nil {
		return txf, nil, err
	}

	var pk cryptotypes.PubKey
	if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &pk); err != nil {
		return txf, nil, err
	}

	moniker, _ := fs.GetString(FlagMoniker)
	identity, _ := fs.GetString(FlagIdentity)
	website, _ := fs.GetString(FlagWebsite)
	security, _ := fs.GetString(FlagSecurityContact)
	details, _ := fs.GetString(FlagDetails)
	description := stakingtypes.NewDescription(
		moniker,
		identity,
		website,
		security,
		details,
	)

	msg, err := types.NewMsgCreateValidator(valAddr, pk, amount, description)
	if err != nil {
		return txf, nil, err
	}
	if err := msg.ValidateBasic(); err != nil {
		return txf, nil, err
	}

	genOnly, _ := fs.GetBool(flags.FlagGenerateOnly)
	if genOnly {
		ip, _ := fs.GetString(FlagIP)
		nodeID, _ := fs.GetString(FlagNodeID)

		if nodeID != "" && ip != "" {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txf, msg, nil
}

func NewEditValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-validator",
		Short: "edit an existing validator account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			valAddr := clientCtx.GetFromAddress()
			moniker, _ := cmd.Flags().GetString(FlagMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)
			description := stakingtypes.NewDescription(moniker, identity, website, security, details)

			msg := types.NewMsgUpdateValidator(valAddr, description)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(flagSetValidatorDescription(stakingtypes.DoNotModifyDesc))
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CreateValidatorMsgFlagSet Return the flagset, particular flags, and a description of defaults
// this is anticipated to be used with the gen-tx
func CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	fsCreateValidator := flag.NewFlagSet("", flag.ContinueOnError)
	fsCreateValidator.String(FlagIP, ipDefault, "The node's public IP")
	fsCreateValidator.String(FlagNodeID, "", "The node's NodeID")

	fsCreateValidator.AddFlagSet(flagSetValidatorDescription(""))
	fsCreateValidator.AddFlagSet(FlagSetAmount())
	fsCreateValidator.AddFlagSet(FlagSetPublicKey())

	defaultsDesc = fmt.Sprintf(`delegation amount: %s`, defaultAmount)

	return fsCreateValidator, defaultsDesc
}

func flagSetValidatorDescription(defaultValue string) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagMoniker, defaultValue, "The validator's name")
	fs.String(FlagIdentity, defaultValue, "The (optional) identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, defaultValue, "The validator's (optional) website")
	fs.String(FlagSecurityContact, defaultValue, "The validator's (optional) security contact email")
	fs.String(FlagDetails, defaultValue, "The validator's (optional) details")
	return fs
}

type TxCreateValidatorConfig struct {
	ChainID string
	NodeID  string
	Moniker string

	Amount string

	PubKey cryptotypes.PubKey

	IP              string
	Website         string
	SecurityContact string
	Details         string
	Identity        string
}

func PrepareConfigForTxCreateValidator(flagSet *flag.FlagSet, moniker, nodeID, chainID string, valPubKey cryptotypes.PubKey) (TxCreateValidatorConfig, error) {
	c := TxCreateValidatorConfig{}

	ip, err := flagSet.GetString(stakingcli.FlagIP)
	if err != nil {
		return c, err
	}
	if ip == "" {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't retrieve an external IP; "+
			"the tx's memo field will be unset")
	}
	c.IP = ip

	website, err := flagSet.GetString(FlagWebsite)
	if err != nil {
		return c, err
	}
	c.Website = website

	securityContact, err := flagSet.GetString(FlagSecurityContact)
	if err != nil {
		return c, err
	}
	c.SecurityContact = securityContact

	details, err := flagSet.GetString(FlagDetails)
	if err != nil {
		return c, err
	}
	c.SecurityContact = details

	identity, err := flagSet.GetString(FlagIdentity)
	if err != nil {
		return c, err
	}
	c.Identity = identity

	c.Amount, err = flagSet.GetString(FlagAmount)
	if err != nil {
		return c, err
	}

	c.NodeID = nodeID
	c.PubKey = valPubKey
	c.Website = website
	c.SecurityContact = securityContact
	c.Details = details
	c.Identity = identity
	c.ChainID = chainID
	c.Moniker = moniker

	if c.Amount == "" {
		c.Amount = defaultAmount
	}

	return c, nil
}

// BuildCreateValidatorMsg makes a new MsgCreateValidator.
func BuildCreateValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool) (tx.Factory, sdk.Msg, error) {
	amounstStr := config.Amount
	amount, err := sdk.ParseCoinNormalized(amounstStr)

	if err != nil {
		return txBldr, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	description := stakingtypes.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	msg, err := types.NewMsgCreateValidator(valAddr, config.PubKey, amount, description)
	if err != nil {
		return txBldr, msg, err
	}
	if generateOnly {
		ip := config.IP
		nodeID := config.NodeID

		if nodeID != "" && ip != "" {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:26656", nodeID, ip))
		}
	}

	return txBldr, msg, nil
}

func NewDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self-delegate <liquid-amount> <vesting-amount>",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid and illiquid tokens to a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Delegate an amount of liquid and/or illiquid (vesting) coins to a validator from your wallet.

Examples:
$ %s tx poe self-delegate 1000stake 0stake --from mykey
$ %s tx poe self-delegate 500stake 500stake --from mykey
`,
				version.AppName,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			vesting_amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			delAddr := clientCtx.GetFromAddress()
			msg := types.NewMsgDelegate(delAddr, amount, vesting_amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewUnbondCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond [amount]",
		Short: "Unbond shares from a validator",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unbond an amount of bonded shares from a validator.

Example:
$ %s tx poe unbond 100stake --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgUndelegate(delAddr, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewUnjailTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Args:  cobra.NoArgs,
		Short: "unjail validator previously jailed for downtime",
		Long: fmt.Sprintf(`unjail a jailed validator:

$ %s tx poe unjail --from mykey
`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ContractAddress(cmd.Context(), &types.QueryContractAddressRequest{ContractType: types.PoEContractTypeValset})
			if err != nil {
				return errors.Wrap(err, "query valset contract address")
			}
			nodeOperator := clientCtx.GetFromAddress()
			unjailMsg := &poecontracts.TG4ValsetExecute{
				Unjail: &poecontracts.UnjailMsg{},
			}
			unjailBz, err := json.Marshal(unjailMsg)
			if err != nil {
				return errors.Wrap(err, "encode msg payload")
			}

			msg := &wasmtypes.MsgExecuteContract{
				Sender:   nodeOperator.String(),
				Contract: res.Address,
				Msg:      unjailBz,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
