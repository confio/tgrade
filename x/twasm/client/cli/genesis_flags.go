package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
)

// GenesisSetPrivileged cli command to enable privileges for a contract in the genesis
func GenesisSetPrivileged(defaultNodeHome string, genesisMutator *GenesisIO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-privileged [contract_addr_bech32]",
		Short: "Set privileged flag for contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := sdk.AccAddressFromBech32(args[0]); err != nil {
				return sdkerrors.Wrap(err, "contract address")
			}
			return genesisMutator.AlterTWasmModuleState(cmd, func(state *types.GenesisState, appState map[string]json.RawMessage) error {
				for _, v := range state.PrivilegedContractAddresses {
					if v == args[0] {
						return sdkerrors.Wrap(wasmtypes.ErrDuplicate, "contract address already privileged")
					}
				}
				state.PrivilegedContractAddresses = append(state.PrivilegedContractAddresses, args[0])
				return nil
			})
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	return cmd
}

// GenesisSetPinned returns a cli command to pin a contract into VM cache.
func GenesisSetPinned(defaultNodeHome string, genesisMutator *GenesisIO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-pinned [code_id]",
		Short: "Set pinned flag for WASM code to be permanent in cache",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("code ID is not a valid number: %w", err)
			}
			return genesisMutator.AlterTWasmModuleState(cmd, func(state *types.GenesisState, appState map[string]json.RawMessage) error {
				for _, pinned := range state.PinnedCodeIDs {
					if pinned == codeID {
						return sdkerrors.Wrap(wasmtypes.ErrDuplicate, "code already pinned")
					}
				}
				state.PinnedCodeIDs = append(state.PinnedCodeIDs, codeID)
				return nil
			})
		},
	}
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	return cmd
}
