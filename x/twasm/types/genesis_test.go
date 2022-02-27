package types

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"

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
		"unique pinned codeIDs": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PinnedCodeIDs = []uint64{1, 2, 3}
				state.Wasm.Codes = []types.Code{newCode(1), newCode(2), newCode(3), newCode(4)}
			}),
			expErr: false,
		},
		"duplicate pinned codeIDs": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PinnedCodeIDs = []uint64{1, 2, 3, 3}
				state.Wasm.Codes = []types.Code{newCode(1), newCode(2), newCode(3), newCode(4)}
			}),
			expErr: true,
		},
		"pinned codeIDs do not exist in genesis codeIDs": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				state.PinnedCodeIDs = []uint64{1, 2, 3}
				state.Wasm.Codes = []types.Code{newCode(1), newCode(2), newCode(4), newCode(5)}
			}),
			expErr: true,
		},
		"privileged contract addresses  exist in genesis contract addresses": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				addresses := []string{RandomBech32Address(t), RandomBech32Address(t), RandomBech32Address(t)}
				state.PrivilegedContractAddresses = addresses
				state.Wasm.Contracts = []types.Contract{newContract(addresses[0]), newContract(addresses[1]), newContract(addresses[2])}
			}),
			expErr: false,
		},
		"privileged contract addresses  do not exist in genesis contract addresses": {
			state: GenesisStateFixture(t, func(state *GenesisState) {
				addresses := []string{RandomBech32Address(t), RandomBech32Address(t), RandomBech32Address(t)}
				state.PrivilegedContractAddresses = addresses
				state.Wasm.Contracts = []types.Contract{newContract(addresses[0]), newContract(addresses[1]), newContract(RandomBech32Address(t))}
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

// newCode returns Code with custom codeID
func newCode(codeID uint64) types.Code {
	code := types.CodeFixture(func(c *types.Code) {
		c.CodeID = codeID
	})
	return code
}

// newContract returns Contract with custom address
func newContract(addr string) types.Contract {
	contract := types.ContractFixture(func(c *types.Contract) {
		c.ContractAddress = addr
	})
	return contract
}
