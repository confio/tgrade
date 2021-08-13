package poe

import (
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"time"
)

type abciKeeper interface {
	types.Sudoer
	IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

// EndBlocker calls the Valset contract for the validator diff.
func EndBlocker(parentCtx sdk.Context, k abciKeeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	logger := keeper.ModuleLogger(parentCtx)

	var diff []abci.ValidatorUpdate
	// allow validator set updates for this group only
	k.IteratePrivilegedContractsByType(parentCtx, twasmtypes.PrivilegeTypeValidatorSetUpdate, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Info("privileged contract callback", "type", twasmtypes.PrivilegeTypeValidatorSetUpdate.String())
		ctx, commit := parentCtx.CacheContext()
		var err error
		diff, err = contract.CallEndBlockWithValidatorUpdate(ctx, contractAddr, k)
		if err != nil {
			logger.Error("validator set update failed", "cause", err, "contract-address", contractAddr, "position", pos)
			panic(err) // this breaks consensus
		}
		commit()
		if len(diff) != 0 {
			logger.Info("update validator set", "new", diff)
		}
		return true // stop at first contract
	})
	return diff
}
