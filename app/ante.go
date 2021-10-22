package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/confio/tgrade/x/globalfee"
	"github.com/confio/tgrade/x/poe"
	poekeeper "github.com/confio/tgrade/x/poe/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	channelkeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/keeper"
	ibcante "github.com/cosmos/cosmos-sdk/x/ibc/core/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(
	ak ante.AccountKeeper,
	bankKeeper bankkeeper.SendKeeper,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	txCounterStoreKey sdk.StoreKey,
	channelKeeper channelkeeper.Keeper,
	paramStore paramtypes.Subspace,
	contractSource poekeeper.ContractSource,
) sdk.AnteHandler {
	// initial version copied from sdk https://github.com/cosmos/cosmos-sdk/blob/v0.42.9/x/auth/ante/ante.go
	// globalfee was added and poe.NewDeductFeeDecorator replaces ante.NewDeductFeeDecorator
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewCountTXDecorator(txCounterStoreKey),
		ante.NewRejectExtensionOptionsDecorator(),
		ante.NewMempoolFeeDecorator(),
		globalfee.NewGlobalMinimumChainFeeDecorator(paramStore), // after local min fee check
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewRejectFeeGranterDecorator(),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		poe.NewDeductFeeDecorator(bankKeeper, contractSource),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
		ibcante.NewAnteDecorator(channelKeeper),
	)
}
