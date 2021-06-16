package types

const (
	EventTypeSetPrivileged     = "set_privileged_contract"
	EventTypeUnsetPrivileged   = "unset_privileged_contract"
	EventTypeRegisterPrivilege = "register_privilege"
	EventTypeReleasePrivilege  = "release_privilege"
)
const ( // event attributes
	AttributeKeyCallbackType = "privilege_type"
)
