package types

import (
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// defined via in protobuf package structure. Note the leading `/`
const tgradeExtType = "/confio.twasm.v1beta1.TgradeContractDetails"

func (g GenesisState) ValidateBasic() error {
	wasmState := g.RawWasmState()
	if err := wasmState.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "wasm")
	}
	for _, c := range wasmState.Contracts {
		if c.ContractInfo.Extension != nil {
			if tgradeExtType != c.ContractInfo.Extension.TypeUrl {
				return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "invalid extension type: %s, contract %s", c.ContractInfo.Extension.TypeUrl, c.ContractAddress)
			}
		}
	}

	uniqueAddr := make(map[string]struct{}, len(g.PrivilegedContractAddresses))
	for i, a := range g.PrivilegedContractAddresses {
		if _, err := sdk.AccAddressFromBech32(a); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "privileged contract [%d, %s]: %s", i, a, err.Error())
		}
		if _, exists := uniqueAddr[a]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "privileged contract %s", a)
		}
		uniqueAddr[a] = struct{}{}
	}

	uniquePinnedCodeIDs := make(map[uint64]struct{}, len(g.PinnedCodeIDs))
	for _, code := range g.PinnedCodeIDs {
		if _, exists := uniquePinnedCodeIDs[code]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "pinned codeID %d", code)
		}
		uniquePinnedCodeIDs[code] = struct{}{}
	}

	genesisCodes := wasmcli.GetAllCodes(&wasmState)
	for _, code := range genesisCodes {
		delete(uniquePinnedCodeIDs, code.CodeID)
	}
	if len(uniquePinnedCodeIDs) > 0 {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "%d pinned codeIDs not found in genesis codeIDs", len(uniquePinnedCodeIDs))
	}

	genesisContracts := wasmcli.GetAllContracts(&wasmState)
	for _, contract := range genesisContracts {
		delete(uniqueAddr, contract.ContractAddress)
	}
	if len(uniqueAddr) > 0 {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "%d privileged contract addresses not found in genesis contract addresses", len(uniqueAddr))
	}

	return nil
}

// RawWasmState convert to wasm genesis state for vanilla import.
// Custom data models for privileged contracts are not included
func (g GenesisState) RawWasmState() wasmtypes.GenesisState {
	contracts := make([]wasmtypes.Contract, len(g.Contracts))
	for i, v := range g.Contracts {
		var s []wasmtypes.Model
		if m := v.GetKvModel(); m != nil {
			s = m.Models
		}
		contracts[i] = wasmtypes.Contract{
			ContractAddress: v.ContractAddress,
			ContractInfo:    v.ContractInfo,
			ContractState:   s,
		}
	}
	return wasmtypes.GenesisState{
		Params:    g.Params,
		Codes:     g.Codes,
		Contracts: contracts,
		Sequences: g.Sequences,
		GenMsgs:   g.GenMsgs,
	}
}

var _ codectypes.UnpackInterfacesMessage = GenesisState{}

// UnpackInterfaces implements codectypes.UnpackInterfaces
func (g GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, v := range g.Contracts {
		if err := v.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

var _ codectypes.UnpackInterfacesMessage = &Contract{}

// UnpackInterfaces implements codectypes.UnpackInterfaces
func (m *Contract) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.ContractInfo.UnpackInterfaces(unpacker)
}
