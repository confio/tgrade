package keeper

import (
	"testing"
	"time"

	"github.com/confio/tgrade/x/poe/contract"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/types"
)

var _ PoEKeeper = PoEKeeperMock{}
var _ stakingQuerierKeeper = &PoEKeeperMock{}

// PoEKeeperMock mocks Keeper methods
type PoEKeeperMock struct {
	GetPoEContractAddressFn               func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
	SetValidatorInitialEngagementPointsFn func(ctx sdk.Context, address sdk.AccAddress, value sdk.Coin) error
	SetPoEContractAddressFn               func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
	setPoESystemAdminAddressFn            func(ctx sdk.Context, admin sdk.AccAddress)
	setParamsFn                           func(ctx sdk.Context, params types.Params)
	GetBondDenomFn                        func(ctx sdk.Context) string
	HistoricalEntriesFn                   func(ctx sdk.Context) uint32
	UnbondingTimeFn                       func(ctx sdk.Context) time.Duration
	GetHistoricalInfoFn                   func(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	DistributionContractFn                func(ctx sdk.Context) DistributionContract
	ValsetContractFn                      func(ctx sdk.Context) ValsetContract
	StakeContractFn                       func(ctx sdk.Context) StakeContract
}

func (m PoEKeeperMock) setParams(ctx sdk.Context, params types.Params) {
	if m.setParamsFn == nil {
		panic("not expected to be called")
	}
	m.setParamsFn(ctx, params)
}

func (m PoEKeeperMock) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if m.GetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	return m.GetPoEContractAddressFn(ctx, ctype)
}

func (m PoEKeeperMock) SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
	if m.SetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	m.SetPoEContractAddressFn(ctx, ctype, contractAddr)
}

func (m PoEKeeperMock) setPoESystemAdminAddress(ctx sdk.Context, admin sdk.AccAddress) {
	if m.setPoESystemAdminAddressFn == nil {
		panic("not expected to be called")
	}
	m.setPoESystemAdminAddressFn(ctx, admin)
}

func (m PoEKeeperMock) GetBondDenom(ctx sdk.Context) string {
	if m.GetBondDenomFn == nil {
		panic("not expected to be called")
	}
	return m.GetBondDenomFn(ctx)
}

func (m PoEKeeperMock) HistoricalEntries(ctx sdk.Context) uint32 {
	if m.HistoricalEntriesFn == nil {
		panic("not expected to be called")
	}
	return m.HistoricalEntriesFn(ctx)
}

func (m PoEKeeperMock) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	if m.GetHistoricalInfoFn == nil {
		panic("not expected to be called")
	}
	return m.GetHistoricalInfoFn(ctx, height)
}

func (m PoEKeeperMock) SetValidatorInitialEngagementPoints(ctx sdk.Context, opAddr sdk.AccAddress, points sdk.Coin) error {
	if m.SetValidatorInitialEngagementPointsFn == nil {
		panic("not expected to be called")
	}
	return m.SetValidatorInitialEngagementPointsFn(ctx, opAddr, points)
}

func (m PoEKeeperMock) DistributionContract(ctx sdk.Context) DistributionContract {
	if m.DistributionContractFn == nil {
		panic("not expected to be called")
	}
	return m.DistributionContractFn(ctx)
}

func (m PoEKeeperMock) StakeContract(ctx sdk.Context) StakeContract {
	if m.StakeContractFn == nil {
		panic("not expected to be called")
	}
	return m.StakeContractFn(ctx)
}

// CapturedPoEContractAddress data type
type CapturedPoEContractAddress struct {
	Ctype        types.PoEContractType
	ContractAddr sdk.AccAddress
}

// CaptureSetPoEContractAddressFn helper for mocks to capture data when called
func CaptureSetPoEContractAddressFn() (func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress), *[]CapturedPoEContractAddress) {
	var r []CapturedPoEContractAddress
	return func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
		r = append(r, CapturedPoEContractAddress{Ctype: ctype, ContractAddr: contractAddr})
	}, &r
}

var _ types.SmartQuerier = SmartQuerierMock{}

// SmartQuerierMock Mock queries to a contract
type SmartQuerierMock struct {
	QuerySmartFn func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
}

func (m SmartQuerierMock) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn == nil {
		panic("not expected to be called")
	}
	return m.QuerySmartFn(ctx, contractAddr, req)
}

func (m PoEKeeperMock) ValsetContract(ctx sdk.Context) ValsetContract {
	if m.ValsetContractFn == nil {
		panic("not expected to be called")
	}
	return m.ValsetContractFn(ctx)
}

// return matching type or fail
func newContractSourceMock(t *testing.T, myValsetContract sdk.AccAddress, myStakingContract sdk.AccAddress) PoEKeeperMock {
	return PoEKeeperMock{
		GetPoEContractAddressFn: SwitchPoEContractAddressFn(t, myValsetContract, myStakingContract),
	}
}

func SwitchPoEContractAddressFn(t *testing.T, myValsetContract sdk.AccAddress, myStakingContract sdk.AccAddress) func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	return func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
		switch ctype {
		case types.PoEContractTypeValset:
			if myValsetContract == nil {
				t.Fatalf("unexpected call to %s", types.PoEContractTypeValset)
			}
			return myValsetContract, nil
		case types.PoEContractTypeStaking:
			if myStakingContract == nil {
				t.Fatalf("unexpected call to %s", types.PoEContractTypeValset)
			}
			return myStakingContract, nil
		default:
			t.Fatalf("unexpected type: %s", ctype)
			return nil, nil
		}
	}
}

var _ TwasmKeeper = TwasmKeeperMock{}

// TwasmKeeperMock mock smart queries and sudo calls
type TwasmKeeperMock struct {
	QuerySmartFn func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	SudoFn       func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

func (m TwasmKeeperMock) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn == nil {
		panic("not expected to be called")
	}
	return m.QuerySmartFn(ctx, contractAddr, req)
}

func (m TwasmKeeperMock) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if m.SudoFn == nil {
		panic("not expected to be called")
	}
	return m.SudoFn(ctx, contractAddress, msg)
}

var _ DistributionContract = DistributionContractMock{}

type DistributionContractMock struct {
	ValidatorOutstandingRewardFn func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error)
}

func (m DistributionContractMock) ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
	if m.ValidatorOutstandingRewardFn == nil {
		panic("not expected to be called")
	}
	return m.ValidatorOutstandingRewardFn(ctx, addr)
}

var _ ValsetContract = ValsetContractMock{}

type ValsetContractMock struct {
	QueryValidatorFn func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error)
	ListValidatorsFn func(ctx sdk.Context) ([]stakingtypes.Validator, error)
	QueryConfigFn    func(ctx sdk.Context) (*contract.ValsetConfigResponse, error)
}

func (m ValsetContractMock) QueryValidator(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
	if m.QueryValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.QueryValidatorFn(ctx, opAddr)
}

func (m ValsetContractMock) ListValidators(ctx sdk.Context) ([]stakingtypes.Validator, error) {
	if m.ListValidatorsFn == nil {
		panic("not expected to be called")
	}
	return m.ListValidatorsFn(ctx)
}

func (m ValsetContractMock) QueryConfig(ctx sdk.Context) (*contract.ValsetConfigResponse, error) {
	if m.QueryConfigFn == nil {
		panic("not expected to be called")
	}
	return m.QueryConfigFn(ctx)
}

var _ StakeContract = StakeContractMock{}

type StakeContractMock struct {
	QueryStakingUnbondingPeriodFn func(ctx sdk.Context) (time.Duration, error)
	QueryStakingUnbondingFn       func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error)
	QueryStakedAmountFn           func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error)
}

func (m StakeContractMock) QueryStakedAmount(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
	if m.QueryStakedAmountFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakedAmountFn(ctx, opAddr)
}

func (m StakeContractMock) QueryStakingUnbondingPeriod(ctx sdk.Context) (time.Duration, error) {
	if m.QueryStakingUnbondingPeriodFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingPeriodFn(ctx)
}
func (m StakeContractMock) QueryStakingUnbonding(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
	if m.QueryStakingUnbondingFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingFn(ctx, opAddr)
}
