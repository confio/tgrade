package twasm

import (
	"encoding/json"
	"fmt"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/keeper"
	"github.com/confio/tgrade/x/twasm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	"runtime/debug"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Sudoer used in abci
type Sudoer interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
	IterateContractCallbacksByType(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func BeginBlocker(parentCtx sdk.Context, k Sudoer, b abci.RequestBeginBlock) {
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
	k.IterateContractCallbacksByType(parentCtx, types.CallbackTypeBeginBlock, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Debug("privileged contract callback", "type", "begin-block", "msg", string(msgBz))
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

func EndBlocker(parentCtx sdk.Context, k Sudoer) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	sudoMsg := contract.TgradeSudoMsg{EndBlock: &struct{}{}}
	msgBz, err := json.Marshal(sudoMsg)
	if err != nil {
		panic(err) // this will break consensus
	}
	logger := keeper.ModuleLogger(parentCtx)
	k.IterateContractCallbacksByType(parentCtx, types.CallbackTypeEndBlock, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Debug("privileged contract callback", "type", "end-block", "msg", string(msgBz))
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

	sudoMsg = contract.TgradeSudoMsg{EndWithValidatorUpdate: &struct{}{}}
	msgBz, err = json.Marshal(sudoMsg)
	if err != nil {
		panic(err) // this will break consensus
	}

	var result []abci.ValidatorUpdate
	// allow validator set updates for this group only
	k.IterateContractCallbacksByType(parentCtx, types.CallbackTypeValidatorSetUpdate, func(pos uint8, contractAddr sdk.AccAddress) bool {
		logger.Info("privileged contract callback", "type", "validator-set-update", "msg", string(msgBz))
		ctx, commit := parentCtx.CacheContext()

		result, err = callValidatorSetUpdaterContract(contractAddr, k, ctx, msgBz, logger)
		if err != nil {
			logger.Error("validator set update failed",
				"cause", err, "contract-address", contractAddr, "position", pos,
			)
			panic(err) // this breaks consensus
		}
		commit()
		return true // stop at first contract
	})
	return result
}

func callValidatorSetUpdaterContract(contractAddr sdk.AccAddress, k Sudoer, ctx sdk.Context, msgBz []byte, logger log.Logger) ([]abci.ValidatorUpdate, error) {
	resp, err := k.Sudo(ctx, contractAddr, msgBz)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sudo")
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	var contractResult contract.EndWithValidatorUpdateResponse
	if err := json.Unmarshal(resp.Data, &contractResult); err != nil {
		return nil, sdkerrors.Wrap(err, "contract response")
	}
	if len(contractResult.Diffs) == 0 {
		return nil, nil
	}

	result := make([]abci.ValidatorUpdate, len(contractResult.Diffs))
	for i, v := range contractResult.Diffs {
		result[i] = abci.ValidatorUpdate{
			PubKey: getPubKey(v.PubKey),
			Power:  int64(v.Power),
		}
	}
	logger.Info("privileged contract callback", "type", "validator-set-update", "result", result)
	return result, nil
}

func getPubKey(key []byte) crypto.PublicKey {
	return crypto.PublicKey{
		Sum: &crypto.PublicKey_Ed25519{
			Ed25519: key,
		},
	}
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
				Error("panic in begin block callback",
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
