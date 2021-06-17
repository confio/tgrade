package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func PromoteProposalFixture(mutators ...func(*PromoteToPrivilegedContractProposal)) *PromoteToPrivilegedContractProposal {
	const anyAddress = "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du"
	p := &PromoteToPrivilegedContractProposal{
		Title:       "Foo",
		Description: "Bar",
		Contract:    anyAddress,
	}
	for _, m := range mutators {
		m(p)
	}
	return p
}

func DemoteProposalFixture(mutators ...func(proposal *DemotePrivilegedContractProposal)) *DemotePrivilegedContractProposal {
	const anyAddress = "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du"
	p := &DemotePrivilegedContractProposal{
		Title:       "Foo",
		Description: "Bar",
		Contract:    anyAddress,
	}
	for _, m := range mutators {
		m(p)
	}
	return p
}

func GenesisStateFixture(t *testing.T, mutators ...func(*GenesisState)) GenesisState {
	t.Helper()
	anyContractAddr := RandomBech32Address(t)
	genesisState := GenesisState{
		Wasm: wasmtypes.GenesisFixture(func(state *wasmtypes.GenesisState) {
			state.Codes[1] = wasmtypes.CodeFixture(func(code *wasmtypes.Code) {
				code.CodeID = 2
				code.Pinned = true
			})
			state.Contracts[1] = wasmtypes.ContractFixture(func(contract *wasmtypes.Contract) {
				contract.ContractAddress = anyContractAddr
				contract.ContractInfo.CodeID = 2
			})
			state.Sequences = []wasmtypes.Sequence{
				{IDKey: wasmtypes.KeyLastCodeID, Value: 10},
				{IDKey: wasmtypes.KeyLastInstanceID, Value: 11},
			}
			state.GenMsgs = nil
		}),
		PrivilegedContractAddresses: []string{anyContractAddr},
	}
	for _, m := range mutators {
		m(&genesisState)
	}
	return genesisState
}

func TgradeContractDetailsFixture(t *testing.T, mutators ...func(d *TgradeContractDetails)) TgradeContractDetails {
	t.Helper()
	d := TgradeContractDetails{
		RegisteredPrivileges: []RegisteredPrivilege{{
			Position:      1,
			PrivilegeType: "begin_blocker",
		}},
	}
	for _, m := range mutators {
		m(&d)
	}
	return d
}

func RandomAddress(_ *testing.T) sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

func RandomBech32Address(t *testing.T) string {
	return RandomAddress(t).String()
}
