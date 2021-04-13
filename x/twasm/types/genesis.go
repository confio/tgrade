package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (g GenesisState) ValidateBasic() error {
	if err := g.Wasm.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "wasm")
	}
	for i, a := range g.PrivilegedContractAddresses {
		if _, err := sdk.AccAddressFromBech32(a); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "privileged contract [%d, %s]: %s", i, a, err.Error())
		}
	}
	return nil
}
