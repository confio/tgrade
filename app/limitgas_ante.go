package app

import sdk "github.com/cosmos/cosmos-sdk/types"

type LimitSimulationGasDecorator struct {
	gasLimit *sdk.Gas
}

// NewLimitSimulationGasDecorator constructor accepts nil value to fallback to block gas limit
func NewLimitSimulationGasDecorator(gasLimit *sdk.Gas) *LimitSimulationGasDecorator {
	return &LimitSimulationGasDecorator{gasLimit: gasLimit}
}

func (d LimitSimulationGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !simulate {
		// Wasm code is not executed in checkTX so that we don't need to limit it further.
		// Tendermint rejects the TX afterwards when the tx.gas > max block gas.
		// On deliverTX we rely on the tendermint/sdk mechanics that ensure
		// tx has gas set and gas < max block gas
		return next(ctx, tx, simulate)
	}

	// apply custom node gas limit
	if d.gasLimit != nil {
		return next(ctx.WithGasMeter(sdk.NewGasMeter(*d.gasLimit)), tx, simulate)
	}

	// default to max block gas instead of infinite to be on the safe side
	if maxGas := ctx.ConsensusParams().GetBlock().MaxGas; maxGas > 0 {
		return next(ctx.WithGasMeter(sdk.NewGasMeter(sdk.Gas(maxGas))), tx, simulate)
	}
	return next(ctx, tx, simulate)
}
