package types

import (
	"bytes"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/tendermint/tendermint/libs/rand"
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

// DeterministicGenesisStateFixture is the same as GenesisStateFixture but with deterministic addresses and codes
func DeterministicGenesisStateFixture(t *testing.T, mutators ...func(*GenesisState)) GenesisState {
	genesisState := GenesisStateFixture(t)
	for i := range genesisState.Contracts {
		genesisState.Contracts[i].ContractAddress = wasmkeeper.BuildContractAddress(uint64(i), uint64(i)).String()
	}
	for i := range genesisState.Codes {
		wasmCode := bytes.Repeat([]byte{byte(i)}, 20)
		genesisState.Codes[i].CodeInfo = wasmtypes.CodeInfoFixture(wasmtypes.WithSHA256CodeHash(wasmCode))
		genesisState.Codes[i].CodeBytes = wasmCode
	}
	for _, m := range mutators {
		m(&genesisState)
	}
	return genesisState
}

// GenesisStateFixture test data fixture
func GenesisStateFixture(t *testing.T, mutators ...func(*GenesisState)) GenesisState {
	t.Helper()
	anyContractAddr := RandomBech32Address(t)
	wasmState := wasmtypes.GenesisFixture(func(state *wasmtypes.GenesisState) {
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
	})
	contracts := make([]Contract, len(wasmState.Contracts))
	for i, v := range wasmState.Contracts {
		contracts[i] = Contract{
			ContractAddress: v.ContractAddress,
			ContractInfo:    v.ContractInfo,
			ContractState:   &Contract_KvModel{&KVModel{v.ContractState}},
		}
	}
	genesisState := GenesisState{
		Params:                      wasmState.Params,
		Codes:                       wasmState.Codes,
		Contracts:                   contracts,
		Sequences:                   wasmState.Sequences,
		GenMsgs:                     wasmState.GenMsgs,
		PrivilegedContractAddresses: []string{anyContractAddr},
	}
	for _, m := range mutators {
		m(&genesisState)
	}
	return genesisState
}

// ContractFixture test data factory
func ContractFixture(t *testing.T, mutators ...func(contract *Contract)) Contract {
	t.Helper()
	wasmContract := wasmtypes.ContractFixture()
	c := Contract{
		ContractAddress: wasmContract.ContractAddress,
		ContractInfo:    wasmContract.ContractInfo,
		ContractState:   &Contract_KvModel{&KVModel{wasmContract.ContractState}},
	}
	for _, m := range mutators {
		m(&c)
	}
	return c
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
	return rand.Bytes(address.Len)
}

func RandomBech32Address(t *testing.T) string {
	return RandomAddress(t).String()
}
