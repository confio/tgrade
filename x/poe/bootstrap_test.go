package poe

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
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
		engagementContractAddr    = wasmkeeper.BuildContractAddress(1, 1)
		ocContractAddr            = wasmkeeper.BuildContractAddress(2, 2)
		stakingContractAdddr      = wasmkeeper.BuildContractAddress(3, 3)
		mixerContractAddr         = wasmkeeper.BuildContractAddress(4, 4)
		communityPoolContractAddr = wasmkeeper.BuildContractAddress(5, 5)
		valsetContractAddr        = wasmkeeper.BuildContractAddress(6, 6)
		distributionContractAddr  = wasmkeeper.BuildContractAddress(1, 6) // created by a contract so not really persisted
		ocGovProposalContractAddr = wasmkeeper.BuildContractAddress(7, 7) // instanceID = 7
		valVotingContractAddr     = wasmkeeper.BuildContractAddress(8, 8)
	)
	var (
		defaultLimit     uint64 = 20
		expFeePercentage        = contract.DecimalFromProMille(500)
		mySystemAdmin           = types.RandomAccAddress().String()
		myUser                  = types.RandomAccAddress().String()
		myOtherUser             = types.RandomAccAddress().String()
	)

	type contractSetup struct {
		expInitMsg   interface{}
		wasmFile     string
		contractAddr sdk.AccAddress
		codeID       uint64
		pinned       bool
		privileged   bool
	}
	allContracts := map[types.PoEContractType]contractSetup{
		types.PoEContractTypeEngagement: {
			expInitMsg: contract.TG4EngagementInitMsg{
				Admin:            mySystemAdmin, // updated later
				PreAuthsHooks:    1,
				PreAuthsSlashing: 1,
				Members:          []contract.TG4Member{{Addr: myUser, Points: 10}, {Addr: myOtherUser, Points: 11}},
				Denom:            "utgd",
				Halflife:         15552000,
			},
			wasmFile:     "tg4_engagement.wasm",
			contractAddr: engagementContractAddr,
			codeID:       1,
			pinned:       true,
		},
		types.PoEContractTypeOversightCommunity: {
			expInitMsg: contract.TrustedCircleInitMsg{
				Name:                      "Oversight Community",
				EscrowAmount:              sdk.NewInt(1_000_000),
				VotingPeriod:              30,
				Quorum:                    sdk.NewDecWithPrec(51, 2),
				Threshold:                 sdk.NewDecWithPrec(55, 2),
				AllowEndEarly:             true,
				DenyList:                  "",
				EditTrustedCircleDisabled: true,
				InitialMembers:            []string{},
				RewardDenom:               "utgd",
			},
			wasmFile:     "tgrade_trusted_circle.wasm",
			contractAddr: ocContractAddr,
			codeID:       2,
			pinned:       true,
		},
		types.PoEContractTypeStaking: {
			expInitMsg: contract.TG4StakeInitMsg{
				Admin:            mySystemAdmin,
				Denom:            "utgd",
				MinBond:          1,
				TokensPerPoint:   1,
				UnbondingPeriod:  21 * 24 * 60 * 60,
				AutoReturnLimit:  &defaultLimit,
				PreAuthsHooks:    1,
				PreAuthsSlashing: 1,
			},
			wasmFile:     "tg4_stake.wasm",
			contractAddr: stakingContractAdddr,
			codeID:       3,
			privileged:   true,
		},
		types.PoEContractTypeMixer: {
			expInitMsg: contract.TG4MixerInitMsg{
				PreAuthsSlashing: 1,
				LeftGroup:        engagementContractAddr.String(),
				RightGroup:       stakingContractAdddr.String(),
				FunctionType: contract.MixerFunction{
					GeometricMean: &struct{}{},
				},
			},
			wasmFile:     "tg4_mixer.wasm",
			contractAddr: mixerContractAddr,
			codeID:       4,
			pinned:       true,
		},
		types.PoEContractTypeValset: {
			expInitMsg: contract.ValsetInitMsg{
				Admin:                mySystemAdmin,
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        expFeePercentage,
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: 1,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementContractAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolContractAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				}},
			wasmFile:     "tgrade_valset.wasm",
			contractAddr: valsetContractAddr,
			codeID:       6,
			privileged:   true,
		},
		types.PoEContractTypeDistribution: {
			codeID:       1,
			contractAddr: distributionContractAddr,
		},
		types.PoEContractTypeOversightCommunityGovProposals: {
			expInitMsg: contract.OCProposalsInitMsg{
				ValsetContractAddress:     valsetContractAddr.String(),
				GroupContractAddress:      ocContractAddr.String(),
				EngagementContractAddress: engagementContractAddr.String(),
				VotingRules: contract.VotingRules{
					VotingPeriod:  30,
					Quorum:        sdk.NewDecWithPrec(51, 2),
					Threshold:     sdk.NewDecWithPrec(55, 2),
					AllowEndEarly: true,
				},
			},
			wasmFile:     "tgrade_oc_proposals.wasm",
			contractAddr: ocGovProposalContractAddr,
			codeID:       7,
			pinned:       true,
		},
		types.PoEContractTypeCommunityPool: {
			expInitMsg: contract.CommunityPoolInitMsg{
				VotingRules: contract.VotingRules{
					VotingPeriod:  21,
					Quorum:        sdk.NewDecWithPrec(10, 2),
					Threshold:     sdk.NewDecWithPrec(60, 2),
					AllowEndEarly: true,
				},
				GroupAddress: engagementContractAddr.String(),
			},
			wasmFile:     "tgrade_community_pool.wasm",
			contractAddr: communityPoolContractAddr,
			codeID:       5,
			pinned:       true,
		},
		types.PoEContractTypeValidatorVoting: {
			expInitMsg: contract.ValidatorVotingInitMsg{
				VotingRules: contract.VotingRules{
					VotingPeriod:  14,
					Quorum:        sdk.NewDecWithPrec(40, 2),
					Threshold:     sdk.NewDecWithPrec(66, 2),
					AllowEndEarly: true,
				},
				GroupAddress: distributionContractAddr.String(),
			},
			wasmFile:     "tgrade_validator_voting.wasm",
			contractAddr: valVotingContractAddr,
			codeID:       8,
			privileged:   true,
		},
	}
	expContractSetup := make([]contractSetup, 0, len(allContracts))
	bootstrapOrder := []types.PoEContractType{
		types.PoEContractTypeEngagement,
		types.PoEContractTypeOversightCommunity,
		types.PoEContractTypeStaking,
		types.PoEContractTypeMixer,
		types.PoEContractTypeCommunityPool,
		types.PoEContractTypeValset,
		types.PoEContractTypeDistribution,
		types.PoEContractTypeOversightCommunityGovProposals,
		types.PoEContractTypeValidatorVoting,
	}
	for _, v := range bootstrapOrder {
		if allContracts[v].expInitMsg == nil { // skip all that are not setup externally
			continue
		}
		expContractSetup = append(expContractSetup, allContracts[v])
	}
	// setup mocks
	cFn, capCreate := twasmtesting.CaptureCreateFn()
	iFn, capInst := twasmtesting.CaptureInstantiateFn(1, 2, 3, 4, 5, 6, 7, 8)
	pFn, capPin := twasmtesting.CapturePinCodeFn()
	uFn, capWasmAdminUpdates := captureWasmAdminUpdates()
	cm := twasmtesting.ContractOpsKeeperMock{
		CreateFn:              cFn,
		InstantiateFn:         iFn,
		PinCodeFn:             pFn,
		UpdateContractAdminFn: uFn,
	}

	spFn, capPriv := CaptureSetPrivilegedFn()
	hpFn := CaptureHasPrivilegedContractFn(capPriv)
	tm := twasmKeeperMock{
		SetPrivilegedFn:         spFn,
		HasPrivilegedContractFn: hpFn,
	}
	sFn, capSetAddr := keeper.CaptureSetPoEContractAddressFn()
	pm := keeper.PoEKeeperMock{
		SetPoEContractAddressFn: sFn,
		ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
			return poetesting.ValsetContractMock{
				QueryConfigFn: func(ctx sdk.Context) (*contract.ValsetConfigResponse, error) {
					return &contract.ValsetConfigResponse{ValidatorGroup: distributionContractAddr.String()}, nil
				},
				UpdateAdminFn: func(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error {
					assert.Equal(t, ocGovProposalContractAddr, newAdmin)
					assert.Equal(t, mySystemAdmin, sender.String())
					return nil
				},
			}
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
			require.Contains(t, allContracts, ctype)
			return allContracts[ctype].contractAddr, nil
		},
	}

	// when
	ctx := sdk.Context{}.WithLogger(log.TestingLogger())
	genesis := types.GenesisStateFixture(func(m *types.GenesisState) {
		m.SystemAdminAddress = mySystemAdmin
		m.Engagement = []types.TG4Member{{Address: myUser, Points: 10}, {Address: myOtherUser, Points: 11}}
	})
	gotErr := bootstrapPoEContracts(ctx, cm, tm, pm, genesis)

	// then
	require.NoError(t, gotErr)

	pos, pos2 := 0, 0
	for i, e := range expContractSetup {
		// and assert all codes got uploaded
		c, err := ioutil.ReadFile(filepath.Join("contract", e.wasmFile))
		require.NoError(t, err)
		require.Equal(t, c, (*capCreate)[i].WasmCode, e.wasmFile)
		// and contracts proper instantiated
		got := reflect.New(reflect.TypeOf(e.expInitMsg))
		require.NoError(t, json.Unmarshal((*capInst)[i].InitMsg, got.Interface()), "unmarshal json back to original type: %d", i)
		require.Equal(t, e.expInitMsg, got.Elem().Interface())
		// and code cache set
		switch {
		case e.pinned:
			require.Equal(t, e.codeID, (*capPin)[pos], "pinned")
			pos++
		case e.privileged:
			require.Equal(t, e.contractAddr, (*capPriv)[pos2])
			pos2++
		default:
			t.Fatal("not pinned or privileged")
		}
	}

	// and all contract addr stored for types
	for i, v := range bootstrapOrder {
		require.Equal(t, v, (*capSetAddr)[i].Ctype)
		require.Equal(t, allContracts[v].contractAddr, (*capSetAddr)[i].ContractAddr)
	}
	assert.Len(t, *capWasmAdminUpdates, len(allContracts))
	gotUpdates := make(map[string]sdk.AccAddress, len(allContracts))
	for _, v := range *capWasmAdminUpdates {
		gotUpdates[v.contractAddr.String()] = v.newAdmin
	}
	expUpdates := make(map[string]sdk.AccAddress, len(allContracts))
	for _, v := range allContracts {
		expUpdates[v.contractAddr.String()] = valVotingContractAddr
	}
	assert.Equal(t, expUpdates, gotUpdates)
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
	communityPoolAddr := types.RandomAccAddress()
	minDecimal := sdk.NewDec(1).QuoInt64(1_000_000_000_000_000_000)
	engagementID := uint64(7)
	engagementAddr := types.RandomAccAddress()
	systemAdmin := types.RandomAccAddress()

	specs := map[string]struct {
		genesis types.GenesisState
		exp     contract.ValsetInitMsg
	}{
		"default": {
			genesis: types.DefaultGenesisState(),
			exp: contract.ValsetInitMsg{
				Admin:                systemAdmin.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(500),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("50.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                systemAdmin.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(501),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with after comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                systemAdmin.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(1),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with min comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.0000000000000001")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                systemAdmin.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        &minDecimal,
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := newValsetInitMsg(spec.genesis, systemAdmin, mixerContractAddr, engagementAddr, communityPoolAddr, engagementID)
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

func CaptureHasPrivilegedContractFn(r *[]sdk.AccAddress) func(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType twasmtypes.PrivilegeType) (bool, error) {
	return func(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType twasmtypes.PrivilegeType) (bool, error) {
		for _, addr := range *r {
			if addr.Equals(contractAddr) {
				return true, nil
			}
		}
		return false, nil
	}
}
