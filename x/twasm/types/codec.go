package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	wasmtypes.RegisterLegacyAminoCodec(cdc)
	cdc.RegisterConcrete(&PromoteToPrivilegedContractProposal{}, "twasm/PromoteToPrivilegedContractProposal", nil)
	cdc.RegisterConcrete(&DemotePrivilegedContractProposal{}, "twasm/DemotePrivilegedContractProposal", nil)
	cdc.RegisterConcrete(&TgradeContractDetails{}, "twasm/TgradeContractDetails", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	wasmtypes.RegisterInterfaces(registry)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&PromoteToPrivilegedContractProposal{},
		&DemotePrivilegedContractProposal{},
	)
	registry.RegisterImplementations(
		(*wasmtypes.ContractInfoExtension)(nil),
		&TgradeContractDetails{},
	)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/wasm module codec.

	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
