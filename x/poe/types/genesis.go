package types

import (
	"errors"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/rand"
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
		},
		ValsetContractConfig: &ValsetContractConfig{
			MinWeight:             1,
			MaxValidators:         100,
			EpochLength:           60 * time.Second,
			EpochReward:           sdk.NewCoin(DefaultBondDenom, sdk.NewInt(100_000)),
			Scaling:               1,
			FeePercentage:         sdk.NewDec(50),
			AutoUnjail:            false,
			ValidatorsRewardRatio: 50,
			DoubleSignSlashRatio:  sdk.NewDec(50),
		},
		EngagmentContractConfig: &EngagementContractConfig{
			Halflife: 180 * 24 * time.Hour,
		},
		OversightCommitteeContractConfig: &OversightCommitteeContractConfig{
			Name:         "Oversight Community",
			EscrowAmount: sdk.NewCoin(DefaultBondDenom, sdk.NewInt(1_000_000)),
			VotingRules: VotingRules{
				VotingPeriod:  30,
				Quorum:        sdk.NewDec(51),
				Threshold:     sdk.NewDec(55),
				AllowEndEarly: true,
			},
		},
		CommunityPoolContractConfig: &CommunityPoolContractConfig{
			VotingRules: VotingRules{
				VotingPeriod:  21,
				Quorum:        sdk.NewDec(10),
				Threshold:     sdk.NewDec(60),
				AllowEndEarly: true,
			},
		},
		ValidatorVotingContractConfig: &ValidatorVotingContractConfig{
			VotingRules: VotingRules{
				VotingPeriod:  14,
				Quorum:        sdk.NewDec(40),
				Threshold:     sdk.NewDec(66),
				AllowEndEarly: true,
			},
		},
		SystemAdminAddress: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
		Params:             DefaultParams(),
	}
}

func ValidateGenesis(g GenesisState, txJSONDecoder sdk.TxDecoder) error {
	if g.SeedContracts {
		if len(g.Contracts) != 0 {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "seed enabled but PoE contracts addresses provided")
		}
		if len(g.Engagement) == 0 {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty engagement group")
		}
		if g.EngagmentContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty engagement contract config")
		}
		if err := g.EngagmentContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "engagement contract config")
		}
		if err := sdk.ValidateDenom(g.BondDenom); err != nil {
			return sdkerrors.Wrap(err, "bond denom")
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

		if g.OversightCommitteeContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty oversight committee contract config")
		}
		if err := g.OversightCommitteeContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "oversight committee config")
		}
		if g.OversightCommitteeContractConfig.EscrowAmount.Denom != g.BondDenom {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "escrow not in bonded denom")
		}
		if g.CommunityPoolContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty community pool contract config")
		}
		if err := g.CommunityPoolContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "community pool config")
		}
		if g.ValidatorVotingContractConfig == nil {
			return sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty validator voting contract config")
		}
		if err := g.ValidatorVotingContractConfig.ValidateBasic(); err != nil {
			return sdkerrors.Wrap(err, "validator voting config")
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
func (c *EngagementContractConfig) ValidateBasic() error {
	if c.Halflife.Truncate(time.Second) != c.Halflife {
		return sdkerrors.Wrap(ErrInvalid, "halflife must not contain anything less than seconds")
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
	if c.ValidatorsRewardRatio > 100 {
		return sdkerrors.Wrap(ErrInvalid, "validator reward ratio must not be greater 100")
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

const (
	minNameLength   = 1
	maxNameLength   = 100
	minEscrowAmount = 999_999
)

// ValidateBasic ensure basic constraints
func (c OversightCommitteeContractConfig) ValidateBasic() error {
	if l := len(c.Name); l < minNameLength {
		return sdkerrors.Wrap(ErrEmpty, "name")
	} else if l > maxNameLength {
		return sdkerrors.Wrapf(ErrInvalid, "name length > %d", maxNameLength)
	}
	if c.EscrowAmount.Amount.LTE(sdk.NewInt(minEscrowAmount)) {
		return sdkerrors.Wrapf(ErrInvalid, "escrow amount must be greater %d", minEscrowAmount)
	}
	if err := c.VotingRules.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "voting rules")
	}
	if c.DenyListContractAddress != "" {
		if _, err := sdk.AccAddressFromBech32(c.DenyListContractAddress); err != nil {
			return sdkerrors.Wrap(ErrInvalid, "deny list contract address")
		}
	}
	return nil
}

// ValidateBasic ensure basic constraints
func (c CommunityPoolContractConfig) ValidateBasic() error {
	if err := c.VotingRules.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "voting rules")
	}
	return nil
}

// ValidateBasic ensure basic constraints
func (c ValidatorVotingContractConfig) ValidateBasic() error {
	if err := c.VotingRules.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "voting rules")
	}
	return nil
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

// ValidateBasic ensure basic constraints
func (v VotingRules) ValidateBasic() error {
	if v.VotingPeriod == 0 {
		return sdkerrors.Wrap(ErrEmpty, "voting period")
	}
	if v.Quorum.IsNil() || v.Quorum.LT(sdk.OneDec()) {
		return sdkerrors.Wrap(ErrInvalid, "quorum must be > 0")
	}
	if v.Quorum.GT(sdk.NewDec(100)) {
		return sdkerrors.Wrap(ErrInvalid, "quorum must be <=100")
	}
	if v.Threshold.IsNil() || v.Threshold.LT(sdk.NewDec(50)) {
		return sdkerrors.Wrap(ErrInvalid, "threshold must be => 50")
	}
	if v.Threshold.GT(sdk.NewDec(100)) {
		return sdkerrors.Wrap(ErrInvalid, "threshold must be <=100")
	}

	return nil
}
