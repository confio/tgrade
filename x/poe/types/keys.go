package types

import genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

const (
	// ModuleName is the name of the gentx module
	ModuleName = genutiltypes.ModuleName // todo (Alex): rename to POE

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName
)

// nolint
var (
	PoEContractPrefix = []byte{0x01}
)
