package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"
)

// DefaultGenesisState default values
func DefaultGenesisState() GenesisState {
	return GenesisState{
		SeedContracts:      true,
		SystemAdminAddress: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
	}
}

func ValidateGenesis(g GenesisState, txJSONDecoder sdk.TxDecoder) error {
	if g.SeedContracts && len(g.Contracts) != 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "seed enabled but PoE contracts addresses provided")
	} else if !g.SeedContracts && len(g.Contracts) == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "seed disabled but no PoE contract addresses provided")
	}
	if len(g.Engagement) == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty engagement group")
	}
	if len(g.GenTxs) == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty gentx")
	}
	uniqueContractTypes := make(map[PoEContractTypes]struct{}, len(g.Contracts))
	for i, v := range g.Contracts {
		if err := v.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "contract %d", i)
		}
		if _, exists := uniqueContractTypes[v.ContractType]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "contract type %s", v.ContractType.String())
		}
	}
	if len(uniqueContractTypes) != len(PoEContractTypes_name)-1 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "PoE contract(s) missing")
	}
	if _, err := sdk.ValAddressFromBech32(g.SystemAdminAddress); err != nil {
		return sdkerrors.Wrap(err, "system admin address")
	}

	uniqueMembers := make(map[string]struct{}, len(g.Engagement))
	for i, v := range g.Engagement {
		if err := v.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "contract %d", i)
		}
		if _, exists := uniqueMembers[v.Address]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "member: %s", v.Address)
		}
	}

	for i, v := range g.GenTxs {
		genTx, err := txJSONDecoder(v)
		if err != nil {
			return sdkerrors.Wrapf(err, "gentx %d", i)
		}
		msgs := genTx.GetMsgs()
		if len(msgs) != 1 {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "tx with single message required")
		}
		msg := msgs[0].(*MsgCreateValidator)
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "gentx %d", i)
		}
		if _, ok := uniqueMembers[msg.DelegatorAddress]; !ok {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "gen tx delegator not in engagement group: %q, gentx: %d", msg.DelegatorAddress, i)
		}
	}
	return nil
}

func (c PoEContract) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	return sdkerrors.Wrap(c.ContractType.ValidateBasic(), "contract type")
}

func (t PoEContractTypes) ValidateBasic() error {
	if t == PoEContractTypes_UNDEFINED {
		return wasmtypes.ErrInvalid
	}
	if _, ok := PoEContractTypes_name[int32(t)]; !ok {
		return wasmtypes.ErrNotFound
	}
	return nil
}

func (c TG4Member) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	if c.Weight == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "weight")
	}
	return nil
}
