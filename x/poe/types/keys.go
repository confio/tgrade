package types

const (
	// ModuleName is the name of the gentx module
	ModuleName = "poe"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = ModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = ModuleName
)

// nolint
var (
	ContractPrefix    = []byte{0x01}
	HistoricalInfoKey = []byte{0x02}
)
