package cli

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the PoE module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		GetCmdShowPoEContract(),
		GetCmdQueryValidators(),
		GetCmdQueryValidator(),
		GetCmdQueryUnbondingPeriod(),
	)
	return queryCmd
}

func GetCmdShowPoEContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contract-address <contract_type>",
		Short:   "Show contract address for given contract type",
		Long:    fmt.Sprintf("Show contract address for PoE type [%s]", allPoEContractTypes()),
		Aliases: []string{"ca"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			cbt := types.PoEContractTypeFrom(args[0])
			if cbt == types.PoEContractTypeUndefined {
				return fmt.Errorf("unknown contract type: %q", args[0])
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ContractAddress(
				cmd.Context(),
				&types.QueryContractAddressRequest{
					ContractType: cbt,
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryUnbondingPeriod() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-period",
		Short: "Query the global unbonding period",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.UnbondingPeriod(
				cmd.Context(),
				&types.QueryUnbondingPeriodRequest{},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all validators on a network.

Example:
$ %s query staking validators
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			_ = pageReq // todo (Alex): support pagination
			result, err := queryClient.Validators(cmd.Context(), &stakingtypes.QueryValidatorsRequest{
				// Leaving status empty on purpose to query all validators.
				//Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	// flags.AddPaginationFlagsToCmd(cmd, "validators")
	return cmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator [operator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query poe validator %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &stakingtypes.QueryValidatorRequest{ValidatorAddr: addr.String()}
			res, err := queryClient.Validator(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Validator)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func allPoEContractTypes() string {
	r := make([]string, 0, len(types.PoEContractType_name)-1)
	for _, v := range types.PoEContractType_name {
		if v == types.PoEContractTypeUndefined.String() {
			continue
		}
		r = append(r, v)
	}
	sort.Strings(r)
	return strings.Join(r, ", ")
}

// sdk ReadPageRequest expects binary but we encoded to base64 in our marshaller
func withPageKeyDecoded(flagSet *flag.FlagSet) *flag.FlagSet {
	encoded, err := flagSet.GetString(flags.FlagPageKey)
	if err != nil {
		panic(err.Error())
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err.Error())
	}
	if err := flagSet.Set(flags.FlagPageKey, string(raw)); err != nil {
		panic(err.Error())
	}
	return flagSet
}
