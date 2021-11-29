package main

import (
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/confio/tgrade/x/twasm/client/cli"
)

func AddGenesisWasmMsgCmd(defaultNodeHome string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "wasm-genesis-message",
		Short:                      "Wasm genesis message subcommands",
		Aliases:                    []string{"wasm-genesis-msg", "wasm-genesis-messages", "add-wasm-genesis-message"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	genIO := cli.NewGenesisIO()
	txCmd.AddCommand(
		wasmcli.GenesisStoreCodeCmd(defaultNodeHome, genIO),
		wasmcli.GenesisInstantiateContractCmd(defaultNodeHome, genIO),
		wasmcli.GenesisExecuteContractCmd(defaultNodeHome, genIO),
		wasmcli.GenesisListContractsCmd(defaultNodeHome, genIO),
		wasmcli.GenesisListCodesCmd(defaultNodeHome, genIO),
	)
	return txCmd

}
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
