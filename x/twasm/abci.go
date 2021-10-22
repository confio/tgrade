package twasm

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/keeper"
	"github.com/confio/tgrade/x/twasm/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type abciKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func BeginBlocker(parentCtx sdk.Context, k abciKeeper, b abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	evidence := make([]contract.Evidence, len(b.ByzantineValidators))
	for i, e := range b.ByzantineValidators {
		var et contract.EvidenceType
		switch e.Type {
		case abci.EvidenceType_DUPLICATE_VOTE:
			et = contract.EvidenceDuplicateVote
		case abci.EvidenceType_LIGHT_CLIENT_ATTACK:
			et = contract.EvidenceLightClientAttack
		default:
			panic(fmt.Sprintf("unsupported evidence type: %s", e.Type.String()))
		}

		evidence[i] = contract.Evidence{
			EvidenceType: et,
			Validator: contract.Validator{
				Address: e.Validator.Address,
				Power:   convUint64(e.Validator.Power),
			},
			Height:           convUint64(e.Height),
			Time:             convUint64(e.Time.Unix()),
			TotalVotingPower: convUint64(e.TotalVotingPower),
		}
	}
	msg := contract.TgradeSudoMsg{BeginBlock: &contract.BeginBlock{
		Evidence: evidence,
	}}

	msgBz, err := json.Marshal(msg)
	if err != nil {
		panic(err) // todo (reviewer): this will break consensus
	}
	logger := keeper.ModuleLogger(parentCtx)
	k.IteratePrivilegedContractsByType(parentCtx, types.PrivilegeTypeBeginBlock, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Debug("privileged contract callback", "type", types.PrivilegeTypeBeginBlock.String(), "msg", string(msgBz))
		ctx, commit := parentCtx.CacheContext()

		// any panic will crash the node so we are better taking care of them here
		defer recoverToLog(logger, contractAddr)()

		if _, err := k.Sudo(ctx, contractAddr, msgBz); err != nil {
			logger.Error("begin block contract failed",
				"cause", err, "contract-address", contractAddr, "position", pos,
			)
			return false
		}
		commit()
		return false
	})
}

func EndBlocker(parentCtx sdk.Context, k abciKeeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	sudoMsg := contract.TgradeSudoMsg{EndBlock: &struct{}{}}
	msgBz, err := json.Marshal(sudoMsg)
	if err != nil {
		panic(err) // this will break consensus
	}
	logger := keeper.ModuleLogger(parentCtx)
	k.IteratePrivilegedContractsByType(parentCtx, types.PrivilegeTypeEndBlock, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Debug("privileged contract callback", "type", types.PrivilegeTypeEndBlock.String(), "msg", string(msgBz))
		ctx, commit := parentCtx.CacheContext()

		// any panic will crash the node so we are better taking care of them here
		defer recoverToLog(logger, contractAddr)()

		if _, err := k.Sudo(ctx, contractAddr, msgBz); err != nil {
			logger.Error("end block contract failed",
				"cause", err, "contract-address", contractAddr, "position", pos,
			)
			return false
		}
		commit()
		return false
	})
	return nil
}

func recoverToLog(logger log.Logger, contractAddr sdk.AccAddress) func() {
	return func() {
		if r := recover(); r != nil {
			var cause string
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				cause = fmt.Sprintf("out of gas in location: %v", rType.Descriptor)
			default:
				cause = "unknown reason"
			}
			logger.
				Error("panic executing callback",
					"cause", cause,
					"contract-address", contractAddr.String(),
					"stacktrace", string(debug.Stack()),
				)
		}
	}
}

// convUint64 ensures source is not negative before type conversion
func convUint64(s int64) uint64 {
	if s < 0 {
		panic("must not be negative")
	}
	return uint64(s)
}
