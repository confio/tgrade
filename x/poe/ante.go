package poe

import (
	"fmt"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type bankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Calls next AnteHandler on success
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	bankKeeper     bankKeeper
	contractSource keeper.ContractSource
}

func NewDeductFeeDecorator(bk bankKeeper, cs keeper.ContractSource) DeductFeeDecorator {
	return DeductFeeDecorator{
		bankKeeper:     bk,
		contractSource: cs,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feePayer := feeTx.FeePayer()
	feeCollector, err := dfd.contractSource.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		panic(fmt.Sprintf("%s contract address has not been set", types.PoEContractTypeValset))
	}

	if fee := feeTx.GetFee(); !fee.IsZero() {
		// deduct the fees
		if !fee.IsValid() {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", fee)
		}
		if err := dfd.bankKeeper.SendCoins(ctx, feePayer, feeCollector, fee); err != nil {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	return next(ctx, tx, simulate)
}
