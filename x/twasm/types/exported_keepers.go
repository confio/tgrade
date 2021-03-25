package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Sudoer used in abci begin block
type Sudoer interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
	IterateContractCallbacksByType(ctx sdk.Context, callbackType PriviledgedCallbackType, cb func(prio uint64, contractAddr sdk.AccAddress) bool)
}
