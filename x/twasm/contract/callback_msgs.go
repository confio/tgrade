package contract

// TgradeSudoMsg callback message sent to a contract
type TgradeSudoMsg struct {
	PrivilegeChange *PrivilegeChangeMsg `json:"privilege_change,omitempty"`
}

/// These are called on a contract when it is made privileged or demoted
type PrivilegeChangeMsg struct {
	/// This is called when a contract gets "privileged status".
	/// It is a proper place to call `RegisterXXX` methods that require this status.
	/// Contracts that require this should be in a "frozen" state until they get this callback.
	Promoted *struct{} `json:"promoted,omitempty"`
	/// This is called when a contract looses "privileged status"
	Demoted *struct{} `json:"demoted,omitempty"`
}
