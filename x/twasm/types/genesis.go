package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (g GenesisState) ValidateBasic() error {
	if err := g.Wasm.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "wasm")
	}
	return nil
}
