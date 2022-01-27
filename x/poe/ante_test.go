package poe

import (
	"errors"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/types/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

func TestDeductFeeDecorator(t *testing.T) {
	t.Skip("TODO: Use mocks and test fee grant")
	var (
		myContractAddr sdk.AccAddress = rand.Bytes(address.Len)
		mySenderAddr   sdk.AccAddress = rand.Bytes(address.Len)
	)

	cs := keeper.PoEKeeperMock{GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
		require.Equal(t, types.PoEContractTypeValset, ctype)
		return myContractAddr, nil
	}}

	specs := map[string]struct {
		feeAmount sdk.Coins
		bankMock  bankKeeper
		expErr    bool
	}{
		"with fee": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())},
			bankMock: bankKeeperMock{func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				assert.Equal(t, mySenderAddr, fromAddr)
				assert.Equal(t, myContractAddr, toAddr)
				assert.Equal(t, sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())}, amt)
				return nil
			}},
		},
		"zero fee": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.ZeroInt())},
			bankMock:  bankKeeperMock{},
		},
		"invalid denom fee": {
			feeAmount: sdk.Coins{sdk.Coin{Denom: "ALX$%^&", Amount: sdk.OneInt()}},
			bankMock:  bankKeeperMock{},
			expErr:    true,
		},
		"with multiple fees": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			bankMock: bankKeeperMock{func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				assert.Equal(t, mySenderAddr, fromAddr)
				assert.Equal(t, myContractAddr, toAddr)
				assert.Equal(t, sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))}, amt)
				return nil
			}},
		},
		"with multiple fees one amount is zero": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.ZeroInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			bankMock:  bankKeeperMock{},
			expErr:    true,
		},
		"with multiple fees one denom is invalid": {
			feeAmount: sdk.Coins{sdk.Coin{Denom: "ALX$%^&", Amount: sdk.OneInt()}, sdk.NewCoin("BLX", sdk.NewInt(2))},
			bankMock:  bankKeeperMock{},
			expErr:    true,
		},
		"bank send fails": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			bankMock: bankKeeperMock{func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return errors.New("testing")
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			nextAnte, gotCalled := captureNextHandlerCall()
			ctx := sdk.Context{}
			decorator := NewDeductFeeDecorator(accountKeeperMock{}, spec.bankMock, feegrantMock{}, cs)
			_, gotErr := decorator.AnteHandle(ctx, newFeeTXMock(spec.feeAmount, mySenderAddr), false, nextAnte)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, *gotCalled, "next ante handler called")
		})
	}
}

type bankKeeperMock struct {
	SendCoinsFn func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

func (m bankKeeperMock) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFn(ctx, fromAddr, toAddr, amt)
}

func captureNextHandlerCall() (sdk.AnteHandler, *bool) {
	var called bool
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		called = true
		return ctx, nil
	}, &called
}

type feeTXMock struct {
	sdk.FeeTx
	fee   sdk.Coins
	payer sdk.AccAddress
}

func newFeeTXMock(fee sdk.Coins, payer sdk.AccAddress) *feeTXMock {
	return &feeTXMock{fee: fee, payer: payer}
}

func (f feeTXMock) GetFee() sdk.Coins {
	return f.fee
}

func (f feeTXMock) FeePayer() sdk.AccAddress {
	return f.payer
}

type feegrantMock struct {
	UseGrantedFeesFn func(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

func (m feegrantMock) UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	if m.UseGrantedFeesFn == nil {
		panic("not expected to be called")
	}
	return m.UseGrantedFeesFn(ctx, granter, grantee, fee, msgs)
}

type accountKeeperMock struct {
	GetAccountFn func(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

func (m accountKeeperMock) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	if m.GetAccountFn == nil {
		panic("not expected to be called")
	}
	return m.GetAccountFn(ctx, addr)
}
