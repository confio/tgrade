package wasm

import (
	"github.com/confio/tgrade/x/twasm/types"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/keeper/poetesting"
	poetypes "github.com/confio/tgrade/x/poe/types"
)

func TestStakingQuerier(t *testing.T) {
	t.Log(types.RandomBech32Address(t))
	specs := map[string]struct {
		src     wasmvmtypes.StakingQuery
		mock    ViewKeeper
		expJson string
		expErr  bool
	}{
		"bonded denum": {
			src: wasmvmtypes.StakingQuery{BondedDenom: &struct{}{}},
			mock: ViewKeeperMock{GetBondDenomFn: func(ctx sdk.Context) string {
				return "alx"
			}},
			expJson: `{"denom": "alx"}`,
		},
		"all validators - single": {
			src: wasmvmtypes.StakingQuery{AllValidators: &wasmvmtypes.AllValidatorsQuery{}},
			mock: ViewKeeperMock{ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
				return poetesting.ValsetContractMock{
					ListValidatorsFn: func(ctx sdk.Context) ([]stakingtypes.Validator, error) {
						resp := []stakingtypes.Validator{
							poetypes.ValidatorFixture(func(m *stakingtypes.Validator) {
								m.OperatorAddress = "myOperatorAddress"
							}),
						}
						return resp, nil
					},
				}
			}},
			expJson: `{"validators":[{"address":"myOperatorAddress","commission":"0.000000000000000000","max_commission":"0.000000000000000000","max_change_rate":"0.000000000000000000"}]}`,
		},
		"all validators - multiple": {
			src: wasmvmtypes.StakingQuery{AllValidators: &wasmvmtypes.AllValidatorsQuery{}},
			mock: ViewKeeperMock{ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
				return poetesting.ValsetContractMock{
					ListValidatorsFn: func(ctx sdk.Context) ([]stakingtypes.Validator, error) {
						resp := []stakingtypes.Validator{
							poetypes.ValidatorFixture(func(m *stakingtypes.Validator) {
								m.OperatorAddress = "myOperatorAddress"
							}),
							poetypes.ValidatorFixture(func(m *stakingtypes.Validator) {
								m.OperatorAddress = "myOtherOperatorAddress"
							}),
						}
						return resp, nil
					},
				}
			}},
			expJson: `{"validators":[
{"address":"myOperatorAddress","commission":"0.000000000000000000","max_commission":"0.000000000000000000","max_change_rate":"0.000000000000000000"},
{"address":"myOtherOperatorAddress","commission":"0.000000000000000000","max_commission":"0.000000000000000000","max_change_rate":"0.000000000000000000"}
		]}`,
		},
		"query validator": {
			src: wasmvmtypes.StakingQuery{Validator: &wasmvmtypes.ValidatorQuery{Address: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
				return poetesting.ValsetContractMock{
					QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
						val := poetypes.ValidatorFixture(func(m *stakingtypes.Validator) {
							m.OperatorAddress = "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"
						})
						return &val, nil
					},
				}
			}},
			expJson: `{"validator":{"address":"cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3","commission":"0.000000000000000000","max_commission":"0.000000000000000000","max_change_rate":"0.000000000000000000"}}`,
		},
		"query validator - unknown address": {
			src: wasmvmtypes.StakingQuery{Validator: &wasmvmtypes.ValidatorQuery{Address: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{ValsetContractFn: func(ctx sdk.Context) keeper.ValsetContract {
				return poetesting.ValsetContractMock{
					QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
						return nil, nil
					},
				}
			}},
			expJson: `{"validator": null}`,
		},
		"query validator - invalid address": {
			src:    wasmvmtypes.StakingQuery{Validator: &wasmvmtypes.ValidatorQuery{Address: "not a valid address"}},
			expErr: true,
		},
		"all delegations": {
			src: wasmvmtypes.StakingQuery{AllDelegations: &wasmvmtypes.AllDelegationsQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{StakeContractFn: func(ctx sdk.Context) keeper.StakeContract {
				return poetesting.StakeContractMock{
					QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
						myValue := sdk.OneInt()
						return &myValue, nil
					},
				}
			},
				GetBondDenomFn: func(ctx sdk.Context) string {
					return "alx"
				}},
			expJson: `{"delegations":[{"delegator":"cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3","validator":"cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3","amount":{"denom":"alx","amount":"1"}}]}`,
		},
		"all delegations - unknown address": {
			src: wasmvmtypes.StakingQuery{AllDelegations: &wasmvmtypes.AllDelegationsQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{StakeContractFn: func(ctx sdk.Context) keeper.StakeContract {
				return poetesting.StakeContractMock{
					QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
						return nil, nil
					},
				}
			}},
			expJson: `{"delegations":[]}`,
		},
		"all delegations - invalid address": {
			src:    wasmvmtypes.StakingQuery{AllDelegations: &wasmvmtypes.AllDelegationsQuery{Delegator: "not a valid address"}},
			expErr: true,
		},
		"query delegation": {
			src: wasmvmtypes.StakingQuery{Delegation: &wasmvmtypes.DelegationQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3", Validator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{
				StakeContractFn: func(ctx sdk.Context) keeper.StakeContract {
					return poetesting.StakeContractMock{
						QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
							myValue := sdk.OneInt()
							return &myValue, nil
						},
					}
				},
				GetBondDenomFn: func(ctx sdk.Context) string {
					return "alx"
				},
				DistributionContractFn: func(ctx sdk.Context) keeper.DistributionContract {
					return poetesting.DistributionContractMock{ValidatorOutstandingRewardFn: func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
						return sdk.NewCoin("alx", sdk.NewInt(2)), nil
					}}
				},
			},
			expJson: `{
  "delegation": {
    "delegator": "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3",
    "validator": "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3",
    "amount": {
      "denom": "alx",
      "amount": "1"
    },
    "accumulated_rewards": [
      {
        "denom": "alx",
        "amount": "2"
      }
    ],
    "can_redelegate": {
      "denom": "alx",
      "amount": "1"
    }
  }
}
`,
		},
		"query delegation - address do not match - return empty result": {
			src: wasmvmtypes.StakingQuery{Delegation: &wasmvmtypes.DelegationQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3", Validator: "cosmos17emnuddq662fpxpnd43ch0396452d48vc8ufsw"}},
			mock: ViewKeeperMock{
				StakeContractFn: func(ctx sdk.Context) keeper.StakeContract {
					return poetesting.StakeContractMock{
						QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
							myValue := sdk.OneInt()
							return &myValue, nil
						},
					}
				},
				GetBondDenomFn: func(ctx sdk.Context) string {
					return "alx"
				},
				DistributionContractFn: func(ctx sdk.Context) keeper.DistributionContract {
					return poetesting.DistributionContractMock{ValidatorOutstandingRewardFn: func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
						return sdk.NewCoin("alx", sdk.NewInt(2)), nil
					}}
				},
			},
			expJson: `{}`,
		},
		"query delegation - unknown address return empty result": {
			src: wasmvmtypes.StakingQuery{Delegation: &wasmvmtypes.DelegationQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3", Validator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			mock: ViewKeeperMock{
				StakeContractFn: func(ctx sdk.Context) keeper.StakeContract {
					return poetesting.StakeContractMock{
						QueryStakedAmountFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
							return nil, nil
						},
					}
				},
				GetBondDenomFn: func(ctx sdk.Context) string {
					return "alx"
				},
			},
			expJson: `{}`,
		},
		"query delegation - invalid delegator address": {
			src:    wasmvmtypes.StakingQuery{Delegation: &wasmvmtypes.DelegationQuery{Delegator: "not a valid address", Validator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3"}},
			expErr: true,
		},
		"query delegation - invalid validator address": {
			src:    wasmvmtypes.StakingQuery{Delegation: &wasmvmtypes.DelegationQuery{Delegator: "cosmos1yq8zt83jznmp94jkj65yvfz9n52akmxt52ehm3", Validator: "not a valid address"}},
			expErr: true,
		},
		"unknown query": {
			src:    wasmvmtypes.StakingQuery{},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			q := StakingQuerier(spec.mock)
			gotRsp, gotErr := q(sdk.Context{}, &spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.JSONEq(t, spec.expJson, string(gotRsp), string(gotRsp))
		})
	}
}

type ViewKeeperMock struct {
	GetBondDenomFn         func(ctx sdk.Context) string
	DistributionContractFn func(ctx sdk.Context) keeper.DistributionContract
	ValsetContractFn       func(ctx sdk.Context) keeper.ValsetContract
	StakeContractFn        func(ctx sdk.Context) keeper.StakeContract
}

func (m ViewKeeperMock) GetBondDenom(ctx sdk.Context) string {
	if m.GetBondDenomFn == nil {
		panic("not expected to be called")
	}
	return m.GetBondDenomFn(ctx)
}

func (m ViewKeeperMock) DistributionContract(ctx sdk.Context) keeper.DistributionContract {
	if m.DistributionContractFn == nil {
		panic("not expected to be called")
	}
	return m.DistributionContractFn(ctx)
}

func (m ViewKeeperMock) ValsetContract(ctx sdk.Context) keeper.ValsetContract {
	if m.ValsetContractFn == nil {
		panic("not expected to be called")
	}
	return m.ValsetContractFn(ctx)
}

func (m ViewKeeperMock) StakeContract(ctx sdk.Context) keeper.StakeContract {
	if m.StakeContractFn == nil {
		panic("not expected to be called")
	}
	return m.StakeContractFn(ctx)
}
