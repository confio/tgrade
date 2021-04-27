package cli

import (
	"encoding/json"
	"fmt"
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/server"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
)

var _ wasmcli.GenesisMutator = GenesisIO{}

// GenesisIO to alter the genesis state for this module. To be used with the wasm cli extension point.
type GenesisIO struct {
	GenesisReader
}

func NewGenesisIO() *GenesisIO {
	return &GenesisIO{GenesisReader: GenesisReader{}}
}

// AlterWasmModuleState loads the genesis from the default or set home dir,
// unmarshalls the wasm module section into the object representation
// calls the callback function to modify it
// and marshals the modified state back into the genesis file
func (x GenesisIO) AlterWasmModuleState(cmd *cobra.Command, callback func(state *wasmtypes.GenesisState, appState map[string]json.RawMessage) error) error {
	return x.AlterTWasmModuleState(cmd, func(state *types.GenesisState, appState map[string]json.RawMessage) error {
		return callback(&state.Wasm, appState)
	})
}

// AlterTWasmModuleState loads the genesis from the default or set home dir,
// unmarshalls the twasm module section into the object representation
// calls the callback function to modify it
// and marshals the modified state back into the genesis file
func (x GenesisIO) AlterTWasmModuleState(cmd *cobra.Command, callback func(state *types.GenesisState, appState map[string]json.RawMessage) error) error {
	g, err := x.ReadTWasmGenesis(cmd)
	if err != nil {
		return err
	}
	if err := callback(&g.twasmModuleState, g.AppState); err != nil {
		return err
	}
	// and store update
	if err := g.twasmModuleState.ValidateBasic(); err != nil {
		return err
	}
	clientCtx := client.GetClientContextFromCmd(cmd)

	twasmGenStateBz, err := clientCtx.JSONMarshaler.MarshalJSON(&g.twasmModuleState)
	if err != nil {
		return sdkerrors.Wrap(err, "marshal twasm genesis state")
	}

	g.AppState[types.ModuleName] = twasmGenStateBz
	appStateJSON, err := json.Marshal(g.AppState)
	if err != nil {
		return sdkerrors.Wrap(err, "marshal application genesis state")
	}

	g.GenDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(g.GenDoc, g.GenesisFile)
}

// TWasmGenesisData extends the wasmcli GenesisData for this module state.
type TWasmGenesisData struct {
	*wasmcli.GenesisData
	twasmModuleState types.GenesisState
}

var _ wasmcli.GenesisReader = GenesisReader{}

// GenesisReader reads the genesis data for this module. To be used with the wasm cli extension point
type GenesisReader struct{}

func (d GenesisReader) ReadTWasmGenesis(cmd *cobra.Command) (*TWasmGenesisData, error) {
	clientCtx := client.GetClientContextFromCmd(cmd)
	serverCtx := server.GetServerContextFromCmd(cmd)
	config := serverCtx.Config
	config.SetRoot(clientCtx.HomeDir)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}
	var twasmGenesisState types.GenesisState
	if appState[types.ModuleName] != nil {
		clientCtx := client.GetClientContextFromCmd(cmd)
		clientCtx.JSONMarshaler.MustUnmarshalJSON(appState[types.ModuleName], &twasmGenesisState)
	}
	return &TWasmGenesisData{
		GenesisData: wasmcli.NewGenesisData(
			genFile,
			genDoc,
			appState,
			&twasmGenesisState.Wasm,
		),
		twasmModuleState: twasmGenesisState,
	}, nil
}

func (d GenesisReader) ReadWasmGenesis(cmd *cobra.Command) (*wasmcli.GenesisData, error) {
	r, err := d.ReadTWasmGenesis(cmd)
	if err != nil {
		return nil, err
	}
	return r.GenesisData, nil
}
