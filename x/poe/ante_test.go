package poe

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth/ante"

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
	var (
		myContractAddr   sdk.AccAddress = rand.Bytes(address.Len)
		mySenderAddr     sdk.AccAddress = rand.Bytes(address.Len)
		myFeeGranterAddr sdk.AccAddress = rand.Bytes(address.Len)
	)

	cs := keeper.PoEKeeperMock{GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
		require.Equal(t, types.PoEContractTypeValset, ctype)
		return myContractAddr, nil
	}}

	accountsMock := func(expAddr sdk.AccAddress) types.AccountKeeper {
		return accountKeeperMock{func(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
			require.Equal(t, expAddr, addr)
			return authtypes.NewBaseAccount(expAddr, nil, 1, 1)
		}}
	}

	capturingGrantKeeper, capturedGrantedFees := captureUseGrantedFees()
	specs := map[string]struct {
		feeAmount      sdk.Coins
		granter        sdk.AccAddress
		bank           types.BankKeeper
		grants         ante.FeegrantKeeper
		accounts       types.AccountKeeper
		expErr         bool
		expFeesGranted []capturedGrantedFee
	}{
		"with fee": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())},
			accounts:  accountsMock(mySenderAddr),
			bank: bankKeeperMock{SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				assert.Equal(t, mySenderAddr, fromAddr)
				assert.Equal(t, myContractAddr, toAddr)
				assert.Equal(t, sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())}, amt)
				return nil
			}},
		},
		"zero fee": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.ZeroInt())},
			accounts:  accountsMock(mySenderAddr),
			bank:      bankKeeperMock{},
		},
		"invalid denom fee": {
			feeAmount: sdk.Coins{sdk.Coin{Denom: "ALX$%^&", Amount: sdk.OneInt()}},
			accounts:  accountsMock(mySenderAddr),
			bank:      bankKeeperMock{},
			expErr:    true,
		},
		"with multiple fees": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			accounts:  accountsMock(mySenderAddr),
			bank: bankKeeperMock{SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				assert.Equal(t, mySenderAddr, fromAddr)
				assert.Equal(t, myContractAddr, toAddr)
				assert.Equal(t, sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))}, amt)
				return nil
			}},
		},
		"with multiple fees one amount is zero": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.ZeroInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			accounts:  accountsMock(mySenderAddr),
			bank:      bankKeeperMock{},
			expErr:    true,
		},
		"with multiple fees one denom is invalid": {
			feeAmount: sdk.Coins{sdk.Coin{Denom: "ALX$%^&", Amount: sdk.OneInt()}, sdk.NewCoin("BLX", sdk.NewInt(2))},
			accounts:  accountsMock(mySenderAddr),
			bank:      bankKeeperMock{},
			expErr:    true,
		},
		"with feegranter": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())},
			granter:   myFeeGranterAddr,
			accounts:  accountsMock(myFeeGranterAddr),
			grants:    capturingGrantKeeper,
			bank: bankKeeperMock{SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				assert.Equal(t, myFeeGranterAddr, fromAddr)
				assert.Equal(t, myContractAddr, toAddr)
				assert.Equal(t, sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())}, amt)
				return nil
			}},
			expFeesGranted: []capturedGrantedFee{
				{feeGranter: myFeeGranterAddr, feePayer: mySenderAddr, fee: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())}},
			},
		},
		"with feegranter rejected": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt())},
			granter:   myFeeGranterAddr,
			accounts:  accountsMock(myFeeGranterAddr),
			grants: feegrantMock{func(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
				return errors.New("testing")
			}},
			expErr: true,
		},
		"bank send fails": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			accounts:  accountsMock(mySenderAddr),
			bank: bankKeeperMock{SendCoinsFn: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
				return errors.New("testing")
			}},
			expErr: true,
		},
		"unknown account": {
			feeAmount: sdk.Coins{sdk.NewCoin("ALX", sdk.OneInt()), sdk.NewCoin("BLX", sdk.NewInt(2))},
			accounts: accountKeeperMock{func(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
				return nil
			}},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			*capturedGrantedFees = nil
			nextAnte, gotCalled := captureNextHandlerCall()
			em := sdk.NewEventManager()
			ctx := sdk.Context{}.WithEventManager(em)
			decorator := NewDeductFeeDecorator(spec.accounts, spec.bank, spec.grants, cs)
			_, gotErr := decorator.AnteHandle(ctx, newFeeTXMock(spec.feeAmount, mySenderAddr).WithGranter(spec.granter), false, nextAnte)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.True(t, *gotCalled, "next ante handler called")
			// and an event emitted
			require.Len(t, em.Events(), 1)
			require.Len(t, em.Events()[0].Attributes, 1)
			require.Equal(t, []byte(sdk.AttributeKeyFee), em.Events()[0].Attributes[0].Key)
			assert.Equal(t, spec.expFeesGranted, *capturedGrantedFees)
		})
	}
}

type capturedGrantedFee struct {
	feeGranter, feePayer sdk.AccAddress
	fee                  sdk.Coins
}

func captureUseGrantedFees() (feegrantMock, *[]capturedGrantedFee) {
	var result []capturedGrantedFee
	return feegrantMock{func(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
		result = append(result, capturedGrantedFee{granter, grantee, fee})
		return nil
	}}, &result
}

type bankKeeperMock struct {
	SendCoinsFn                          func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModuleFn       func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccountFn       func(ctx sdk.Context, s string, addr sdk.AccAddress, amt sdk.Coins) error
	DelegateCoinsFromAccountToModuleFn   func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccountFn func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

func (m bankKeeperMock) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.SendCoinsFromAccountToModuleFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

func (m bankKeeperMock) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsFromModuleToAccountFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

func (m bankKeeperMock) DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.DelegateCoinsFromAccountToModuleFn == nil {
		panic("not expected to be called")
	}
	return m.DelegateCoinsFromAccountToModuleFn(ctx, senderAddr, recipientModule, amt)
}

func (m bankKeeperMock) UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.UndelegateCoinsFromModuleToAccountFn == nil {
		panic("not expected to be called")
	}
	return m.UndelegateCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
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

var _ sdk.FeeTx = feeTXMock{}

type feeTXMock struct {
	fee     sdk.Coins
	payer   sdk.AccAddress
	granter sdk.AccAddress
	msgs    []sdk.Msg
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

func (f feeTXMock) FeeGranter() sdk.AccAddress {
	return f.granter
}

func (f feeTXMock) GetMsgs() []sdk.Msg {
	return f.msgs
}

func (f feeTXMock) ValidateBasic() error {
	panic("not expected to be called")
}

func (f feeTXMock) GetGas() uint64 {
	panic("not expected to be called")
}

func (f *feeTXMock) WithGranter(granter sdk.AccAddress) feeTXMock {
	f.granter = granter
	return *f
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
