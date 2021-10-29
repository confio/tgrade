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
		"rewards ratio zero": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorsRewardRatio = 0
				},
			).ValsetContractConfig,
		},
		"rewards ratio 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorsRewardRatio = 100
				},
			).ValsetContractConfig,
		},
		"rewards ratio > 100": {
			src: *GenesisStateFixture(
				func(m *GenesisState) {
					m.ValsetContractConfig.ValidatorsRewardRatio = 101
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

func TestTestValidateOversightCommitteeContractConfig(t *testing.T) {
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
		"voting period empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.VotingPeriod = 0
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"quorum empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Quorum = sdk.Dec{}
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"quorum zero": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Quorum = sdk.ZeroDec()
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"quorum greater max": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Quorum = sdk.NewDec(101)
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"threshold empty": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Threshold = sdk.Dec{}
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"threshold zero": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Threshold = sdk.ZeroDec()
			}).OversightCommitteeContractConfig,
			expErr: true,
		},
		"threshold greater max": {
			src: *GenesisStateFixture(func(m *GenesisState) {
				m.OversightCommitteeContractConfig.Threshold = sdk.NewDec(101)
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
