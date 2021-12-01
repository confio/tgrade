package types

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func ConsensusParamsFixture(mutators ...func(*abci.ConsensusParams)) *abci.ConsensusParams {
	r := &abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxBytes: 200000,
			MaxGas:   2000000,
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: 302400,
			MaxAgeDuration:  504 * time.Hour,
			MaxBytes:        10000,
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: []string{
				tmtypes.ABCIPubKeyTypeEd25519,
			},
		},
	}
	for _, m := range mutators {
		m(r)
	}
	return r
}
