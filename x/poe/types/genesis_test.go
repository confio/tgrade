package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateGenesis(t *testing.T) {
	var anyAccAddr = RandomAccAddress()
	myGenTx, myOperatorAddr, myPubKey := RandomGenTX(t, 100)
	myPk, err := codectypes.NewAnyWithValue(myPubKey)
	require.NoError(t, err)
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
				m.SeedContracts.Engagement = []TG4Member{}
			}),
			expErr: true,
		},
		"seed with duplicates in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.Engagement = []TG4Member{
					{Address: anyAccAddr.String(), Points: 1},
					{Address: anyAccAddr.String(), Points: 2},
				}
			}),
			expErr: true,
		},
		"seed with invalid addr in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.Engagement = []TG4Member{
					{Address: "invalid", Points: 1},
				}
			}),
			expErr: true,
		},
		"seed with invalid weight in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.Engagement = []TG4Member{
					{Address: RandomAccAddress().String(), Points: 0},
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
				m.SeedContracts.BondDenom = ""
			}),
			expErr: true,
		},
		"invalid bond denum": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.BondDenom = "&&&"
			}),
			expErr: true,
		},
		"empty system admin": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.SystemAdminAddress = ""
			}),
			expErr: true,
		},
		"invalid system admin": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.SystemAdminAddress = "invalid"
			}),
			expErr: true,
		},
		"valid gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.GenTxs = []json.RawMessage{myGenTx}
				m.SeedContracts.Engagement = append(m.SeedContracts.Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Points:  11,
				})
			}),
		},
		"duplicate gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.GenTxs = []json.RawMessage{myGenTx, myGenTx}
				m.SeedContracts.Engagement = append(m.SeedContracts.Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Points:  11,
				})
			}),
			expErr: true,
		},
		"validator not in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.GenTxs = []json.RawMessage{myGenTx}
			}),
			expErr: true,
		},
		"invalid gentx json": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.GenTxs = []json.RawMessage{[]byte("invalid")}
			}),
			expErr: true,
		},
		"engagement contract not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.EngagementContractConfig = nil
			}),
			expErr: true,
		},
		"invalid engagement contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.EngagementContractConfig.Halflife = time.Nanosecond
			}),
			expErr: true,
		},
		"valset contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ValsetContractConfig = nil
			}),
			expErr: true,
		},
		"invalid valset contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ValsetContractConfig.MaxValidators = 0
			}),
			expErr: true,
		},
		"invalid valset contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ValsetContractConfig.EpochReward = sdk.NewCoin("alx", m.SeedContracts.ValsetContractConfig.EpochReward.Amount)
			}),
			expErr: true,
		},
		"stake contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.StakeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid stake contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.StakeContractConfig.UnbondingPeriod = 0
			}),
			expErr: true,
		},
		"oversight committee contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid oversight committee contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.Name = ""
			}),
			expErr: true,
		},
		"invalid oversight committee contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin("alx", m.SeedContracts.OversightCommitteeContractConfig.EscrowAmount.Amount)
			}),
			expErr: true,
		},
		"community pool contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.CommunityPoolContractConfig = nil
			}),
			expErr: true,
		},
		"invalid community pool contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.CommunityPoolContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
		"validator voting contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ValidatorVotingContractConfig = nil
			}),
			expErr: true,
		},
		"invalid validator voting contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ValidatorVotingContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
		"duplicate gentx pub keys": {
			source: GenesisStateFixture(func(m *GenesisState) {
				genTx1, opAddr1, _ := RandomGenTX(t, 101, func(m *MsgCreateValidator) {
					m.Pubkey = myPk
				})
				m.SeedContracts.GenTxs = []json.RawMessage{myGenTx, genTx1}
				m.SeedContracts.Engagement = []TG4Member{
					{Address: myOperatorAddr.String(), Points: 1},
					{Address: opAddr1.String(), Points: 2},
				}
			}),
			expErr: true,
		},
		"empty oversight community members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommunityMembers = nil
			}),
			expErr: true,
		},
		"duplicate oversight community members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommunityMembers = append(m.SeedContracts.OversightCommunityMembers, m.SeedContracts.OversightCommunityMembers[0])
			}),
			expErr: true,
		},
		"invalid oc members address": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommunityMembers = append(m.SeedContracts.OversightCommunityMembers, "invalid address")

			}),
			expErr: true,
		},
		"empty arbiter pool members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolMembers = nil
			}),
			expErr: true,
		},
		"duplicate arbiter pool members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolMembers = append(m.SeedContracts.ArbiterPoolMembers, m.SeedContracts.ArbiterPoolMembers[0])
			}),
			expErr: true,
		},
		"invalid ap members address": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolMembers = append(m.SeedContracts.ArbiterPoolMembers, "invalid address")

			}),
			expErr: true,
		},
		"arbiter pool contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig = nil
			}),
			expErr: true,
		},
		"invalid arbiter pool contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.DenyListContractAddress = "invalid address"
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
			src: DefaultGenesisState().SeedContracts.EngagementContractConfig,
		},
		"halflife empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.EngagementContractConfig.Halflife = 0
			}).SeedContracts.EngagementContractConfig,
		},
		"halflife contains elements < second": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.EngagementContractConfig.Halflife = time.Minute + time.Millisecond
			}).SeedContracts.EngagementContractConfig,
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
			src: *DefaultGenesisState().SeedContracts.ValsetContractConfig,
		},
		"max validators empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.ValsetContractConfig.MaxValidators = 0 },
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"scaling empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.ValsetContractConfig.Scaling = 0 },
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"epoch length empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.ValsetContractConfig.EpochLength = 0 },
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"fee percentage at min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.0000000000000001"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.SeedContracts.ValsetContractConfig.FeePercentage = val
				},
			).SeedContracts.ValsetContractConfig,
		},
		"fee percentage below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.00000000000000009"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.SeedContracts.ValsetContractConfig.FeePercentage = val
				},
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"validator rewards ratio zero": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("52.5")
				},
			).SeedContracts.ValsetContractConfig,
		},
		"validator rewards ratio 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(100)
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).SeedContracts.ValsetContractConfig,
		},
		"validator rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(101)
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"engagement rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.NewDec(101)
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"community pool rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.NewDec(101)
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("49")
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("49")
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("3")
				},
			).SeedContracts.ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio < 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("10")
					m.SeedContracts.ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("10")
					m.SeedContracts.ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("10")
				},
			).SeedContracts.ValsetContractConfig,
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
			src: *DefaultGenesisState().SeedContracts.StakeContractConfig,
		},
		"min bond empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.StakeContractConfig.MinBond = 0 },
			).SeedContracts.StakeContractConfig,
			expErr: true,
		},
		"tokens per weight empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.StakeContractConfig.TokensPerPoint = 0 },
			).SeedContracts.StakeContractConfig,
			expErr: true,
		},
		"unbonding period empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.SeedContracts.StakeContractConfig.UnbondingPeriod = 0 },
			).SeedContracts.StakeContractConfig,
			expErr: true,
		},
		"unbonding period below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.StakeContractConfig.UnbondingPeriod = time.Second - time.Nanosecond
				},
			).SeedContracts.StakeContractConfig,
			expErr: true,
		},
		"not convertable to seconds": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.SeedContracts.StakeContractConfig.UnbondingPeriod = time.Second + time.Nanosecond
				},
			).SeedContracts.StakeContractConfig,
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
			src: *DefaultGenesisState().SeedContracts.OversightCommitteeContractConfig,
		},
		"name empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.Name = ""
			}).SeedContracts.OversightCommitteeContractConfig,
			expErr: true,
		},
		"name too long": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.Name = strings.Repeat("a", 101)
			}).SeedContracts.OversightCommitteeContractConfig,
			expErr: true,
		},
		"escrow amount too low": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin(DefaultBondDenom, sdk.NewInt(999_999))
			}).SeedContracts.OversightCommitteeContractConfig,
			expErr: true,
		},
		"voting rules invalid": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).SeedContracts.OversightCommitteeContractConfig,
			expErr: true,
		},
		"deny contract address not an address": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.DenyListContractAddress = "not-an-address"
			}).SeedContracts.OversightCommitteeContractConfig,
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
			src: DefaultGenesisState().SeedContracts.OversightCommitteeContractConfig.VotingRules,
		},
		"default community pool": {
			src: DefaultGenesisState().SeedContracts.CommunityPoolContractConfig.VotingRules,
		},
		"default validator voting": {
			src: DefaultGenesisState().SeedContracts.ValidatorVotingContractConfig.VotingRules,
		},
		"voting period empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.Dec{}
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.OneDec()
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
		},
		"quorum less than min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.ZeroDec()
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(100)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
		},
		"quorum greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(101)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.Dec{}
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(50)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
		},
		"threshold lower min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(49)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(100)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
		},
		"threshold greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(101)
			}).SeedContracts.OversightCommitteeContractConfig.VotingRules,
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

func TestValidateArbiterPoolContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    ArbiterPoolContractConfig
		expErr bool
	}{
		"default": {
			src: *DefaultGenesisState().SeedContracts.ArbiterPoolContractConfig,
		},
		"name empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.Name = ""
			}).SeedContracts.ArbiterPoolContractConfig,
			expErr: true,
		},
		"name too long": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.Name = strings.Repeat("a", 101)
			}).SeedContracts.ArbiterPoolContractConfig,
			expErr: true,
		},
		"escrow amount too low": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.EscrowAmount = sdk.NewCoin(DefaultBondDenom, sdk.NewInt(999_999))
			}).SeedContracts.ArbiterPoolContractConfig,
			expErr: true,
		},
		"voting rules invalid": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.VotingRules.VotingPeriod = 0
			}).SeedContracts.ArbiterPoolContractConfig,
			expErr: true,
		},
		"deny contract address not an address": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.DenyListContractAddress = "not-an-address"
			}).SeedContracts.ArbiterPoolContractConfig,
			expErr: true,
		},
		"not convertible to seconds": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.SeedContracts.ArbiterPoolContractConfig.WaitingPeriod = time.Second + time.Nanosecond
			}).SeedContracts.ArbiterPoolContractConfig,
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
