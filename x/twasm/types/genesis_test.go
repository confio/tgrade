package types

import (
	"github.com/stretchr/testify/require"
	"testing"
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
