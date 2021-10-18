package contract

import "testing"

// TG4MixerInitMsg contract init message
//See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-mixer/schema/instantiate_msg.json
type TG4MixerInitMsg struct {
	LeftGroup    string        `json:"left_group"`
	RightGroup   string        `json:"right_group"`
	Preauths     uint64        `json:"preauths,omitempty"`
	FunctionType MixerFunction `json:"function_type"`
}

type MixerFunction struct {
	GeometricMean    *struct{}         `json:"geometric_mean,omitempty"`
	Sigmoid          *Sigmoid          `json:"sigmoid,omitempty"`
	SigmoidSqrt      *SigmoidSqrt      `json:"sigmoid_sqrt,omitempty"`
	AlgebraicSigmoid *AlgebraicSigmoid `json:"algebaic_sigmoid,omitempty"`
}

type Sigmoid struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	P          Decimal `json:"p,string"`
	S          Decimal `json:"s,string"`
}

type SigmoidSqrt struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	S          Decimal `json:"s,string"`
}

type AlgebraicSigmoid struct {
	MaxRewards uint64  `json:"max_rewards,string"`
	A          Decimal `json:"a,string"`
	P          Decimal `json:"p,string"`
	S          Decimal `json:"s,string"`
}

func (m TG4MixerInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}
