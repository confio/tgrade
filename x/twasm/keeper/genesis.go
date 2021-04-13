package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis sets supply information for genesis.
//
// CONTRACT: all types of accounts must have been already initialized/created
func InitGenesis(ctx sdk.Context, keeper *Keeper, data types.GenesisState, stakingKeeper wasmkeeper.ValidatorSetSource, msgHandler sdk.Handler) ([]abci.ValidatorUpdate, error) {
	result, err := wasmkeeper.InitGenesis(ctx, &keeper.Keeper, data.Wasm, stakingKeeper, msgHandler)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "wasm")
	}
	for i, a := range data.PrivilegedContractAddresses {
		addr, err := sdk.AccAddressFromBech32(a)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "privileged contract: %d", i)
		}
		keeper.SetPrivileged(ctx, addr)
	}
	return result, nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper *Keeper) *types.GenesisState {
	var genState types.GenesisState
	genState.Wasm = *wasmkeeper.ExportGenesis(ctx, &keeper.Keeper)
	keeper.IteratePrivileged(ctx, func(contract sdk.AccAddress) bool {
		genState.PrivilegedContractAddresses = append(genState.PrivilegedContractAddresses, contract.String())
		return false
	})
	return &genState
}
