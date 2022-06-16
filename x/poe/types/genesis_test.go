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
	anyAccAddr := RandomAccAddress()
	myGenTx, myOperatorAddr, myPubKey := RandomGenTX(t, 100)
	myPk, err := codectypes.NewAnyWithValue(myPubKey)
	require.NoError(t, err)
	txConfig := MakeEncodingConfig(t).TxConfig
	specs := map[string]struct {
		source *GenesisState
		expErr bool
	}{
		"all good": {
			source: GenesisStateFixture(),
		},
		"seed with empty engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().Engagement = []TG4Member{}
			}),
			expErr: true,
		},
		"seed with duplicates in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().Engagement = []TG4Member{
					{Address: anyAccAddr.String(), Points: 1},
					{Address: anyAccAddr.String(), Points: 2},
				}
			}),
			expErr: true,
		},
		"seed with invalid addr in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().Engagement = []TG4Member{
					{Address: "invalid", Points: 1},
				}
			}),
			expErr: true,
		},
		"seed with invalid weight in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().Engagement = []TG4Member{
					{Address: RandomAccAddress().String(), Points: 0},
				}
			}),
			expErr: true,
		},
		"seed with legacy contracts": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.SetupMode = &GenesisState_ImportDump{&ImportDump{Contracts: []PoEContract{{ContractType: PoEContractTypeValset, Address: RandomAccAddress().String()}}}}
			}),
			expErr: true,
		},
		"empty bond denum": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().BondDenom = ""
			}),
			expErr: true,
		},
		"invalid bond denum": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().BondDenom = "&&&"
			}),
			expErr: true,
		},
		"empty bootstrap account": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().BootstrapAccountAddress = ""
			}),
			expErr: true,
		},
		"invalid bootstrap account": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().BootstrapAccountAddress = "invalid"
			}),
			expErr: true,
		},
		"valid gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().GenTxs = []json.RawMessage{myGenTx}
				m.GetSeedContracts().Engagement = append(m.GetSeedContracts().Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Points:  11,
				})
			}),
		},
		"duplicate gentx": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().GenTxs = []json.RawMessage{myGenTx, myGenTx}
				m.GetSeedContracts().Engagement = append(m.GetSeedContracts().Engagement, TG4Member{
					Address: myOperatorAddr.String(),
					Points:  11,
				})
			}),
			expErr: true,
		},
		"validator not in engagement group": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().GenTxs = []json.RawMessage{myGenTx}
			}),
			expErr: true,
		},
		"invalid gentx json": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().GenTxs = []json.RawMessage{[]byte("invalid")}
			}),
			expErr: true,
		},
		"engagement contract not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().EngagementContractConfig = nil
			}),
			expErr: true,
		},
		"invalid engagement contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().EngagementContractConfig.Halflife = time.Nanosecond
			}),
			expErr: true,
		},
		"valset contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ValsetContractConfig = nil
			}),
			expErr: true,
		},
		"invalid valset contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ValsetContractConfig.MaxValidators = 0
			}),
			expErr: true,
		},
		"invalid valset contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ValsetContractConfig.EpochReward = sdk.NewCoin("alx", m.GetSeedContracts().ValsetContractConfig.EpochReward.Amount)
			}),
			expErr: true,
		},
		"stake contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().StakeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid stake contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().StakeContractConfig.UnbondingPeriod = 0
			}),
			expErr: true,
		},
		"oversight committee contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig = nil
			}),
			expErr: true,
		},
		"invalid oversight committee contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.Name = ""
			}),
			expErr: true,
		},
		"invalid oversight committee contract denom": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin("alx", m.GetSeedContracts().OversightCommitteeContractConfig.EscrowAmount.Amount)
			}),
			expErr: true,
		},
		"community pool contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().CommunityPoolContractConfig = nil
			}),
			expErr: true,
		},
		"invalid community pool contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().CommunityPoolContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
		"validator voting contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ValidatorVotingContractConfig = nil
			}),
			expErr: true,
		},
		"invalid validator voting contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ValidatorVotingContractConfig.VotingRules.VotingPeriod = 0
			}),
			expErr: true,
		},
		"duplicate gentx pub keys": {
			source: GenesisStateFixture(func(m *GenesisState) {
				genTx1, opAddr1, _ := RandomGenTX(t, 101, func(m *MsgCreateValidator) {
					m.Pubkey = myPk
				})
				m.GetSeedContracts().GenTxs = []json.RawMessage{myGenTx, genTx1}
				m.GetSeedContracts().Engagement = []TG4Member{
					{Address: myOperatorAddr.String(), Points: 1},
					{Address: opAddr1.String(), Points: 2},
				}
			}),
			expErr: true,
		},
		"empty oversight community members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommunityMembers = nil
			}),
			expErr: true,
		},
		"duplicate oversight community members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommunityMembers = append(m.GetSeedContracts().OversightCommunityMembers, m.GetSeedContracts().OversightCommunityMembers[0])
			}),
			expErr: true,
		},
		"invalid oc members address": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommunityMembers = append(m.GetSeedContracts().OversightCommunityMembers, "invalid address")
			}),
			expErr: true,
		},
		"empty arbiter pool members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolMembers = nil
			}),
			expErr: true,
		},
		"duplicate arbiter pool members": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolMembers = append(m.GetSeedContracts().ArbiterPoolMembers, m.GetSeedContracts().ArbiterPoolMembers[0])
			}),
			expErr: true,
		},
		"invalid ap members address": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolMembers = append(m.GetSeedContracts().ArbiterPoolMembers, "invalid address")
			}),
			expErr: true,
		},
		"arbiter pool contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig = nil
			}),
			expErr: true,
		},
		"invalid arbiter pool contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.DenyListContractAddress = "invalid address"
			}),
			expErr: true,
		},
		"mixer contract config not set": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig = nil
			}),
			expErr: true,
		},
		"invalid mixer contract config": {
			source: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.S = sdk.Dec{}
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := ValidateGenesis(*spec.source, txConfig.TxJSONDecoder())
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
			src: DefaultGenesisState().GetSeedContracts().EngagementContractConfig,
		},
		"halflife empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().EngagementContractConfig.Halflife = 0
			}).GetSeedContracts().EngagementContractConfig,
		},
		"halflife contains elements < second": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().EngagementContractConfig.Halflife = time.Minute + time.Millisecond
			}).GetSeedContracts().EngagementContractConfig,
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
			src: *DefaultGenesisState().GetSeedContracts().ValsetContractConfig,
		},
		"max validators empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().ValsetContractConfig.MaxValidators = 0 },
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"scaling empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().ValsetContractConfig.Scaling = 0 },
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"epoch length empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().ValsetContractConfig.EpochLength = 0 },
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"epoch length contains sub seconds": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.EpochLength = time.Second + time.Millisecond
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"fee percentage at min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.0000000000000001"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.GetSeedContracts().ValsetContractConfig.FeePercentage = val
				},
			).GetSeedContracts().ValsetContractConfig,
		},
		"fee percentage below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					const min_fee_percentage = "0.00000000000000009"
					val, err := sdk.NewDecFromStr(min_fee_percentage)
					require.NoError(t, err)
					m.GetSeedContracts().ValsetContractConfig.FeePercentage = val
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"validator rewards ratio zero": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("52.5")
				},
			).GetSeedContracts().ValsetContractConfig,
		},
		"validator rewards ratio 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(100)
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).GetSeedContracts().ValsetContractConfig,
		},
		"validator rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.NewDec(101)
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"engagement rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.NewDec(101)
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.ZeroDec()
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"community pool rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.NewDec(101)
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.ZeroDec()
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.ZeroDec()
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("49")
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("49")
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("3")
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"total rewards ratio < 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.CommunityPoolRewardRatio = sdk.MustNewDecFromStr("10")
					m.GetSeedContracts().ValsetContractConfig.EngagementRewardRatio = sdk.MustNewDecFromStr("10")
					m.GetSeedContracts().ValsetContractConfig.ValidatorRewardRatio = sdk.MustNewDecFromStr("10")
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"offline jail duration is empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().ValsetContractConfig.OfflineJailDuration = 0 },
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"offline jail duration contains sub seconds": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().ValsetContractConfig.OfflineJailDuration = time.Second + time.Millisecond
				},
			).GetSeedContracts().ValsetContractConfig,
			expErr: true,
		},
		"verify validators not supported": { // see https://github.com/confio/tgrade/issues/389
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().ValsetContractConfig.VerifyValidators = true },
			).GetSeedContracts().ValsetContractConfig,
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
			src: *DefaultGenesisState().GetSeedContracts().StakeContractConfig,
		},
		"min bond empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().StakeContractConfig.MinBond = 0 },
			).GetSeedContracts().StakeContractConfig,
			expErr: true,
		},
		"tokens per weight empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().StakeContractConfig.TokensPerPoint = 0 },
			).GetSeedContracts().StakeContractConfig,
			expErr: true,
		},
		"unbonding period empty": {
			src: *GenesisStateFixture(
				func(m *GenesisState) { m.GetSeedContracts().StakeContractConfig.UnbondingPeriod = 0 },
			).GetSeedContracts().StakeContractConfig,
			expErr: true,
		},
		"unbonding period below min": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().StakeContractConfig.UnbondingPeriod = time.Second - time.Nanosecond
				},
			).GetSeedContracts().StakeContractConfig,
			expErr: true,
		},
		"not convertable to seconds": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.GetSeedContracts().StakeContractConfig.UnbondingPeriod = time.Second + time.Nanosecond
				},
			).GetSeedContracts().StakeContractConfig,
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
			src: *DefaultGenesisState().GetSeedContracts().OversightCommitteeContractConfig,
		},
		"name empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.Name = ""
			}).GetSeedContracts().OversightCommitteeContractConfig,
			expErr: true,
		},
		"name too long": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.Name = strings.Repeat("a", 101)
			}).GetSeedContracts().OversightCommitteeContractConfig,
			expErr: true,
		},
		"escrow amount too low": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.EscrowAmount = sdk.NewCoin(DefaultBondDenom, sdk.NewInt(999_999))
			}).GetSeedContracts().OversightCommitteeContractConfig,
			expErr: true,
		},
		"voting rules invalid": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).GetSeedContracts().OversightCommitteeContractConfig,
			expErr: true,
		},
		"deny contract address not an address": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.DenyListContractAddress = "not-an-address"
			}).GetSeedContracts().OversightCommitteeContractConfig,
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
			src: DefaultGenesisState().GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
		},
		"default community pool": {
			src: DefaultGenesisState().GetSeedContracts().CommunityPoolContractConfig.VotingRules,
		},
		"default validator voting": {
			src: DefaultGenesisState().GetSeedContracts().ValidatorVotingContractConfig.VotingRules,
		},
		"voting period empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.VotingPeriod = 0
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Quorum = sdk.Dec{}
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Quorum = sdk.OneDec()
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
		},
		"quorum less than min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Quorum = sdk.ZeroDec()
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"quorum at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(100)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
		},
		"quorum greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Quorum = sdk.NewDec(101)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold empty": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Threshold = sdk.Dec{}
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(50)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
		},
		"threshold lower min": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(49)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
			expErr: true,
		},
		"threshold at max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(100)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
		},
		"threshold greater max": {
			src: GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().OversightCommitteeContractConfig.VotingRules.Threshold = sdk.NewDec(101)
			}).GetSeedContracts().OversightCommitteeContractConfig.VotingRules,
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
			src: *(DefaultGenesisState().GetSeedContracts()).ArbiterPoolContractConfig,
		},
		"name empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.Name = ""
			}).GetSeedContracts().ArbiterPoolContractConfig,
			expErr: true,
		},
		"name too long": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.Name = strings.Repeat("a", 101)
			}).GetSeedContracts().ArbiterPoolContractConfig,
			expErr: true,
		},
		"escrow amount too low": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.EscrowAmount = sdk.NewCoin(DefaultBondDenom, sdk.NewInt(999_999))
			}).GetSeedContracts().ArbiterPoolContractConfig,
			expErr: true,
		},
		"voting rules invalid": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.VotingRules.VotingPeriod = 0
			}).GetSeedContracts().ArbiterPoolContractConfig,
			expErr: true,
		},
		"deny contract address not an address": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.DenyListContractAddress = "not-an-address"
			}).GetSeedContracts().ArbiterPoolContractConfig,
			expErr: true,
		},
		"not convertible to seconds": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().ArbiterPoolContractConfig.WaitingPeriod = time.Second + time.Nanosecond
			}).GetSeedContracts().ArbiterPoolContractConfig,
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

func TestValidateMixerContractConfig(t *testing.T) {
	specs := map[string]struct {
		src    MixerContractConfig
		expErr bool
	}{
		"default": {
			src: *(DefaultGenesisState().GetSeedContracts()).MixerContractConfig,
		},
		"max rewards empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.MaxRewards = 0
			}).GetSeedContracts().MixerContractConfig,
			expErr: true,
		},
		"sigmoid p empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.P = sdk.NewDec(0)
			}).GetSeedContracts().MixerContractConfig,
			expErr: true,
		},
		"sigmoid p unset": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.P = sdk.Dec{}
			}).GetSeedContracts().MixerContractConfig,
			expErr: true,
		},
		"sigmoid s empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.S = sdk.NewDec(0)
			}).GetSeedContracts().MixerContractConfig,
			expErr: true,
		},
		"sigmoid s unset": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.GetSeedContracts().MixerContractConfig.Sigmoid.S = sdk.Dec{}
			}).GetSeedContracts().MixerContractConfig,
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
