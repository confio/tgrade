package keeper

// nolint
var (
	// leave enough space to not conflict with wasm prefixes

	privilegedContractsSecondaryIndexPrefix = []byte{0xa0}
	contractCallbacksSecondaryIndexPrefix   = []byte{0xa1}
)
