package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func StargateContentProposalFixture(mutators ...func(proposal *StargateContentProposal)) *StargateContentProposal {
	anyProto, err := NewStargateContentProposal("nested", "proto", &govtypes.TextProposal{Title: "another nested", Description: "3rd level"})
	if err != nil {
		panic(err)
	}

	// new stargate with a protobuf type that implements govtypes.Content and has another Any
	p, err := NewStargateContentProposal("foo", "bar", anyProto)
	if err != nil {
		panic(err)
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
		RegisteredCallbacks: []*RegisteredCallback{{
			Position:     1,
			CallbackType: "begin_block",
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
