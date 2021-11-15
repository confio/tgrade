package poe

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/confio/tgrade/x/poe/keeper/poetesting"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	twasmtesting "github.com/confio/tgrade/x/twasm/testing"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

func TestBootstrapPoEContracts(t *testing.T) {
	var (
		defaultLimit     uint64 = 20
		expFeePercentage        = contract.DecimalFromProMille(500)
		mySystemAdmin           = types.RandomAccAddress().String()
		myUser                  = types.RandomAccAddress().String()
		myOtherUser             = types.RandomAccAddress().String()
	)

	var (
		engagementContractAddr    = wasmkeeper.BuildContractAddress(1, 1)
		ocContractAddr            = wasmkeeper.BuildContractAddress(2, 2)
		ocGovProposalContractAddr = wasmkeeper.BuildContractAddress(3, 3)
		stakingContractAdddr      = wasmkeeper.BuildContractAddress(4, 4)
		mixerContractAddr         = wasmkeeper.BuildContractAddress(5, 5)
		valsetContractAddr        = wasmkeeper.BuildContractAddress(6, 6)
		distributionContractAddr  = wasmkeeper.BuildContractAddress(1, 7) // created by a contract so not really persisted

	)

	specs := map[string]struct {
		genesis                   types.GenesisState
		expEngagementInit         contract.TG4EngagementInitMsg
		expStakerInit             contract.TG4StakeInitMsg
		expValsetInit             contract.ValsetInitMsg
		expOversightCommitteeInit contract.TrustedCircleInitMsg
		expOCGovProposalsInit     contract.OCProposalsInitMsg
		expErr                    bool
		expMixerInit              contract.TG4MixerInitMsg
	}{
		"all contracts setup": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				m.SystemAdminAddress = mySystemAdmin
				m.Engagement = []types.TG4Member{{Address: myUser, Weight: 10}, {Address: myOtherUser, Weight: 11}}
				m.ValsetContractConfig.FeePercentage = sdk.NewDec(50)
			}),
			expEngagementInit: contract.TG4EngagementInitMsg{
				Admin:            mySystemAdmin, // updated later
				PreAuthsHooks:    1,
				PreAuthsSlashing: 1,
				Members:          []contract.TG4Member{{Addr: myUser, Weight: 10}, {Addr: myOtherUser, Weight: 11}},
				Token:            "utgd",
				Halflife:         15552000,
			},
			expStakerInit: contract.TG4StakeInitMsg{
				Admin:            mySystemAdmin,
				Denom:            "utgd",
				MinBond:          1,
				TokensPerWeight:  1,
				UnbondingPeriod:  21 * 24 * 60 * 60,
				AutoReturnLimit:  &defaultLimit,
				PreAuthsHooks:    1,
				PreAuthsSlashing: 1,
			},
			expValsetInit: contract.ValsetInitMsg{
				Membership:            mixerContractAddr.String(),
				MinWeight:             1,
				MaxValidators:         100,
				EpochLength:           60,
				EpochReward:           sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:               1,
				FeePercentage:         expFeePercentage,
				InitialKeys:           []contract.Validator{},
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
				RewardsCodeID:         1,
				DistributionContract:  engagementContractAddr.String(),
			},
			expOversightCommitteeInit: contract.TrustedCircleInitMsg{
				Name:                      "Oversight Community",
				EscrowAmount:              sdk.NewInt(1_000_000),
				VotingPeriod:              1,
				Quorum:                    sdk.NewDecWithPrec(50, 2),
				Threshold:                 sdk.NewDecWithPrec(66, 2),
				AllowEndEarly:             true,
				DenyList:                  "",
				EditTrustedCircleDisabled: true,
				InitialMembers:            []string{},
			},
			expOCGovProposalsInit: contract.OCProposalsInitMsg{
				GroupContractAddress:      ocContractAddr.String(),
				EngagementContractAddress: engagementContractAddr.String(),
				VotingRules: contract.VotingRules{
					VotingPeriod:  1,
					Quorum:        sdk.NewDecWithPrec(50, 2),
					Threshold:     sdk.NewDecWithPrec(66, 2),
					AllowEndEarly: true,
				},
			},
			expMixerInit: contract.TG4MixerInitMsg{
				PreAuthsSlashing: 1,
				LeftGroup:        engagementContractAddr.String(),
				RightGroup:       stakingContractAdddr.String(),
				FunctionType: contract.MixerFunction{
					GeometricMean: &struct{}{},
				},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cFn, capCreate := twasmtesting.CaptureCreateFn()
			iFn, capInst := twasmtesting.CaptureInstantiateFn(1, 2, 3, 4, 5, 6, 7, 8)
			pFn, capPin := twasmtesting.CapturePinCodeFn()
			uFn, capAdminUpdates := captureWasmAdminUpdates()
			cm := twasmtesting.ContractOpsKeeperMock{
				CreateFn:              cFn,
				InstantiateFn:         iFn,
				PinCodeFn:             pFn,
				UpdateContractAdminFn: uFn,
			}

			spFn, capPriv := CaptureSetPrivilegedFn()
			tm := twasmKeeperMock{
				SetPrivilegedFn: spFn,
				QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
					require.Equal(t, valsetContractAddr, contractAddr)
					cfg := contract.ValsetConfigResponse{
						DistributionContract: distributionContractAddr.String(),
					}
					return json.Marshal(cfg)
				},
			}
			sFn, capSetAddr := keeper.CaptureSetPoEContractAddressFn()
			pm := keeper.PoEKeeperMock{
				SetPoEContractAddressFn: sFn,
				ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
					return poetesting.ValsetContractMock{QueryConfigFn: func(ctx sdk.Context) (*contract.ValsetConfigResponse, error) {
						return &contract.ValsetConfigResponse{DistributionContract: distributionContractAddr.String()}, nil
					}}
				},
				EngagementContractFn: func(ctx sdk.Context) keeper.EngagementContract {
					return poetesting.EngagementContractMock{
						UpdateAdminFn: func(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error {
							assert.Equal(t, ocGovProposalContractAddr, newAdmin)
							assert.Equal(t, mySystemAdmin, sender.String())
							return nil
						},
					}
				},
				GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
					require.Equal(t, types.PoEContractTypeEngagement, ctype)
					return engagementContractAddr, nil
				},
			}
			// when
			ctx := sdk.Context{}.WithLogger(log.TestingLogger())
			gotErr := bootstrapPoEContracts(ctx, cm, tm, pm, spec.genesis)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			// and assert all codes got uploaded
			require.Equal(t, 6, len(*capCreate))
			for i, f := range []string{"tg4_engagement.wasm", "tgrade_trusted_circle.wasm", "tgrade_oc_proposals.wasm", "tg4_stake.wasm", "tg4_mixer.wasm", "tgrade_valset.wasm"} {
				c, err := ioutil.ReadFile(filepath.Join("contract", f))
				require.NoError(t, err)
				assert.Equal(t, c, (*capCreate)[i].WasmCode)
			}
			// and contracts proper instantiated
			require.Len(t, *capInst, 6)

			var (
				gotEngagementInit    contract.TG4EngagementInitMsg
				gotOCInit            contract.TrustedCircleInitMsg
				gotOCGovProposalInit contract.OCProposalsInitMsg
				gotStakerInit        contract.TG4StakeInitMsg
				gotMixerInit         contract.TG4MixerInitMsg
				gotValsetInit        contract.ValsetInitMsg
			)
			for i, ref := range []interface{}{&gotEngagementInit, &gotOCInit, &gotOCGovProposalInit, &gotStakerInit, &gotMixerInit, &gotValsetInit} {
				require.NoError(t, json.Unmarshal((*capInst)[i].InitMsg, ref))
			}
			assert.Equal(t, spec.expEngagementInit, gotEngagementInit)
			assert.Equal(t, spec.expStakerInit, gotStakerInit)
			assert.Equal(t, spec.expValsetInit, gotValsetInit)
			assert.Equal(t, spec.expOversightCommitteeInit, gotOCInit)
			assert.Equal(t, spec.expOCGovProposalsInit, gotOCGovProposalInit)
			assert.Equal(t, spec.expMixerInit, gotMixerInit)

			// and pinned
			assert.Equal(t, []uint64{1, 2, 3, 5}, *capPin)
			// or privileged
			require.Equal(t, []sdk.AccAddress{stakingContractAdddr, valsetContractAddr}, *capPriv)

			// and contract addr stored for types
			assert.Equal(t, []keeper.CapturedPoEContractAddress{
				{Ctype: types.PoEContractTypeEngagement, ContractAddr: engagementContractAddr},
				{Ctype: types.PoEContractTypeOversightCommunity, ContractAddr: ocContractAddr},
				{Ctype: types.PoEContractTypeOversightCommunityGovProposals, ContractAddr: ocGovProposalContractAddr},
				{Ctype: types.PoEContractTypeStaking, ContractAddr: stakingContractAdddr},
				{Ctype: types.PoEContractTypeMixer, ContractAddr: mixerContractAddr},
				{Ctype: types.PoEContractTypeValset, ContractAddr: valsetContractAddr},
				{Ctype: types.PoEContractTypeDistribution, ContractAddr: distributionContractAddr},
			}, *capSetAddr)

			assert.Empty(t, *capAdminUpdates)
		})
	}
}

type capturedContractAdminUpdate struct {
	contractAddr, newAdmin sdk.AccAddress
}

func captureWasmAdminUpdates() (func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newAdmin sdk.AccAddress) error, *[]capturedContractAdminUpdate) {
	var result []capturedContractAdminUpdate
	return func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newAdmin sdk.AccAddress) error {
		result = append(result, capturedContractAdminUpdate{contractAddr: contractAddress, newAdmin: newAdmin})
		return nil
	}, &result
}

func TestCreateValsetInitMsg(t *testing.T) {
	mixerContractAddr := types.RandomAccAddress()
	minDecimal := sdk.NewDec(1).QuoInt64(1_000_000_000_000_000_000)
	engagementID := uint64(7)
	engagementAddr := types.RandomAccAddress()

	specs := map[string]struct {
		genesis types.GenesisState
		exp     contract.ValsetInitMsg
	}{
		"default": {
			genesis: types.DefaultGenesisState(),
			exp: contract.ValsetInitMsg{
				Membership:            mixerContractAddr.String(),
				MinWeight:             1,
				MaxValidators:         100,
				EpochLength:           60,
				EpochReward:           sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:               1,
				FeePercentage:         contract.DecimalFromProMille(500),
				InitialKeys:           []contract.Validator{},
				RewardsCodeID:         engagementID,
				DistributionContract:  engagementAddr.String(),
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
			},
		},
		"fee percentage with comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("50.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Membership:            mixerContractAddr.String(),
				MinWeight:             1,
				MaxValidators:         100,
				EpochLength:           60,
				EpochReward:           sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:               1,
				FeePercentage:         contract.DecimalFromProMille(501),
				InitialKeys:           []contract.Validator{},
				RewardsCodeID:         engagementID,
				DistributionContract:  engagementAddr.String(),
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
			},
		},
		"fee percentage with after comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Membership:            mixerContractAddr.String(),
				MinWeight:             1,
				MaxValidators:         100,
				EpochLength:           60,
				EpochReward:           sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:               1,
				FeePercentage:         contract.DecimalFromProMille(1),
				InitialKeys:           []contract.Validator{},
				RewardsCodeID:         engagementID,
				DistributionContract:  engagementAddr.String(),
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
			},
		},
		"fee percentage with min comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.0000000000000001")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Membership:            mixerContractAddr.String(),
				MinWeight:             1,
				MaxValidators:         100,
				EpochLength:           60,
				EpochReward:           sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:               1,
				FeePercentage:         &minDecimal,
				InitialKeys:           []contract.Validator{},
				RewardsCodeID:         engagementID,
				DistributionContract:  engagementAddr.String(),
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := newValsetInitMsg(spec.genesis, mixerContractAddr, engagementAddr, engagementID)
			assert.Equal(t, spec.exp, got)
		})
	}
}

var _ twasmKeeper = twasmKeeperMock{}

type twasmKeeperMock struct {
	QuerySmartFn            func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	SudoFn                  func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	SetPrivilegedFn         func(ctx sdk.Context, contractAddr sdk.AccAddress) error
	HasPrivilegedContractFn func(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType twasmtypes.PrivilegeType) (bool, error)
}

func (m twasmKeeperMock) GetContractKeeper() wasmtypes.ContractOpsKeeper {
	panic("implement me")
}

func (m twasmKeeperMock) IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	panic("implement me")
}

func (m twasmKeeperMock) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn == nil {
		panic("not expected to be called")
	}
	return m.QuerySmartFn(ctx, contractAddr, req)
}

func (m twasmKeeperMock) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if m.SudoFn == nil {
		panic("not expected to be called")
	}
	return m.SudoFn(ctx, contractAddress, msg)
}

func (m twasmKeeperMock) SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	if m.SetPrivilegedFn == nil {
		panic("not expected to be called")
	}
	return m.SetPrivilegedFn(ctx, contractAddr)
}

func (m twasmKeeperMock) HasPrivilegedContract(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType twasmtypes.PrivilegeType) (bool, error) {
	if m.HasPrivilegedContractFn == nil {
		panic("not expected to be called")
	}
	return m.HasPrivilegedContractFn(ctx, contractAddr, privilegeType)
}

func CaptureSetPrivilegedFn() (func(ctx sdk.Context, contractAddr sdk.AccAddress) error, *[]sdk.AccAddress) {
	var r []sdk.AccAddress
	return func(ctx sdk.Context, contractAddr sdk.AccAddress) error {
		r = append(r, contractAddr)
		return nil
	}, &r
}
