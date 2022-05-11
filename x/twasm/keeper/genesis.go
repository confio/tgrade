package keeper

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/twasm/contract"

	"github.com/confio/tgrade/x/twasm/types"
)

type noopValsetUpdater struct{}

func (n noopValsetUpdater) ApplyAndReturnValidatorSetUpdates(context sdk.Context) (updates []abci.ValidatorUpdate, err error) {
	return nil, nil
}

// InitGenesis sets supply information for genesis.
//
// CONTRACT: all types of accounts must have been already initialized/created
func InitGenesis(
	ctx sdk.Context,
	keeper *Keeper,
	data types.GenesisState,
	msgHandler sdk.Handler,
) ([]abci.ValidatorUpdate, error) {
	result, err := wasmkeeper.InitGenesis(ctx, &keeper.Keeper, data.RawWasmState(), noopValsetUpdater{}, msgHandler)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "wasm")
	}

	// import privileges from dumped contract infos
	for i, m := range data.Contracts {
		info := m.ContractInfo
		var d types.TgradeContractDetails
		if err := info.ReadExtension(&d); err != nil {
			return nil, sdkerrors.Wrapf(err, "extension contract: %d, %s", i, m.ContractAddress)
		}
		if len(d.RegisteredPrivileges) == 0 {
			continue // nothing to do
		}

		addr, err := sdk.AccAddressFromBech32(m.ContractAddress)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "contract: %d", i)
		}
		if err := keeper.importPrivileged(ctx, addr, info.CodeID, d); err != nil {
			return nil, sdkerrors.Wrapf(err, "privilege registration for contract: %s", m.ContractAddress)
		}

		// import custom state when privilege set
		if d.HasRegisteredPrivilege(types.PrivilegeStateExporterImporter) {
			model := m.GetCustomModel()
			if model == nil {
				return nil, sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "custom state model not set for %s", m.ContractAddress)
			}
			bz, err := json.Marshal(contract.TgradeSudoMsg{Import: &model.Msg})
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "marshal state import for %s", m.ContractAddress)
			}
			if _, err = keeper.Keeper.Sudo(ctx, addr, bz); err != nil {
				return nil, sdkerrors.Wrapf(err, "init custom state for %s", m.ContractAddress)
			}
		}
	}

	// set privileged
	for i, a := range data.PrivilegedContractAddresses {
		// handle new genesis message contract
		addr, err := sdk.AccAddressFromBech32(a)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "privileged contract: %d", i)
		}
		if err := keeper.SetPrivileged(ctx, addr); err != nil {
			return nil, sdkerrors.Wrapf(err, "set privileged flag for contract %s", a)
		}
	}

	// cache requested contracts
	for _, codeID := range data.PinnedCodeIDs {
		if err := keeper.contractKeeper.PinCode(ctx, codeID); err != nil {
			return nil, sdkerrors.Wrapf(err, "pin code with ID %d", codeID)
		}
	}
	return result, nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper *Keeper) *types.GenesisState {
	wasmState := wasmkeeper.ExportGenesis(ctx, &keeper.Keeper)
	contracts := make([]types.Contract, len(wasmState.Contracts))
	for i, v := range wasmState.Contracts {
		contracts[i] = types.Contract{
			ContractAddress: v.ContractAddress,
			ContractInfo:    v.ContractInfo,
		}

		var details types.TgradeContractDetails
		if err := v.ContractInfo.ReadExtension(&details); err != nil {
			panic(fmt.Sprintf("read contract info extension for %s", v.ContractAddress))
		}
		if !details.HasRegisteredPrivilege(types.PrivilegeStateExporterImporter) {
			contracts[i].ContractState = &types.Contract_KvModel{KvModel: &types.KVModel{Models: v.ContractState}}
			continue
		}
		c, err := sdk.AccAddressFromBech32(v.ContractAddress)
		if err != nil {
			panic(fmt.Sprintf("address %s: %s", v.ContractAddress, err))
		}
		bz, err := json.Marshal(contract.TgradeSudoMsg{Export: &struct{}{}})
		if err != nil {
			panic(sdkerrors.Wrapf(err, "marshal state export for %s", c.String()))
		}
		got, err := keeper.Keeper.Sudo(ctx, c, bz)
		if err != nil {
			panic(sdkerrors.Wrapf(err, "export custom state for %s with %q", c.String(), string(bz)))
		}
		contracts[i].ContractState = &types.Contract_CustomModel{CustomModel: &types.CustomModel{Msg: got}}
	}

	genState := types.GenesisState{
		Params:    wasmState.Params,
		Codes:     wasmState.Codes,
		Contracts: contracts,
		Sequences: wasmState.Sequences,
		GenMsgs:   wasmState.GenMsgs,
	}

	// pinned is stored in code info
	// privileges are stored contract info
	// no need to store them again in the genesis fields
	//
	// note: when a privileged contract address is added to the genesis.privileged_contract_addresses then
	// the sudo promote call is executed again so that the contract can register additional privileges
	return &genState
}
