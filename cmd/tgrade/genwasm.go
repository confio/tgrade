package main

import (
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func AddGenesisWasmMsgCmd(defaultNodeHome string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "add-wasm-genesis-message",
		Short:                      "Wasm genesis subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		wasmcli.GenesisStoreCodeCmd(defaultNodeHome),
		wasmcli.GenesisInstantiateContractCmd(defaultNodeHome),
		wasmcli.GenesisExecuteContractCmd(defaultNodeHome),
		wasmcli.GenesisListContractsCmd(defaultNodeHome),
		wasmcli.GenesisListCodesCmd(defaultNodeHome),
	)
	return txCmd

}
