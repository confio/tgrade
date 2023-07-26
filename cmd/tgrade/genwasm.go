package main

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/confio/tgrade/x/twasm/client/cli"
)

func GenesisWasmFlagsCmd(defaultNodeHome string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "wasm-genesis-flags",
		Short:                      "Wasm genesis flag subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	genIO := cli.NewGenesisIO()
	txCmd.AddCommand(
		cli.GenesisSetPrivileged(defaultNodeHome, genIO),
		cli.GenesisSetPinned(defaultNodeHome, genIO),
	)
	return txCmd
}
