package main

import (
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	"github.com/confio/tgrade/x/twasm/client/cli"
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
