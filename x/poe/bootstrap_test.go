package poe

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm"
	twasmtesting "github.com/confio/tgrade/x/twasm/testing"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		engagementContractAddr = twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 1)
		stakingContractAdddr   = twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 2)
		mixerContractAddr      = twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 3)
		valsetContractAddr     = twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 4)
	)

	specs := map[string]struct {
		genesis           types.GenesisState
		expEngagementInit contract.TG4EngagementInitMsg
		expStakerInit     contract.TG4StakeInitMsg
		expValsetInit     contract.ValsetInitMsg
		expErr            bool
	}{
		"all contracts setup": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				m.SystemAdminAddress = mySystemAdmin
				m.Engagement = []types.TG4Member{{Address: myUser, Weight: 10}, {Address: myOtherUser, Weight: 11}}
				m.ValsetContractConfig.FeePercentage = sdk.NewDec(50)
			}),
			expEngagementInit: contract.TG4EngagementInitMsg{
				Admin:    mySystemAdmin,
				Preauths: 1,
				Members:  []contract.TG4Member{{Addr: myUser, Weight: 10}, {Addr: myOtherUser, Weight: 11}},
				Token:    "utgd",
				Halflife: 15552000,
			},
			expStakerInit: contract.TG4StakeInitMsg{
				Admin:           mySystemAdmin,
				Denom:           "utgd",
				MinBond:         1,
				TokensPerWeight: 1,
				UnbondingPeriod: 21 * 24 * 60 * 60,
				AutoReturnLimit: &defaultLimit,
				Preauths:        1,
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
				RewardsCodeId:         1,
				DistributionContract:  "cosmos156r47kpk4va938pmtpuee4fh77847gqcq4xu6e",
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cFn, capCreate := twasmtesting.CaptureCreateFn()
			iFn, capInst := twasmtesting.CaptureInstantiateFn()
			pFn, capPin := twasmtesting.CapturePinCodeFn()
			cm := twasmtesting.ContractOpsKeeperMock{
				CreateFn:      cFn,
				InstantiateFn: iFn,
				PinCodeFn:     pFn,
			}

			spFn, capPriv := CaptureSetPrivilegedFn()
			tm := twasmKeeperMock{
				SetPrivilegedFn: spFn,
			}
			sFn, capSetAddr := keeper.CaptureSetPoEContractAddressFn()
			pm := keeper.PoEKeeperMock{
				SetPoEContractAddressFn: sFn,
			}
			// when
			ctx := sdk.Context{}
			gotErr := bootstrapPoEContracts(ctx, cm, tm, pm, spec.genesis)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			// and codes uploaded
			require.Len(t, *capCreate, 4, "got %d", len(*capCreate))
			for i, f := range []string{"tg4_engagement.wasm", "tg4_stake.wasm", "tg4_mixer.wasm", "tgrade_valset.wasm"} {
				c, err := ioutil.ReadFile(filepath.Join("contract", f))
				require.NoError(t, err)
				assert.Equal(t, c, (*capCreate)[i].WasmCode)
			}
			// and contracts proper instantiated
			require.Len(t, *capInst, 4)

			var (
				gotEngagementInit contract.TG4EngagementInitMsg
				gotStakerInit     contract.TG4StakeInitMsg
				gotMixerInit      contract.TG4MixerInitMsg
				gotValsetInit     contract.ValsetInitMsg
			)
			for i, ref := range []interface{}{&gotEngagementInit, &gotStakerInit, &gotMixerInit, &gotValsetInit} {
				require.NoError(t, json.Unmarshal((*capInst)[i].InitMsg, ref))
			}
			assert.Equal(t, spec.expEngagementInit, gotEngagementInit)
			assert.Equal(t, spec.expStakerInit, gotStakerInit)
			assert.Equal(t, spec.expValsetInit, gotValsetInit)
			expMixerInit := contract.TG4MixerInitMsg{
				LeftGroup:  engagementContractAddr.String(),
				RightGroup: stakingContractAdddr.String(),
				FunctionType: contract.MixerFunction{
					GeometricMean: &struct{}{},
				},
			}
			assert.Equal(t, expMixerInit, gotMixerInit)

			// and pinned or privileged
			assert.Equal(t, []uint64{1, 3}, *capPin)
			require.Equal(t, []sdk.AccAddress{stakingContractAdddr, valsetContractAddr}, *capPriv)

			// and contract addr stored for types
			assert.Equal(t, []keeper.CapturedPoEContractAddress{
				{Ctype: types.PoEContractTypeEngagement, ContractAddr: engagementContractAddr},
				{Ctype: types.PoEContractTypeStaking, ContractAddr: stakingContractAdddr},
				{Ctype: types.PoEContractTypeMixer, ContractAddr: mixerContractAddr},
				{Ctype: types.PoEContractTypeValset, ContractAddr: valsetContractAddr},
			}, *capSetAddr)
		})
	}
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
				RewardsCodeId:         engagementID,
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
				RewardsCodeId:         engagementID,
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
				RewardsCodeId:         engagementID,
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
				RewardsCodeId:         engagementID,
				DistributionContract:  engagementAddr.String(),
				ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := newValsetInitMsg(mixerContractAddr, spec.genesis, engagementAddr, engagementID)
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
