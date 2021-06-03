package keeper

import (
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	marshaler codec.Marshaler
	storeKey  sdk.StoreKey
}

func NewKeeper(marshaler codec.Marshaler, key sdk.StoreKey) Keeper {
	return Keeper{marshaler: marshaler, storeKey: key}
}

func (k Keeper) SetPoeContractAddress(ctx sdk.Context, ctype types.PoEContractTypes, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(poeContractAddressKey(ctype), contractAddr.Bytes())

}

func (k Keeper) GetPoeContractAddress(ctx sdk.Context, ctype types.PoEContractTypes) (sdk.AccAddress, error) {
	if ctype == types.PoEContractTypes_UNDEFINED {
		return nil, sdkerrors.Wrap(wasmtypes.ErrInvalid, "contract type")
	}
	store := ctx.KVStore(k.storeKey)
	addr := store.Get(poeContractAddressKey(ctype))
	if len(addr) == 0 {
		return nil, wasmtypes.ErrNotFound
	}
	return addr, nil
}

func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func poeContractAddressKey(ctype types.PoEContractTypes) []byte {
	return append(types.PoEContractPrefix, []byte(ctype.String())...)

}
