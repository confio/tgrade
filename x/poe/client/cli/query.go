package cli

import (
	"encoding/base64"
	"fmt"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"sort"
	"strings"
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
	flagSet.Set(flags.FlagPageKey, string(raw))
	return flagSet
}
