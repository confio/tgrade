package poe

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

func TestCreateValsetInitMsg(t *testing.T) {
	mixerContractAddr := types.RandomAccAddress()
	communityPoolAddr := types.RandomAccAddress()
	minDecimal := sdk.NewDec(1).QuoInt64(1_000_000_000_000_000_000)
	engagementID := uint64(7)
	engagementAddr := types.RandomAccAddress()
	bootstrapAccountAddr := types.RandomAccAddress()

	specs := map[string]struct {
		genesis *types.GenesisState
		exp     contract.ValsetInitMsg
	}{
		"default": {
			genesis: types.DefaultGenesisState(),
			exp: contract.ValsetInitMsg{
				Admin:                bootstrapAccountAddr.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(500),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				VerifyValidators:     false,
				OfflineJailDuration:  86400,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.GetSeedContracts().ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("50.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                bootstrapAccountAddr.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(501),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				VerifyValidators:     false,
				OfflineJailDuration:  86400,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with after comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.GetSeedContracts().ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.1")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                bootstrapAccountAddr.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        contract.DecimalFromProMille(1),
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				VerifyValidators:     false,
				OfflineJailDuration:  86400,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
		"fee percentage with min comma value": {
			genesis: types.GenesisStateFixture(func(m *types.GenesisState) {
				var err error
				m.GetSeedContracts().ValsetContractConfig.FeePercentage, err = sdk.NewDecFromStr("0.0000000000000001")
				require.NoError(t, err)
			}),
			exp: contract.ValsetInitMsg{
				Admin:                bootstrapAccountAddr.String(),
				Membership:           mixerContractAddr.String(),
				MinPoints:            1,
				MaxValidators:        100,
				EpochLength:          60,
				EpochReward:          sdk.NewCoin("utgd", sdk.NewInt(100_000)),
				Scaling:              1,
				FeePercentage:        &minDecimal,
				InitialKeys:          []contract.Validator{},
				ValidatorGroupCodeID: engagementID,
				VerifyValidators:     false,
				OfflineJailDuration:  86400,
				DistributionContracts: []contract.DistributionContract{
					{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
					{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
				},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := newValsetInitMsg(*spec.genesis.GetSeedContracts(), bootstrapAccountAddr, mixerContractAddr, engagementAddr, communityPoolAddr, engagementID)
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
	IsPinnedCodeFn          func(ctx sdk.Context, codeID uint64) bool
	GetContractInfoFn       func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
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

func (m twasmKeeperMock) IsPinnedCode(ctx sdk.Context, codeID uint64) bool {
	if m.IsPinnedCodeFn == nil {
		panic("not expected to be called")
	}
	return m.IsPinnedCodeFn(ctx, codeID)
}

func (m twasmKeeperMock) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
	if m.GetContractInfoFn == nil {
		panic("not expected to be called")
	}
	return m.GetContractInfoFn(ctx, contractAddress)
}
