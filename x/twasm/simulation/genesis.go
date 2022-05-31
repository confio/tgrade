package simulation

// DONTCOVER

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/confio/tgrade/x/twasm/types"
)

// RandomizedGenState  generates a random GenesisState for wasm
func RandomizedGenState(simstate *module.SimulationState) {
	twasmGenesis := types.GenesisState{
		Params:                      wasmtypes.DefaultParams(),
		Codes:                       nil,
		Contracts:                   nil,
		Sequences:                   nil,
		GenMsgs:                     nil,
		PrivilegedContractAddresses: nil,
		PinnedCodeIDs:               nil,
	}

	simstate.GenState[wasmtypes.ModuleName] = simstate.Cdc.MustMarshalJSON(&twasmGenesis)
}
