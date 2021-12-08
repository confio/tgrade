package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateGenesis(t *testing.T) {
	var anyAccAddr = RandomAccAddress()
	myGenTx, myOperatorAddr, _ := RandomGenTX(t, 100)
	txConfig := MakeEncodingConfig(t).TxConfig
	specs := map[string]struct {
		source GenesisState
		expErr bool
	}{
		"all good": {
			source: GenesisStateFixture(),
		},
		"seed with empty engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.Engagement = []TG4Member{}
			}),
			expErr: true,
		},
		"seed with duplicates in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.Engagement = []TG4Member{
					{Address: anyAccAddr.String(), Weight: 1},
					{Address: anyAccAddr.String(), Weight: 2},
				}
			}),
			expErr: true,
		},
		"seed with invalid addr in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.Engagement = []TG4Member{
					{Address: "invalid", Weight: 1},
				}
			}),
			expErr: true,
		},
		"seed with invalid weight in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.Engagement = []TG4Member{
					{Address: RandomAccAddress().String(), Weight: 0},
				}
			}),
			expErr: true,
		},
		"seed with legacy contracts": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.Contracts = []PoEContract{{ContractType: PoEContractTypeValset, Address: RandomAccAddress().String()}}
			}),
			expErr: true,
		},
		"empty bond denum": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.BondDenom = ""
			}),
			expErr: true,
		},
		"invalid bond denum": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.BondDenom = "&&&"
			}),
			expErr: true,
		},
		"empty system admin": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SystemAdminAddress = ""
			}),
			expErr: true,
		},
		"invalid system admin": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SystemAdminAddress = "invalid"
			}),
			expErr: true,
		},
		"valid gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GenTxs = []json.RawMessage{myGenTx}
				m.Engagement = append(m.Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Weight:  11,
				})
			}),
		},
		"duplicate gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GenTxs = []json.RawMessage{myGenTx, myGenTx}
				m.Engagement = append(m.Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Weight:  11,
				})
			}),
			expErr: true,
		},
		"validator not in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GenTxs = []json.RawMessage{myGenTx}
			}),
			expErr: true,
		},
		"invalid gentx json": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GenTxs = []json.RawMessage{[]byte("invalid")}
			}),
			expErr: true,
		},
		"engagement contract not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.EngagmentContractConfig = nil
			}),
			expErr: true,
		},
		"invalid engagement contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.EngagmentContractConfig.Halflife = time.Nanosecond
			}),
			expErr: true,
		},
		"valset contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.ValsetContractConfig = nil
			}),
			expErr: true,
		},
		"invalid valset contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.ValsetContractConfig.MaxValidators = 0
			}),
			expErr: true,
		},
		"invalid valset contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.ValsetContractConfig.EpochReward = sdk.NewCoin("alx", m.ValsetContractConfig.EpochReward.Amount)
			}),
			expErr: true,
		},
		"stake contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.StakeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid stake contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.StakeContractConfig.UnbondingPeriod = 0
			}),
			expErr: true,
		},
		"oversight committee contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid oversight committee contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Name = ""
			}),
			expErr: true,
		},
		"invalid oversight committee contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin("alx", m.OversightCommitteeContractConfig.EscrowAmount.Amount)
			}),
			expErr: true,
		},
		"community pool contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.CommunityPoolContractConfig = nil
			}),
			expErr: true,
		},
		"invalid community pool contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.CommunityPoolContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
		"validator voting contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.ValidatorVotingContractConfig = nil
			}),
			expErr: true,
		},
		"invalid validator voting contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.ValidatorVotingContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := ValidateGenesis(spec.source, txConfig.TxJSONDecoder())
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestValidateEngagementContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    *EngagementContractConfig
		expErr bool
	}{
		"default": {
			src: DefaultGenesisState().EngagmentContractConfig,
		},
		"halflife empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.EngagmentContractConfig.Halflife = 0
			}).EngagmentContractConfig,
		},
		"halflife contains elements < second": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.EngagmentContractConfig.Halflife = time.Minute + time.Millisecond
			}).EngagmentContractConfig,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestValidateValsetContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    ValsetContractConfig
		expErr bool
	}{
		"default": {
			src: *DefaultGenesisState().ValsetContractConfig,
		},
		"max validators empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.ValsetContractConfig.MaxValidators = 0 },
			).ValsetContractConfig,
			expErr: true,
		},
		"scaling empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.ValsetContractConfig.Scaling = 0 },
			).ValsetContractConfig,
			expErr: true,
		},
		"epoch length empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.ValsetContractConfig.EpochLength = 0 },
			).ValsetContractConfig,
			expErr: true,
		},
		"fee percentage at min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.0000000000000001"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.ValsetContractConfig.FeePercentage = val
				},
			).ValsetContractConfig,
		},
		"fee percentage below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.00000000000000009"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.ValsetContractConfig.FeePercentage = val
				},
			).ValsetContractConfig,
			expErr: true,
		},
		"validator rewards ratio zero": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("52.5")
				},
			).ValsetContractConfig,
		},
		"validator rewards ratio 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(100)
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).ValsetContractConfig,
		},
		"validator rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(101)
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).ValsetContractConfig,
			expErr: true,
		},
		"engagement rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.EngagementRewardRatio = sdk.NewDec(101)
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).ValsetContractConfig,
			expErr: true,
		},
		"community pool rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.NewDec(101)
					m.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("49")
					m.ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("49")
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("3")
				},
			).ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio < 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("10")
					m.ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("10")
					m.ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("10")
				},
			).ValsetContractConfig,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
func TestValidateStakeContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    StakeContractConfig
		expErr bool
	}{
		"default": {
			src: *DefaultGenesisState().StakeContractConfig,
		},
		"min bond empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.StakeContractConfig.MinBond = 0 },
			).StakeContractConfig,
			expErr: true,
		},
		"tokens per weight empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.StakeContractConfig.TokensPerWeight = 0 },
			).StakeContractConfig,
			expErr: true,
		},
		"unbonding period empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.StakeContractConfig.UnbondingPeriod = 0 },
			).StakeContractConfig,
			expErr: true,
		},
		"unbonding period below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.StakeContractConfig.UnbondingPeriod = time.Second - time.Nanosecond },
			).StakeContractConfig,
			expErr: true,
		},
		"not convertable to seconds": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.StakeContractConfig.UnbondingPeriod = time.Second + time.Nanosecond },
			).StakeContractConfig,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestValidateOversightCommitteeContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    OversightCommitteeContractConfig
		expErr bool
	}{
		"default": {
			src: *DefaultGenesisState().OversightCommitteeContractConfig,
		},
		"name empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Name = ""
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"name too long": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Name = strings.Repeat("a", 101)
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"escrow amount too low": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin(DefaultBondDenom, sdk.NewInt(999_999))
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"voting rules invalid": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"deny contract address not an address": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.DenyListContractAddress = "not-an-address"
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestValidateVotingRules(t *testing.T) {
	specs := map[string]struct {
		src    VotingRules
		expErr bool
	}{
		"default oc": {
			src: DefaultGenesisState().OversightCommitteeContractConfig.VotingRules,
		},
		"default community pool": {
			src: DefaultGenesisState().CommunityPoolContractConfig.VotingRules,
		},
		"default validator voting": {
			src: DefaultGenesisState().ValidatorVotingContractConfig.VotingRules,
		},
		"voting period empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.Dec{}
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.OneDec()
			}).OversightCommitteeContractConfig.VotingRules,
		},
		"quorum less than min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.ZeroDec()
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(100)
			}).OversightCommitteeContractConfig.VotingRules,
		},
		"quorum greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(101)
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.Dec{}
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(50)
			}).OversightCommitteeContractConfig.VotingRules,
		},
		"threshold lower min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(49)
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(100)
			}).OversightCommitteeContractConfig.VotingRules,
		},
		"threshold greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(101)
			}).OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {

			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
