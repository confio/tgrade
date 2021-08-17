package keeper

import (
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
)

var _ ContractSource = ContractSourceMock{}

// ContractSourceMock implements ContractSource interface for testing purpose.
// Subset of Keeper
type ContractSourceMock struct {
	GetPoEContractAddressFn func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
}

func (m ContractSourceMock) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if m.GetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	return m.GetPoEContractAddressFn(ctx, ctype)
}

var _ initer = PoEKeeperMock{}

// PoEKeeperMock mocks Keeper methods
type PoEKeeperMock struct {
	ContractSourceMock
	SetPoEContractAddressFn    func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
	setPoESystemAdminAddressFn func(ctx sdk.Context, admin sdk.AccAddress)
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

// return matching type or fail
func newContractSourceMock(t *testing.T, myValsetContract sdk.AccAddress, myStakingContract sdk.AccAddress) ContractSourceMock {
	return ContractSourceMock{
		GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
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
		},
	}
}
