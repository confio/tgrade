package twasm

import (
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// ContractAddress
//
// Deprecated: will be public in wasmd
func ContractAddress(codeID, instanceID uint64) sdk.AccAddress {
	// NOTE: It is possible to get a duplicate address if either codeID or instanceID
	// overflow 32 bits. This is highly improbable, but something that could be refactored.
	contractID := codeID<<32 + instanceID
	addr := make([]byte, 20)
	addr[0] = 'C'
	binary.PutUvarint(addr[1:], contractID)
	return sdk.AccAddress(crypto.AddressHash(addr))
}
