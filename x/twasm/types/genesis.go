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
	return nil
}
