package keeper

import (
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ ContractSource = ContractSourceMock{}

// ContractSourceMock implementes ContractSource interface for testing purpose
type ContractSourceMock struct {
	GetPoEContractAddressFn func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
}

func (m ContractSourceMock) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if m.GetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	return m.GetPoEContractAddressFn(ctx, ctype)
}
