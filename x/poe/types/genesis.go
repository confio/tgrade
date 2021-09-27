package types

import (
	"errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"
	"time"
)

const DefaultBondDenom = "utgd"

// DefaultGenesisState default values
func DefaultGenesisState() GenesisState {
	return GenesisState{
		SeedContracts: true,
		BondDenom:     DefaultBondDenom,
		StakeContractConfig: &StakeContractConfig{
			MinBond:              1,
			TokensPerWeight:      1,
			UnbondingPeriod:      time.Hour * 21 * 24,
			ClaimAutoreturnLimit: 20,
			PreAuths:             1,
		},
		ValsetContractConfig: &ValsetContractConfig{
			MinWeight:     1,
			MaxValidators: 100,
			EpochLength:   60 * time.Second,
			EpochReward:   sdk.NewCoin(DefaultBondDenom, sdk.NewInt(100_000)),
			Scaling:       1,
			FeePercentage: sdk.NewDec(50),
		},
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
		if g.ValsetContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty valset contract config")
		}
		if err := g.ValsetContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "valset contract config")
		}
		if g.StakeContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty stake contract config")
		}
		if err := g.StakeContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "stake contract config")
		}
		if g.ValsetContractConfig.EpochReward.Denom != g.BondDenom {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "rewards not in bonded denom")
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
		if _, ok := uniqueEngagementMembers[msg.OperatorAddress]; !ok {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "gen tx delegator not in engagement group: %q, gentx: %d", msg.OperatorAddress, i)
		}
		if _, exists := uniqueOperators[msg.OperatorAddress]; exists {
			return sdkerrors.Wrapf(wasmtypes.ErrInvalidGenesis, "gen tx delegator used already with another gen tx: %q, gentx: %d", msg.OperatorAddress, i)
		}
		uniqueOperators[msg.OperatorAddress] = struct{}{}
	}
	return nil
}

// ValidateBasic ensure basic constraints
func (c ValsetContractConfig) ValidateBasic() error {
	if c.MaxValidators == 0 {
		return sdkerrors.Wrap(ErrEmpty, "max validators")
	}
	if c.EpochLength == 0 {
		return sdkerrors.Wrap(ErrEmpty, "epoch length")
	}
	if c.Scaling == 0 {
		return sdkerrors.Wrap(ErrEmpty, "scaling")
	}

	minFeePercentage := sdk.NewDecFromIntWithPrec(sdk.OneInt(), 16)
	if c.FeePercentage.LT(minFeePercentage) {
		return sdkerrors.Wrap(ErrEmpty, "fee percentage")
	}
	return nil
}

// ValidateBasic ensure basic constraints
func (c StakeContractConfig) ValidateBasic() error {
	if c.MinBond == 0 {
		return sdkerrors.Wrap(ErrEmpty, "min bond")
	}
	if c.TokensPerWeight == 0 {
		return sdkerrors.Wrap(ErrEmpty, "tokens per weight")
	}
	if c.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrEmpty, "unbonding period")
	}
	if time.Duration(uint64(c.UnbondingPeriod.Seconds()))*time.Second != c.UnbondingPeriod {
		return sdkerrors.Wrap(ErrInvalid, "unbonding period not convertable to seconds")
	}
	return nil
}

// ValidateBasic ensure basic constraints
func (c PoEContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	return sdkerrors.Wrap(c.ContractType.ValidateBasic(), "contract type")
}

// ValidateBasic ensure basic constraints
func (c TG4Member) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(c.Address); err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	if c.Weight == 0 {
		return sdkerrors.Wrap(wasmtypes.ErrInvalid, "weight")
	}
	return nil
}
