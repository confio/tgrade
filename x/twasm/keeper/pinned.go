package keeper

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SetPinned updates the VM cachce and adds requested contract to it.
func (k Keeper) SetPinned(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	contractInfo := k.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract address")
	}
	if err := k.contractKeeper.PinCode(ctx, contractInfo.CodeID); err != nil {
		return sdkerrors.Wrapf(err, "pin code")
	}
	return nil
}
