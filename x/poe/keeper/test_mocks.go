package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/types"
)

var _ PoEKeeper = PoEKeeperMock{}
var _ stakingQuerierKeeper = PoEKeeperMock{}

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

func (m PoEKeeperMock) UnbondingTime(ctx sdk.Context) time.Duration {
	if m.UnbondingTimeFn == nil {
		panic("not expected to be called")
	}
	return m.UnbondingTimeFn(ctx)
}

func (m PoEKeeperMock) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	if m.GetHistoricalInfoFn == nil {
		panic("not expected to be called")
	}
	return m.GetHistoricalInfoFn(ctx, height)
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
