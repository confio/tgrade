package types

const (
	EventTypeSetPrivileged     = "set_privileged_contract"
	EventTypeUnsetPrivileged   = "unset_privileged_contract"
	EventTypeRegisterPrivilege = "register_privilege"
	EventTypeReleasePrivilege  = "release_privilege"
	EventTypeMintTokens        = "mint"
)

const ( // event attributes
	AttributeKeyCallbackType = "privilege_type"
	AttributeKeyRecipient    = "recipient"
)
