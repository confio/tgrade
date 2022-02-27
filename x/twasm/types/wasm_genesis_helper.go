package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Copy pasted from Wasmd - https://github.com/CosmWasm/wasmd/blob/master/x/wasm/client/cli/genesis_msg.go

type codeMeta struct {
	CodeID uint64         `json:"code_id"`
	Info   types.CodeInfo `json:"info"`
}

func getAllCodes(state *types.GenesisState) ([]codeMeta, error) {
	all := make([]codeMeta, len(state.Codes))
	for i, c := range state.Codes {
		all[i] = codeMeta{
			CodeID: c.CodeID,
			Info:   c.CodeInfo,
		}
	}
	// add inflight
	seq := codeSeqValue(state)
	for _, m := range state.GenMsgs {
		if msg := m.GetStoreCode(); msg != nil {
			var accessConfig types.AccessConfig
			if msg.InstantiatePermission != nil {
				accessConfig = *msg.InstantiatePermission
			} else {
				// default
				creator, err := sdk.AccAddressFromBech32(msg.Sender)
				if err != nil {
					return nil, fmt.Errorf("sender: %s", err)
				}
				accessConfig = state.Params.InstantiateDefaultPermission.With(creator)
			}
			hash := sha256.Sum256(msg.WASMByteCode)
			all = append(all, codeMeta{
				CodeID: seq,
				Info: types.CodeInfo{
					CodeHash:          hash[:],
					Creator:           msg.Sender,
					InstantiateConfig: accessConfig,
				},
			})
			seq++
		}
	}
	return all, nil
}

type contractMeta struct {
	ContractAddress string             `json:"contract_address"`
	Info            types.ContractInfo `json:"info"`
}

func getAllContracts(state *types.GenesisState) []contractMeta {
	all := make([]contractMeta, len(state.Contracts))
	for i, c := range state.Contracts {
		all[i] = contractMeta{
			ContractAddress: c.ContractAddress,
			Info:            c.ContractInfo,
		}
	}
	// add inflight
	seq := contractSeqValue(state)
	for _, m := range state.GenMsgs {
		if msg := m.GetInstantiateContract(); msg != nil {
			all = append(all, contractMeta{
				ContractAddress: keeper.BuildContractAddress(msg.CodeID, seq).String(),
				Info: types.ContractInfo{
					CodeID:  msg.CodeID,
					Creator: msg.Sender,
					Admin:   msg.Admin,
					Label:   msg.Label,
				},
			})
			seq++
		}
	}
	return all
}

// contractSeqValue reads the contract sequence from the genesis or
// returns default start value used in the keeper
func contractSeqValue(state *types.GenesisState) uint64 {
	var seq uint64 = 1
	for _, s := range state.Sequences {
		if bytes.Equal(s.IDKey, types.KeyLastInstanceID) {
			seq = s.Value
			break
		}
	}
	return seq
}

// codeSeqValue reads the code sequence from the genesis or
// returns default start value used in the keeper
func codeSeqValue(state *types.GenesisState) uint64 {
	var seq uint64 = 1
	for _, s := range state.Sequences {
		if bytes.Equal(s.IDKey, types.KeyLastCodeID) {
			seq = s.Value
			break
		}
	}
	return seq
}
