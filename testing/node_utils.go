package testing

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/privval"
	"testing"
)

// load validator nodes consensus pub key from disk
func loadValidatorPubKey(t *testing.T, keyFile string) (cryptotypes.PubKey, string) {
	filePV := privval.LoadFilePVEmptyState(keyFile, "")
	pubKey, err := filePV.GetPubKey()
	require.NoError(t, err)
	valPubKey, err := cryptocodec.FromTmPubKeyInterface(pubKey)
	require.NoError(t, err)

	addr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKey)
	require.NoError(t, err)
	return valPubKey, addr
}
