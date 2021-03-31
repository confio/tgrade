package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// todo (Alex): this comes with all proposals enabled
// NewProposalHandler creates a new governance Handler for wasm proposals
func NewProposalHandler(k Keeper) govtypes.Handler {
	wasmProposalHandler := wasmkeeper.NewWasmProposalHandler(k, wasmtypes.EnableAllProposals)
	return func(ctx sdk.Context, content govtypes.Content) error {
		err := wasmProposalHandler(ctx, content)
		switch {
		case err == nil:
			return nil
		case !sdkerrors.ErrUnknownRequest.Is(err):
			return err
		}
		if content == nil {
			return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "content must not be empty")
		}
		switch c := content.(type) {
		case *types.PromoteToPrivilegedContractProposal:
			return handlePromoteContractProposal(ctx, k, *c)
		case *types.DemotePrivilegedContractProposal:
			return handleDemoteContractProposal(ctx, k, *c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized twasm proposal content type: %T", c)
		}
	}
}

func handlePromoteContractProposal(ctx sdk.Context, k Keeper, p types.PromoteToPrivilegedContractProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	contractAddr, err := sdk.AccAddressFromBech32(p.Contract)
	if err != nil {
		return sdkerrors.Wrap(err, "contract address")
	}
	// todo (reviewer) :  emit event ?
	return k.SetPrivileged(ctx, contractAddr)
}

func handleDemoteContractProposal(ctx sdk.Context, k Keeper, p types.DemotePrivilegedContractProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}
	contractAddr, err := sdk.AccAddressFromBech32(p.Contract)
	if err != nil {
		return sdkerrors.Wrap(err, "contract address")
	}

	// todo (reviewer) :  emit event ?
	return k.UnsetPrivileged(ctx, contractAddr)
}
