package types

import (
	"testing"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisValidate(t *testing.T) {
	specs := map[string]struct {
		state  GenesisState
		expErr bool
	}{
		"all good": {
			state: GenesisStateFixture(t),
		},
		"wasm invalid": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.Wasm.Codes[0].CodeID = 0
			}),
			expErr: true,
		},
		"privileged address empty": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PrivilegedContractAddresses = []string{""}
			}),
			expErr: true,
		},
		"privileged address invalid": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PrivilegedContractAddresses = []string{"invalid"}
			}),
			expErr: true,
		},
		"duplicate privileged contract address": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PrivilegedContractAddresses = append(state.PrivilegedContractAddresses, state.PrivilegedContractAddresses[0])
			}),
			expErr: true,
		},
		"invalid extension": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				var invalidType govtypes.Proposal // any protobuf type
				err := state.Wasm.Contracts[0].ContractInfo.SetExtension(&invalidType)
				require.NoError(t, err)
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.state.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
