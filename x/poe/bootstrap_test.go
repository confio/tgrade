package poe

import (
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm"
	twasmtesting "github.com/confio/tgrade/x/twasm/testing"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestBootstrapPoEContracts(t *testing.T) {
	specs := map[string]struct {
		genesis types.GenesisState
		expErr  bool
	}{
		"all contracts setup": {
			genesis: types.GenesisStateFixture(),
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
			sFn, capSetAddr := CaptureSetPoEContractAddressFn()
			pm := poeKeeperMock{
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
			for i, f := range []string{"tg4_group.wasm", "tg4_stake.wasm", "tg4_mixer.wasm", "tgrade_valset.wasm"} {
				c, err := ioutil.ReadFile(filepath.Join("contract", f))
				require.NoError(t, err)
				assert.Equal(t, c, (*capCreate)[i].WasmCode)
			}
			// and contracts instantiated
			require.Len(t, *capInst, 4)
			// and pinned
			assert.Equal(t, []uint64{1, 2, 3}, *capPin)

			assert.Equal(t, []CapturedPoEContractAddress{
				{ctype: types.PoEContractType_ENGAGEMENT, contractAddr: twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 1)},
				{ctype: types.PoEContractType_STAKING, contractAddr: twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 2)},
				{ctype: types.PoEContractType_MIXER, contractAddr: twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 3)},
				{ctype: types.PoEContractType_VALSET, contractAddr: twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 4)},
			}, *capSetAddr)
			// and privilege set
			require.Equal(t, []sdk.AccAddress{twasm.ContractAddress(twasmtesting.DefaultCaptureInstantiateFnCodeID, 4)}, *capPriv)
		})
	}
}

var _ twasmKeeper = twasmKeeperMock{}

type twasmKeeperMock struct {
	QuerySmartFn                    func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	SudoFn                          func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
	SetPrivilegedFn                 func(ctx sdk.Context, contractAddr sdk.AccAddress) error
	HasPrivilegedContractCallbackFn func(ctx sdk.Context, contractAddr sdk.AccAddress, callbackType twasmtypes.PrivilegedCallbackType) (bool, error)
}

func (m twasmKeeperMock) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
	if m.QuerySmartFn == nil {
		panic("not expected to be called")
	}
	return m.QuerySmartFn(ctx, contractAddr, req)
}

func (m twasmKeeperMock) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
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

func (m twasmKeeperMock) HasPrivilegedContractCallback(ctx sdk.Context, contractAddr sdk.AccAddress, callbackType twasmtypes.PrivilegedCallbackType) (bool, error) {
	if m.HasPrivilegedContractCallbackFn == nil {
		panic("not expected to be called")
	}
	return m.HasPrivilegedContractCallbackFn(ctx, contractAddr, callbackType)
}

func CaptureSetPrivilegedFn() (func(ctx sdk.Context, contractAddr sdk.AccAddress) error, *[]sdk.AccAddress) {
	var r []sdk.AccAddress
	return func(ctx sdk.Context, contractAddr sdk.AccAddress) error {
		r = append(r, contractAddr)
		return nil
	}, &r
}

var _ poeKeeper = poeKeeperMock{}

type poeKeeperMock struct {
	keeper.ContractSourceMock
	SetPoEContractAddressFn func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
}

func (m poeKeeperMock) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if m.GetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	return m.GetPoEContractAddressFn(ctx, ctype)
}

func (m poeKeeperMock) SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
	if m.SetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	m.SetPoEContractAddressFn(ctx, ctype, contractAddr)
}

type CapturedPoEContractAddress struct {
	ctype        types.PoEContractType
	contractAddr sdk.AccAddress
}

func CaptureSetPoEContractAddressFn() (func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress), *[]CapturedPoEContractAddress) {
	var r []CapturedPoEContractAddress
	return func(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
		r = append(r, CapturedPoEContractAddress{ctype: ctype, contractAddr: contractAddr})
	}, &r
}
