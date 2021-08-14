package types

import (
	"errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"
)

const DefaultBondDenom = "utgd"

// DefaultGenesisState default values
func DefaultGenesisState() GenesisState {
	return GenesisState{
		SeedContracts:      true,
		BondDenom:          DefaultBondDenom,
		SystemAdminAddress: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
		Params:             DefaultParams(),
	}
}

func ValidateGenesis(g GenesisState, txJSONDecoder sdk.TxDecoder) error {
	if err := sdk.ValidateDenom(g.BondDenom); err != nil {
		return sdkerrors.Wrap(err, "bond denom")
	}

	if g.SeedContracts {
		if len(g.Contracts) != 0 {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "seed enabled but PoE contracts addresses provided")
		}
		if len(g.Engagement) == 0 {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty engagement group")
		}
	} else {
		return errors.New("not supported, yet")
		//if len(g.Contracts) == 0 {
		//	return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "seed disabled but no PoE contract addresses provided")
		//}
		// todo (Alex): if we preserve state in the engagement contract then we need to ensure that there are no
		// new members in the engagement group
		// if we can reset state then the engagement group must not be empty
		//if len(g.Engagement) != 0 {
		//	return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "engagement group set")
		//}
		//uniqueContractTypes := make(map[PoEContractType]struct{}, len(g.Contracts))
		//for i, v := range g.Contracts {
		//	if err := v.ValidateBasic(); err != nil {
		//		return sdkerrors.Wrapf(err, "contract %d", i)
		//	}
		//	if _, exists := uniqueContractTypes[v.ContractType]; exists {
		//		return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "contract type %s", v.ContractType.String())
		//	}
		//}
		//if len(uniqueContractTypes) != len(PoEContractType_name)-1 {
		//	return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "PoE contract(s) missing")
		//}
	}
	if _, err := sdk.AccAddressFromBech32(g.SystemAdminAddress); err != nil {
		return sdkerrors.Wrap(err, "system admin address")
	}

	uniqueEngagementMembers := make(map[string]struct{}, len(g.Engagement))
	for i, v := range g.Engagement {
		if err := v.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "contract %d", i)
		}
		if _, exists := uniqueEngagementMembers[v.Address]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrDuplicate, "member: %s", v.Address)
		}
		uniqueEngagementMembers[v.Address] = struct{}{}
	}

	uniqueOperators := make(map[string]struct{}, len(g.GenTxs))
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
		if _, ok := uniqueEngagementMembers[msg.DelegatorAddress]; !ok {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "gen tx delegator not in engagement group: %q, gentx: %d", msg.DelegatorAddress, i)
		}
		if _, exists := uniqueOperators[msg.DelegatorAddress]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "gen tx delegator used already with another gen tx: %q, gentx: %d", msg.DelegatorAddress, i)
		}
		uniqueOperators[msg.DelegatorAddress] = struct{}{}
	}
	return nil
}

func (c PoEContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	return sdkerrors.Wrap(c.ContractType.ValidateBasic(), "contract type")
}

func (c TG4Member) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	if c.Weight == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "weight")
	}
	return nil
}
