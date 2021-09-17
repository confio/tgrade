package testing

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/privval"
	"path/filepath"
	"testing"
)

// load validator nodes consensus pub key for given node number
func loadValidatorPubKeyForNode(t *testing.T, sut *SystemUnderTest, nodeNumber int) (cryptotypes.PubKey, string) {
	return loadValidatorPubKey(t, filepath.Join(workDir, sut.nodePath(nodeNumber), "config", "priv_validator_key.json"))
}

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

// queryTendermintValidatorPower returns the validator's power from tendermint RPC endpoint. 0 when not found
func queryTendermintValidatorPower(t *testing.T, sut *SystemUnderTest, nodeNumber int) int64 {
	pubKey, _ := loadValidatorPubKeyForNode(t, sut, nodeNumber)
	valResult := NewTgradeCli(t, sut, false).GetTendermintValidatorSet()
	for _, v := range valResult.Validators {
		if v.PubKey.Equals(pubKey) {
			return v.VotingPower
		}
	}
	return 0
}
