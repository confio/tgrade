package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// govKeeper is a subset of Keeper that is needed for the gov proposal handling
type govKeeper interface {
	SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error
	UnsetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error
}

// NewProposalHandler creates a new governance Handler for wasm proposals
func NewProposalHandler(k Keeper) govtypes.Handler {
	wasmProposalHandler := wasmkeeper.NewWasmProposalHandler(k, wasmtypes.EnableAllProposals)
	return NewProposalHandlerX(k, wasmProposalHandler, k.govRouter)
}

// NewProposalHandlerX creates a new governance Handler for wasm proposals
func NewProposalHandlerX(k govKeeper, wasmProposalHandler govtypes.Handler, router govtypes.Router) govtypes.Handler {
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
		case *types.StargateContentProposal:
			nestedContent, ok := c.Content.GetCachedValue().(govtypes.Content)
			if !ok || nestedContent == nil {
				return sdkerrors.Wrap(wasmtypes.ErrInvalid, "not gov content type")
			}
			if !router.HasRoute(nestedContent.ProposalRoute()) {
				return sdkerrors.Wrap(govtypes.ErrNoProposalHandlerExists, nestedContent.ProposalRoute())
			}
			govHandler := router.GetRoute(nestedContent.ProposalRoute())
			return govHandler(ctx, content)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized twasm srcProposal content type: %T", c)
		}
	}
}

func handlePromoteContractProposal(ctx sdk.Context, k govKeeper, p types.PromoteToPrivilegedContractProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	contractAddr, err := sdk.AccAddressFromBech32(p.Contract)
	if err != nil {
		return sdkerrors.Wrap(err, "contract address")
	}

	return k.SetPrivileged(ctx, contractAddr)
}

func handleDemoteContractProposal(ctx sdk.Context, k govKeeper, p types.DemotePrivilegedContractProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}
	contractAddr, err := sdk.AccAddressFromBech32(p.Contract)
	if err != nil {
		return sdkerrors.Wrap(err, "contract address")
	}

	return k.UnsetPrivileged(ctx, contractAddr)
}
