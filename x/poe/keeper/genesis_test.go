package keeper

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/poe/types"
)

func TestInitGenesis(t *testing.T) {
	initBech32Prefixes()

	txConfig := types.MakeEncodingConfig(t).TxConfig

	specs := map[string]struct {
		src                    types.GenesisState
		respCode               uint32
		expErr                 bool
		expDeliveredGenTxCount int
		expContracts           []CapturedPoEContractAddress
		expParams              types.Params
	}{
		"all good": {
			src: types.GenesisStateFixture(func(m *types.GenesisState) {
				m.GenTxs = []json.RawMessage{[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"moniker-0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18katt8evmwr7g0w545g9kgrn2s6z9a0ky27gdp","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"P4qRtdI2pfl5IZN4cv28uuFhsRhSc/CBzrlB2/+ATQs="},"amount":{"denom":"utgd","amount":"100000000"},"vesting_amount":{"denom":"utgd","amount":"0"}}],"memo":"7973f9800a585f9a5e730ee18e4abab9a06214f5@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlOUwEtcgY5rV6cJHCJJntNdrY9Kpe057pY6yewFbhxW"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["cJnG18yHwWsgxjh1Kqf3j/MFv7OpX69c7VLQz1MX1qtndFIylSNPVbkXOMu2i+Ufy52nXH3yujOKsMIVLP62pg=="]}`)}
			},
			),
			expDeliveredGenTxCount: 1,
			expParams:              types.DefaultParams(),
		},
		"deliver genTX failed": {
			src: types.GenesisStateFixture(func(m *types.GenesisState) {
				m.GenTxs = []json.RawMessage{[]byte(`{}`)}
			}),
			expErr:                 true,
			expDeliveredGenTxCount: 1,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var capturedTxs []abci.RequestDeliverTx
			captureTx := func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
				capturedTxs = append(capturedTxs, tx)
				return abci.ResponseDeliverTx{Code: spec.respCode}
			}
			ctx := sdk.Context{}
			cFn, capAddrs := CaptureSetPoEContractAddressFn()
			var capaturedParams types.Params
			m := PoEKeeperMock{
				SetPoEContractAddressFn: cFn,
				setParamsFn: func(ctx sdk.Context, params types.Params) {
					capaturedParams = params
				},
			}
			gotErr := InitGenesis(ctx, m, captureTx, spec.src, txConfig)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Len(t, capturedTxs, spec.expDeliveredGenTxCount)
			assert.Equal(t, spec.expContracts, *capAddrs)
			assert.Equal(t, spec.expParams, capaturedParams)
		})
	}
}

func TestDeliverGenTxs(t *testing.T) {
	initBech32Prefixes()

	txConfig := types.MakeEncodingConfig(t).TxConfig

	specs := map[string]struct {
		src               []json.RawMessage
		respCode          uint32
		expErr            bool
		expDeliveredCount int
	}{
		"valid tx": {
			src:               []json.RawMessage{[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"moniker-0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18katt8evmwr7g0w545g9kgrn2s6z9a0ky27gdp","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"P4qRtdI2pfl5IZN4cv28uuFhsRhSc/CBzrlB2/+ATQs="},"amount":{"denom":"utgd","amount":"100000000"},"vesting_amount":{"denom":"utgd","amount":"0"}}],"memo":"7973f9800a585f9a5e730ee18e4abab9a06214f5@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlOUwEtcgY5rV6cJHCJJntNdrY9Kpe057pY6yewFbhxW"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["cJnG18yHwWsgxjh1Kqf3j/MFv7OpX69c7VLQz1MX1qtndFIylSNPVbkXOMu2i+Ufy52nXH3yujOKsMIVLP62pg=="]}`)},
			expDeliveredCount: 1,
		},
		"multiple valid tx": {
			src: []json.RawMessage{
				[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"moniker-0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18katt8evmwr7g0w545g9kgrn2s6z9a0ky27gdp","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"P4qRtdI2pfl5IZN4cv28uuFhsRhSc/CBzrlB2/+ATQs="},"amount":{"denom":"utgd","amount":"100000000"},"vesting_amount":{"denom":"utgd","amount":"0"}}],"memo":"7973f9800a585f9a5e730ee18e4abab9a06214f5@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlOUwEtcgY5rV6cJHCJJntNdrY9Kpe057pY6yewFbhxW"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["cJnG18yHwWsgxjh1Kqf3j/MFv7OpX69c7VLQz1MX1qtndFIylSNPVbkXOMu2i+Ufy52nXH3yujOKsMIVLP62pg=="]}`),
				[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node001","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade15esst7gqck0uhmk9ppvllut55ayx4tte97f6an","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"39iF6mDX3F99NpqdS07bkpxWNQsWjhrXCdh6iYZ2em8="},"amount":{"denom":"utgd","amount":"0"},"vesting_amount":{"denom":"utgd","amount":"250000000"}}],"memo":"8bdd68de6f52e438f877915d550639a2e5a38a27@192.168.178.24:26656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A0sF4hqaQLQCH4l/vxAEyl/IUwcEJcQtkEiVu/i7gDP2"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":["0+lWFkn76yLiMldaK/nfeuByyctt48DW7QJp5SlQWTFcjEq7nd5mmwJxKH41Aidf0aTBH0iDBWJy7M3ZsMWKHQ=="]}`),
			},
			expDeliveredCount: 2,
		},
		"invalid json": {
			src:    []json.RawMessage{[]byte(`not a json string`)},
			expErr: true,
		},
		"not a tx type json": {
			src:    []json.RawMessage{[]byte(`{}`)},
			expErr: true,
		},
		"error returned for delivery failure": {
			src:      []json.RawMessage{[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18c4hs0jaxt3n0mlyc205zdea63rmfa26exytvg","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"xcxXzdJU7TIi6oxO/EOj09l54C/tKEwlDtDKgic10PA="},"value":{"denom":"utgd","amount":"100000000"}}],"memo":"a5ba21794f6f8a9167dcc49bc79a0b948c6ad386@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Av1jOQcVuiH/SlDOFBjnqzh7BH8nR6aFz92tcBbpXrmq"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["8saPPRQWq6WungLPNH0acgoar1eQWWCu3NmWKMDqHQxD7Ea3F4eRvEgFiK6K7g1WFmNnK4PTw0CNwRyErBEzLg=="]}`)},
			respCode: 1, // any other than `0` is considered an error
			expErr:   true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var capturedTxs []abci.RequestDeliverTx
			captureTx := func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
				capturedTxs = append(capturedTxs, tx)
				return abci.ResponseDeliverTx{Code: spec.respCode}
			}

			gotErr := DeliverGenTxs(spec.src, captureTx, txConfig)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Len(t, capturedTxs, spec.expDeliveredCount)
		})
	}
}

func initBech32Prefixes() {
	const Bech32Prefix = "tgrade"
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator, Bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(Bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus, Bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
}
