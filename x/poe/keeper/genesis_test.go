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

	var myAddr = types.RandomAccAddress()
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
				m.GenTxs = []json.RawMessage{[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18c4hs0jaxt3n0mlyc205zdea63rmfa26exytvg","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"xcxXzdJU7TIi6oxO/EOj09l54C/tKEwlDtDKgic10PA="},"value":{"denom":"utgd","amount":"100000000"}}],"memo":"a5ba21794f6f8a9167dcc49bc79a0b948c6ad386@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Av1jOQcVuiH/SlDOFBjnqzh7BH8nR6aFz92tcBbpXrmq"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["8saPPRQWq6WungLPNH0acgoar1eQWWCu3NmWKMDqHQxD7Ea3F4eRvEgFiK6K7g1WFmNnK4PTw0CNwRyErBEzLg=="]}`)}
				m.SystemAdminAddress = myAddr.String()
			},
			),
			expDeliveredGenTxCount: 1,
			expParams:              types.DefaultParams(),
		},
		"deliver genTX failed": {
			src: types.GenesisStateFixture(func(m *types.GenesisState) {
				m.GenTxs = []json.RawMessage{[]byte(`{}`)}
				m.SystemAdminAddress = myAddr.String()
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
			var capturedAdminAddr sdk.AccAddress
			var capaturedParams types.Params
			m := PoEKeeperMock{
				SetPoEContractAddressFn: cFn,
				setPoESystemAdminAddressFn: func(ctx sdk.Context, admin sdk.AccAddress) {
					capturedAdminAddr = admin
				},
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
			assert.Equal(t, spec.src.SystemAdminAddress, capturedAdminAddr.String())
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
			src:               []json.RawMessage{[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18c4hs0jaxt3n0mlyc205zdea63rmfa26exytvg","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"xcxXzdJU7TIi6oxO/EOj09l54C/tKEwlDtDKgic10PA="},"value":{"denom":"utgd","amount":"100000000"}}],"memo":"a5ba21794f6f8a9167dcc49bc79a0b948c6ad386@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Av1jOQcVuiH/SlDOFBjnqzh7BH8nR6aFz92tcBbpXrmq"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["8saPPRQWq6WungLPNH0acgoar1eQWWCu3NmWKMDqHQxD7Ea3F4eRvEgFiK6K7g1WFmNnK4PTw0CNwRyErBEzLg=="]}`)},
			expDeliveredCount: 1,
		},
		"multiple valid tx": {
			src: []json.RawMessage{
				[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18c4hs0jaxt3n0mlyc205zdea63rmfa26exytvg","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"xcxXzdJU7TIi6oxO/EOj09l54C/tKEwlDtDKgic10PA="},"value":{"denom":"utgd","amount":"100000000"}}],"memo":"a5ba21794f6f8a9167dcc49bc79a0b948c6ad386@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Av1jOQcVuiH/SlDOFBjnqzh7BH8nR6aFz92tcBbpXrmq"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["8saPPRQWq6WungLPNH0acgoar1eQWWCu3NmWKMDqHQxD7Ea3F4eRvEgFiK6K7g1WFmNnK4PTw0CNwRyErBEzLg=="]}`),
				[]byte(`{"body":{"messages":[{"@type":"/confio.poe.v1beta1.MsgCreateValidator","description":{"moniker":"node0","identity":"","website":"","security_contact":"","details":""},"operator_address":"tgrade18c4hs0jaxt3n0mlyc205zdea63rmfa26exytvg","pubkey":{"@type":"/cosmos.crypto.ed25519.PubKey","key":"xcxXzdJU7TIi6oxO/EOj09l54C/tKEwlDtDKgic10PA="},"value":{"denom":"utgd","amount":"100000000"}}],"memo":"a5ba21794f6f8a9167dcc49bc79a0b948c6ad386@192.168.178.24:16656","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Av1jOQcVuiH/SlDOFBjnqzh7BH8nR6aFz92tcBbpXrmq"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["8saPPRQWq6WungLPNH0acgoar1eQWWCu3NmWKMDqHQxD7Ea3F4eRvEgFiK6K7g1WFmNnK4PTw0CNwRyErBEzLg=="]}`),
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
