package types

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func MakeEncodingConfig(_ testing.TB) simappparams.EncodingConfig {
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(legacyAmino)

	RegisterInterfaces(interfaceRegistry)
	RegisterLegacyAminoCodec(legacyAmino)

	return simappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          authtx.NewTxConfig(marshaler, authtx.DefaultSignModes),
		Amino:             legacyAmino,
	}
}

func RandomAccAddress() sdk.AccAddress {
	return rand.Bytes(sdk.AddrLen)
}

// RandomGenTX returns a signed genesis tx
func RandomGenTX(t *testing.T) (json.RawMessage, sdk.AccAddress) {
	t.Helper()
	nodeConfig := cfg.TestConfig()
	nodeConfig.RootDir = t.TempDir()
	nodeConfig.NodeKey = "key.json"
	_, valPubKey, err := genutil.InitializeNodeValidatorFiles(nodeConfig)
	require.NoError(t, err)
	//setup keyring
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, t.TempDir(), nil)
	require.NoError(t, err)
	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	require.NoError(t, err)
	const myKey = "myKey"
	addr, _, err := server.GenerateSaveCoinKey(kb, myKey, true, algo)
	require.NoError(t, err)

	// prepare genesis tx
	valTokens := sdk.TokensFromConsensusPower(100)
	createValMsg, err := NewMsgCreateValidator(
		sdk.ValAddress(addr),
		valPubKey,
		sdk.NewCoin(DefaultBondDenom, valTokens),
		stakingtypes.NewDescription("testing", "", "", "", ""),
	)
	require.NoError(t, err)
	txConfig := MakeEncodingConfig(t).TxConfig
	txBuilder := txConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(createValMsg)
	require.NoError(t, err)

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID("").
		WithKeybase(kb).
		WithTxConfig(txConfig)

	err = tx.Sign(txFactory, myKey, txBuilder, true)
	require.NoError(t, err)

	txBz, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	return txBz, addr
}
