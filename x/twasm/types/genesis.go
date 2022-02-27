package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (g GenesisState) ValidateBasic() error {
	if err := g.Wasm.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "wasm")
	}
	const tgradeExtType = "confio.twasm.v1beta1.TgradeContractDetails"
	for _, c := range g.Wasm.Contracts {
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

	genesisCodes, err := getAllCodes(&g.Wasm)
	if err != nil {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalid, "genesis codes: %s", err.Error())
	}
	for _, code := range genesisCodes {
		delete(uniquePinnedCodeIDs, code.CodeID)
	}
	if len(uniquePinnedCodeIDs) > 0 {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "%d pinned codeIDs not found in genesis codeIDs", len(uniquePinnedCodeIDs))
	}

	genesisContracts := getAllContracts(&g.Wasm)
	for _, contract := range genesisContracts {
		delete(uniqueAddr, contract.ContractAddress)
	}
	if len(uniqueAddr) > 0 {
		return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "%d privileged contract addresses not found in genesis contract addresses", len(uniqueAddr))
	}

	return nil
}
