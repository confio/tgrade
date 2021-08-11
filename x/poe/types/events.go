package types

// staking module event types
const (
	EventTypeCreateValidator = "create_validator"
	EventTypeUpdateValidator = "update_validator"

	AttributeKeyValOperator = "operator"
	AttributeKeyMoniker     = "moniker"
	AttributeKeyPubKeyHex   = "pubkey"
	AttributeValueCategory  = ModuleName
)
