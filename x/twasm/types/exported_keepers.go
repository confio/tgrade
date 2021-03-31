package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Sudoer used in abci begin block
type Sudoer interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
}
