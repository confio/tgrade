package contract

import "testing"

// TG4MixerInitMsg contract init message
//See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-mixer/schema/instantiate_msg.json
type TG4MixerInitMsg struct {
	//Admin      string `json:"admin,omitempty"`
	LeftGroup  string `json:"left_group"`
	RightGroup string `json:"right_group"`
	Preauths   uint64 `json:"preauths,omitempty"`
}

func (m TG4MixerInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}
