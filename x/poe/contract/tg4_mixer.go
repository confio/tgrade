package contract

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TG4MixerInitMsg contract init message
// See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-mixer/schema/instantiate_msg.json
type TG4MixerInitMsg struct {
	LeftGroup        string        `json:"left_group"`
	RightGroup       string        `json:"right_group"`
	PreAuthsHooks    uint64        `json:"preauths_hooks,omitempty"`
	PreAuthsSlashing uint64        `json:"preauths_slashing,omitempty"`
	FunctionType     MixerFunction `json:"function_type"`
}

type MixerFunction struct {
	GeometricMean    *struct{}         `json:"geometric_mean,omitempty"`
	Sigmoid          *Sigmoid          `json:"sigmoid,omitempty"`
	SigmoidSqrt      *SigmoidSqrt      `json:"sigmoid_sqrt,omitempty"`
	AlgebraicSigmoid *AlgebraicSigmoid `json:"algebaic_sigmoid,omitempty"`
}

type Sigmoid struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	P          sdk.Dec `json:"p"`
	S          sdk.Dec `json:"s"`
}

type SigmoidSqrt struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	S          sdk.Dec `json:"s"`
}

type AlgebraicSigmoid struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	A          sdk.Dec `json:"a"`
	P          sdk.Dec `json:"p"`
	S          sdk.Dec `json:"s"`
}

func (m TG4MixerInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}
