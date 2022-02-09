package keeper

import (
	"fmt"
	"testing"

	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
)

func TestTgradeHandlesDispatchMsg(t *testing.T) {
	var (
		contractAddr = RandomAddress(t)
		otherAddr    = RandomAddress(t)
	)
	specs := map[string]struct {
		setup                 func(m *handlerTgradeKeeperMock)
		src                   wasmvmtypes.CosmosMsg
		expErr                *sdkerrors.Error
		expCapturedGovContent []govtypes.Content
		expEvents             []sdk.Event
	}{
		"handle privilege msg": {
			src: wasmvmtypes.CosmosMsg{
				Custom: []byte(`{"privilege":{"request":"begin_blocker"}}`),
			},
			setup: func(m *handlerTgradeKeeperMock) {
				setupHandlerKeeperMock(m)
				m.GetContractInfoFn = emitCtxEventWithGetContractInfoFn(m.GetContractInfoFn, sdk.NewEvent("testing"))
			},
			expEvents: sdk.Events{sdk.NewEvent("testing")},
		},
		"handle execute gov proposal msg": {
			src: wasmvmtypes.CosmosMsg{
				Custom: []byte(`{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"text":{}}}}`),
			},
			setup: func(m *handlerTgradeKeeperMock) {
				setupHandlerKeeperMock(m, withPrivilegeSet(t, types.PrivilegeTypeGovProposalExecutor))
				m.GetContractInfoFn = emitCtxEventWithGetContractInfoFn(m.GetContractInfoFn, sdk.NewEvent("testing"))
			},
			expCapturedGovContent: []govtypes.Content{&govtypes.TextProposal{Title: "foo", Description: "bar"}},
			expEvents:             sdk.Events{sdk.NewEvent("testing")},
		},
		"handle minter msg": {
			src: wasmvmtypes.CosmosMsg{
				Custom: []byte(fmt.Sprintf(`{"mint_tokens":{"amount":"1","denom":"utgd","recipient":%q}}`, otherAddr.String())),
			},
			setup: func(m *handlerTgradeKeeperMock) {
				setupHandlerKeeperMock(m, withPrivilegeSet(t, types.PrivilegeTypeTokenMinter))
			},
			expEvents: sdk.Events{sdk.NewEvent(
				types.EventTypeMintTokens,
				sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, "1utgd"),
				sdk.NewAttribute(types.AttributeKeyRecipient, otherAddr.String()),
			)},
		},
		"handle consensus params change msg": {
			src: wasmvmtypes.CosmosMsg{
				Custom: []byte(`{"consensus_params":{"block":{"max_gas":100000000}}}`),
			},
			setup: func(m *handlerTgradeKeeperMock) {
				setupHandlerKeeperMock(m, withPrivilegeSet(t, types.PrivilegeConsensusParamChanger))
			},
		},
		"non custom msg rejected": {
			src:    wasmvmtypes.CosmosMsg{},
			setup:  func(m *handlerTgradeKeeperMock) {},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"non privileged contracts rejected": {
			src: wasmvmtypes.CosmosMsg{Custom: []byte(`{}`)},
			setup: func(m *handlerTgradeKeeperMock) {
				m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
					return false
				}
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"invalid json rejected": {
			src: wasmvmtypes.CosmosMsg{Custom: []byte(`not json`)},
			setup: func(m *handlerTgradeKeeperMock) {
				m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
					return true
				}
			},
			expErr: sdkerrors.ErrJSONUnmarshal,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cdc := MakeEncodingConfig(t).Codec
			govRouter := &CapturingGovRouter{}
			minterMock := NoopMinterMock()
			mock := handlerTgradeKeeperMock{}
			consensusStoreMock := NoopConsensusParamsStoreMock()
			spec.setup(&mock)
			h := NewTgradeHandler(cdc, mock, minterMock, consensusStoreMock, govRouter)
			em := sdk.NewEventManager()
			ctx := sdk.Context{}.WithEventManager(em)

			// when
			gotEvents, _, gotErr := h.DispatchMsg(ctx, contractAddr, "", spec.src)
			// then
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			assert.Equal(t, spec.expCapturedGovContent, govRouter.captured)
			assert.Equal(t, spec.expEvents, gotEvents)
			assert.Empty(t, em.Events())
		})
	}
}

func withPrivilegeSet(t *testing.T, p types.PrivilegeType) func(info *wasmtypes.ContractInfo) {
	return func(info *wasmtypes.ContractInfo) {
		var details types.TgradeContractDetails
		require.NoError(t, info.ReadExtension(&details))
		details.AddRegisteredPrivilege(p, 1)
		require.NoError(t, info.SetExtension(&details))
	}
}

func emitCtxEventWithGetContractInfoFn(fn func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo, event sdk.Event) func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
	return func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
		ctx.EventManager().EmitEvent(event)
		return fn(ctx, contractAddress)
	}
}

type registration struct {
	cb   types.PrivilegeType
	addr sdk.AccAddress
}
type unregistration struct {
	cb   types.PrivilegeType
	pos  uint8
	addr sdk.AccAddress
}

func TestTgradeHandlesPrivilegeMsg(t *testing.T) {
	myContractAddr := RandomAddress(t)

	var capturedDetails *types.TgradeContractDetails
	captureContractDetails := func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
		require.Equal(t, myContractAddr, contract)
		capturedDetails = details
		return nil
	}

	var capturedRegistrations []registration
	captureRegistrations := func(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error) {
		capturedRegistrations = append(capturedRegistrations, registration{cb: privilegeType, addr: contractAddress})
		return 1, nil
	}
	var capturedUnRegistrations []unregistration
	captureUnRegistrations := func(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddress sdk.AccAddress) bool {
		capturedUnRegistrations = append(capturedUnRegistrations, unregistration{cb: privilegeType, pos: pos, addr: contractAddress})
		return true
	}

	captureWithMock := func(mutators ...func(*wasmtypes.ContractInfo)) func(mock *handlerTgradeKeeperMock) {
		return func(m *handlerTgradeKeeperMock) {
			m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
				f := wasmtypes.ContractInfoFixture(mutators...)
				return &f
			}
			m.setContractDetailsFn = captureContractDetails
			m.appendToPrivilegedContractsFn = captureRegistrations
			m.removePrivilegeRegistrationFn = captureUnRegistrations
		}
	}

	specs := map[string]struct {
		setup              func(m *handlerTgradeKeeperMock)
		src                contract.PrivilegeMsg
		expDetails         *types.TgradeContractDetails
		expRegistrations   []registration
		expUnRegistrations []unregistration
		expErr             *sdkerrors.Error
	}{
		"register begin block": {
			src:   contract.PrivilegeMsg{Request: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
			},
			expRegistrations: []registration{{cb: types.PrivilegeTypeBeginBlock, addr: myContractAddr}},
		},
		"unregister begin block": {
			src: contract.PrivilegeMsg{Release: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredPrivileges: []types.RegisteredPrivilege{}},
			expUnRegistrations: []unregistration{{cb: types.PrivilegeTypeBeginBlock, pos: 1, addr: myContractAddr}},
		},
		"register end block": {
			src:   contract.PrivilegeMsg{Request: types.PrivilegeTypeEndBlock},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "end_blocker"}},
			},
			expRegistrations: []registration{{cb: types.PrivilegeTypeEndBlock, addr: myContractAddr}},
		},
		"unregister end block": {
			src: contract.PrivilegeMsg{Release: types.PrivilegeTypeEndBlock},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "end_blocker"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredPrivileges: []types.RegisteredPrivilege{}},
			expUnRegistrations: []unregistration{{cb: types.PrivilegeTypeEndBlock, pos: 1, addr: myContractAddr}},
		},
		"register validator set update block": {
			src:   contract.PrivilegeMsg{Request: types.PrivilegeTypeValidatorSetUpdate},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "validator_set_updater"}},
			},
			expRegistrations: []registration{{cb: types.PrivilegeTypeValidatorSetUpdate, addr: myContractAddr}},
		},
		"unregister validator set update block": {
			src: contract.PrivilegeMsg{Release: types.PrivilegeTypeValidatorSetUpdate},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "validator_set_updater"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredPrivileges: []types.RegisteredPrivilege{}},
			expUnRegistrations: []unregistration{{cb: types.PrivilegeTypeValidatorSetUpdate, pos: 1, addr: myContractAddr}},
		},
		"register gov proposal executor": {
			src:   contract.PrivilegeMsg{Request: types.PrivilegeTypeGovProposalExecutor},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "gov_proposal_executor"}},
			},
			expRegistrations: []registration{{cb: types.PrivilegeTypeGovProposalExecutor, addr: myContractAddr}},
		},
		"unregister gov proposal executor": {
			src: contract.PrivilegeMsg{Release: types.PrivilegeTypeGovProposalExecutor},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "gov_proposal_executor"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredPrivileges: []types.RegisteredPrivilege{}},
			expUnRegistrations: []unregistration{{cb: types.PrivilegeTypeGovProposalExecutor, pos: 1, addr: myContractAddr}},
		},
		"register privilege fails": {
			src: contract.PrivilegeMsg{Request: types.PrivilegeTypeValidatorSetUpdate},
			setup: func(m *handlerTgradeKeeperMock) {
				m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					r := wasmtypes.ContractInfoFixture()
					return &r
				}
				m.appendToPrivilegedContractsFn = func(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error) {
					return 0, wasmtypes.ErrDuplicate
				}
			},
			expErr: wasmtypes.ErrDuplicate,
		},
		"register begin block with existing registration": {
			src: contract.PrivilegeMsg{Request: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				info.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
				})
			}),
		},
		"register appends to existing callback list": {
			src: contract.PrivilegeMsg{Request: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				info.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 100, PrivilegeType: "end_blocker"}},
				})
			}),
			expDetails: &types.TgradeContractDetails{
				RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 100, PrivilegeType: "end_blocker"}, {Position: 1, PrivilegeType: "begin_blocker"}},
			},
			expRegistrations: []registration{{cb: types.PrivilegeTypeBeginBlock, addr: myContractAddr}},
		},
		"unregister removed from existing callback list": {
			src: contract.PrivilegeMsg{Release: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{
						{Position: 3, PrivilegeType: "validator_set_updater"},
						{Position: 1, PrivilegeType: "begin_blocker"},
						{Position: 100, PrivilegeType: "end_blocker"},
					},
				}
				info.SetExtension(ext)
			}),
			expDetails: &types.TgradeContractDetails{RegisteredPrivileges: []types.RegisteredPrivilege{
				{Position: 3, PrivilegeType: "validator_set_updater"},
				{Position: 100, PrivilegeType: "end_blocker"},
			}},
			expUnRegistrations: []unregistration{{cb: types.PrivilegeTypeBeginBlock, pos: 1, addr: myContractAddr}},
		},

		"unregister begin block with without existing registration": {
			src:   contract.PrivilegeMsg{Release: types.PrivilegeTypeBeginBlock},
			setup: captureWithMock(),
		},
		"empty privilege msg rejected": {
			setup: func(m *handlerTgradeKeeperMock) {
				setupHandlerKeeperMock(m)
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedDetails, capturedRegistrations, capturedUnRegistrations = nil, nil, nil
			mock := handlerTgradeKeeperMock{}
			spec.setup(&mock)
			h := NewTgradeHandler(nil, mock, nil, nil, nil)
			var ctx sdk.Context
			gotErr := h.handlePrivilege(ctx, myContractAddr, &spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			if spec.expErr != nil {
				return
			}
			assert.Equal(t, spec.expDetails, capturedDetails)
			assert.Equal(t, spec.expRegistrations, capturedRegistrations)
			assert.Equal(t, spec.expUnRegistrations, capturedUnRegistrations)
		})
	}
}

func TestHandleGovProposalExecution(t *testing.T) {
	myContractAddr := RandomAddress(t)
	specs := map[string]struct {
		src                   contract.ExecuteGovProposal
		setup                 func(m *handlerTgradeKeeperMock)
		expErr                *sdkerrors.Error
		expCapturedGovContent []govtypes.Content
	}{
		"all good": {
			src:                   contract.ExecuteGovProposalFixture(),
			setup:                 withPrivilegeRegistered(types.PrivilegeTypeGovProposalExecutor),
			expCapturedGovContent: []govtypes.Content{&govtypes.TextProposal{Title: "foo", Description: "bar"}},
		},
		"non consensus params accepted": {
			src: contract.ExecuteGovProposalFixture(func(p *contract.ExecuteGovProposal) {
				p.Proposal = contract.GovProposalFixture(func(x *contract.GovProposal) {
					x.ChangeParams = &[]proposaltypes.ParamChange{
						{Subspace: "foo", Key: "bar", Value: `{"example": "value"}`},
					}
				})
			}),
			setup: withPrivilegeRegistered(types.PrivilegeTypeGovProposalExecutor),
			expCapturedGovContent: []govtypes.Content{&proposaltypes.ParameterChangeProposal{
				Title:       "foo",
				Description: "bar",
				Changes: []proposaltypes.ParamChange{
					{Subspace: "foo", Key: "bar", Value: `{"example": "value"}`}},
			}},
		},
		"unauthorized contract": {
			src: contract.ExecuteGovProposalFixture(),
			setup: func(m *handlerTgradeKeeperMock) {
				m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					c := wasmtypes.ContractInfoFixture()
					return &c
				}
			},
			expErr: sdkerrors.ErrUnauthorized,
		},
		"invalid content": {
			src: contract.ExecuteGovProposalFixture(func(p *contract.ExecuteGovProposal) {
				p.Proposal = contract.GovProposalFixture(func(x *contract.GovProposal) {
					x.RegisterUpgrade = &upgradetypes.Plan{}
				})
			}),
			setup:  withPrivilegeRegistered(types.PrivilegeTypeGovProposalExecutor),
			expErr: sdkerrors.ErrInvalidRequest,
		},
		"no content": {
			src:    contract.ExecuteGovProposal{Title: "foo", Description: "bar"},
			setup:  withPrivilegeRegistered(types.PrivilegeTypeGovProposalExecutor),
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"unknown origin contract": {
			src: contract.ExecuteGovProposalFixture(),
			setup: func(m *handlerTgradeKeeperMock) {
				m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					return nil
				}
			},
			expErr: wasmtypes.ErrNotFound,
		},
		"consensus params rejected": {
			src: contract.ExecuteGovProposalFixture(func(p *contract.ExecuteGovProposal) {
				p.Proposal = contract.GovProposalFixture(func(x *contract.GovProposal) {
					x.ChangeParams = &[]proposaltypes.ParamChange{
						{
							Subspace: "baseapp",
							Key:      "BlockParams",
							Value:    `{"max_bytes": "1"}`,
						},
					}
				})
			}),
			setup:  withPrivilegeRegistered(types.PrivilegeTypeGovProposalExecutor),
			expErr: sdkerrors.ErrUnauthorized,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cdc := MakeEncodingConfig(t).Codec
			mock := handlerTgradeKeeperMock{}
			spec.setup(&mock)
			router := &CapturingGovRouter{}
			h := NewTgradeHandler(cdc, mock, nil, nil, router)
			var ctx sdk.Context
			gotErr := h.handleGovProposalExecution(ctx, myContractAddr, &spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			assert.Equal(t, spec.expCapturedGovContent, router.captured)
		})
	}

}

func TestHandleMintToken(t *testing.T) {
	myContractAddr := RandomAddress(t)
	myRecipientAddr := RandomAddress(t)
	specs := map[string]struct {
		src            contract.MintTokens
		setup          func(k *handlerTgradeKeeperMock)
		expErr         *sdkerrors.Error
		expMintedCoins sdk.Coins
		expRecipient   sdk.AccAddress
	}{
		"all good": {
			src: contract.MintTokens{
				Denom:         "foo",
				Amount:        "123",
				RecipientAddr: myRecipientAddr.String(),
			},
			setup:          withPrivilegeRegistered(types.PrivilegeTypeTokenMinter),
			expMintedCoins: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(123))),
			expRecipient:   myRecipientAddr,
		},
		"unauthorized contract": {
			src: contract.MintTokens{
				Denom:         "foo",
				Amount:        "123",
				RecipientAddr: myRecipientAddr.String(),
			},
			setup: func(k *handlerTgradeKeeperMock) {
				k.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					c := wasmtypes.ContractInfoFixture(func(info *wasmtypes.ContractInfo) {
						info.SetExtension(&types.TgradeContractDetails{
							RegisteredPrivileges: []types.RegisteredPrivilege{},
						})
					})
					return &c
				}
			},
			expErr: sdkerrors.ErrUnauthorized,
		},
		"invalid denom": {
			src: contract.MintTokens{
				Denom:         "&&&foo",
				Amount:        "123",
				RecipientAddr: myRecipientAddr.String(),
			},
			setup:  withPrivilegeRegistered(types.PrivilegeTypeTokenMinter),
			expErr: sdkerrors.ErrInvalidCoins,
		},
		"invalid amount": {
			src: contract.MintTokens{
				Denom:         "foo",
				Amount:        "not-a-number",
				RecipientAddr: myRecipientAddr.String(),
			},
			setup:  withPrivilegeRegistered(types.PrivilegeTypeTokenMinter),
			expErr: sdkerrors.ErrInvalidCoins,
		},
		"invalid recipient": {
			src: contract.MintTokens{
				Denom:         "foo",
				Amount:        "123",
				RecipientAddr: "not-an-address",
			},
			setup: func(k *handlerTgradeKeeperMock) {
				k.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					c := wasmtypes.ContractInfoFixture(func(info *wasmtypes.ContractInfo) {
						info.SetExtension(&types.TgradeContractDetails{
							RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "token_minter"}},
						})
					})
					return &c
				}
			},
			expErr: sdkerrors.ErrInvalidAddress,
		},
		"no content": {
			src:    contract.MintTokens{},
			setup:  withPrivilegeRegistered(types.PrivilegeTypeTokenMinter),
			expErr: sdkerrors.ErrInvalidAddress,
		},
		"unknown origin contract": {
			src: contract.MintTokens{
				Denom:         "foo",
				Amount:        "123",
				RecipientAddr: "not-an-address",
			},
			setup: func(m *handlerTgradeKeeperMock) {
				m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					return nil
				}
			},
			expErr: wasmtypes.ErrNotFound,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cdc := MakeEncodingConfig(t).Codec
			mintFn, capturedMintedCoins := CaptureMintedCoinsFn()
			sendFn, capturedSentCoins := CaptureSendCoinsFn()
			mock := MinterMock{MintCoinsFn: mintFn, SendCoinsFromModuleToAccountFn: sendFn}
			keeperMock := handlerTgradeKeeperMock{}
			spec.setup(&keeperMock)
			h := NewTgradeHandler(cdc, keeperMock, mock, nil, nil)
			var ctx sdk.Context
			gotEvts, gotErr := h.handleMintToken(ctx, myContractAddr, &spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			if spec.expErr != nil {
				assert.Len(t, gotEvts, 0)
				return
			}
			require.Len(t, *capturedMintedCoins, 1)
			assert.Equal(t, spec.expMintedCoins, (*capturedMintedCoins)[0])
			require.Len(t, *capturedSentCoins, 1)
			assert.Equal(t, (*capturedSentCoins)[0].coins, spec.expMintedCoins)
			assert.Equal(t, (*capturedSentCoins)[0].recipientAddr, spec.expRecipient)
			require.Len(t, gotEvts, 1)
			assert.Equal(t, types.EventTypeMintTokens, gotEvts[0].Type)
		})
	}
}

func TestHandleConsensusParamsUpdate(t *testing.T) {
	var (
		myContractAddr = RandomAddress(t)
		// some integers
		one, two, three, four, five int64 = 1, 2, 3, 4, 5
	)
	specs := map[string]struct {
		src       contract.ConsensusParamsUpdate
		setup     func(k *handlerTgradeKeeperMock)
		expErr    *sdkerrors.Error
		expStored *abci.ConsensusParams
	}{
		"all good": {
			src: contract.ConsensusParamsUpdate{
				Block: &contract.BlockParams{
					MaxBytes: &one,
					MaxGas:   &two,
				},
				Evidence: &contract.EvidenceParams{
					MaxAgeNumBlocks: &three,
					MaxAgeDuration:  &four,
					MaxBytes:        &five,
				},
			},
			setup: withPrivilegeRegistered(types.PrivilegeConsensusParamChanger),
			expStored: types.ConsensusParamsFixture(func(c *abci.ConsensusParams) {
				c.Block.MaxBytes = 1
				c.Block.MaxGas = 2
				c.Evidence.MaxAgeNumBlocks = 3
				c.Evidence.MaxAgeDuration = 4 * 1_000_000_000 // nanos
				c.Evidence.MaxBytes = 5
			}),
		},
		"unauthorized": {
			src: contract.ConsensusParamsUpdate{
				Evidence: &contract.EvidenceParams{
					MaxAgeNumBlocks: &one,
				},
			},
			setup: func(k *handlerTgradeKeeperMock) {
				k.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
					c := wasmtypes.ContractInfoFixture()
					return &c
				}
			},
			expErr: sdkerrors.ErrUnauthorized,
		},
		"invalid msg": {
			src:    contract.ConsensusParamsUpdate{},
			setup:  withPrivilegeRegistered(types.PrivilegeConsensusParamChanger),
			expErr: wasmtypes.ErrEmpty,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			cdc := MakeEncodingConfig(t).Codec
			var gotStored *abci.ConsensusParams
			mock := ConsensusParamsStoreMock{
				GetConsensusParamsFn:   func(ctx sdk.Context) *abci.ConsensusParams { return types.ConsensusParamsFixture() },
				StoreConsensusParamsFn: func(ctx sdk.Context, cp *abci.ConsensusParams) { gotStored = cp },
			}

			keeperMock := handlerTgradeKeeperMock{}
			spec.setup(&keeperMock)
			h := NewTgradeHandler(cdc, keeperMock, nil, mock, nil)
			var ctx sdk.Context
			gotEvts, gotErr := h.handleConsensusParamsUpdate(ctx, myContractAddr, &spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			assert.Len(t, gotEvts, 0)
			assert.Equal(t, spec.expStored, gotStored)
		})
	}
}

func withPrivilegeRegistered(p types.PrivilegeType) func(k *handlerTgradeKeeperMock) {
	return func(k *handlerTgradeKeeperMock) {
		k.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
			c := wasmtypes.ContractInfoFixture(func(info *wasmtypes.ContractInfo) {
				info.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: p.String()}},
				})
			})
			return &c
		}
	}
}

// setupHandlerKeeperMock provided method stubs for all methods for registration
func setupHandlerKeeperMock(m *handlerTgradeKeeperMock, mutators ...func(*wasmtypes.ContractInfo)) {
	m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
		return true
	}
	m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
		v := wasmtypes.ContractInfoFixture(append([]func(*wasmtypes.ContractInfo){func(info *wasmtypes.ContractInfo) {
			info.SetExtension(&types.TgradeContractDetails{})
		}}, mutators...)...)
		return &v
	}
	m.appendToPrivilegedContractsFn = func(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error) {
		return 1, nil
	}
	m.setContractDetailsFn = func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
		return nil
	}
}

var _ TgradeWasmHandlerKeeper = handlerTgradeKeeperMock{}

type handlerTgradeKeeperMock struct {
	IsPrivilegedFn                func(ctx sdk.Context, contract sdk.AccAddress) bool
	appendToPrivilegedContractsFn func(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error)
	removePrivilegeRegistrationFn func(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddr sdk.AccAddress) bool
	setContractDetailsFn          func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error
	GetContractInfoFn             func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

func (m handlerTgradeKeeperMock) IsPrivileged(ctx sdk.Context, contract sdk.AccAddress) bool {
	if m.IsPrivilegedFn == nil {
		panic("not expected to be called")
	}
	return m.IsPrivilegedFn(ctx, contract)
}

func (m handlerTgradeKeeperMock) appendToPrivilegedContracts(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error) {
	if m.appendToPrivilegedContractsFn == nil {
		panic("not expected to be called")
	}
	return m.appendToPrivilegedContractsFn(ctx, privilegeType, contractAddress)
}

func (m handlerTgradeKeeperMock) removePrivilegeRegistration(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddr sdk.AccAddress) bool {
	if m.removePrivilegeRegistrationFn == nil {
		panic("not expected to be called")
	}
	return m.removePrivilegeRegistrationFn(ctx, privilegeType, pos, contractAddr)
}

func (m handlerTgradeKeeperMock) setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
	if m.setContractDetailsFn == nil {
		panic("not expected to be called")
	}
	return m.setContractDetailsFn(ctx, contract, details)
}

func (m handlerTgradeKeeperMock) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
	if m.GetContractInfoFn == nil {
		panic("not expected to be called")
	}
	return m.GetContractInfoFn(ctx, contractAddress)
}

// MinterMock test helper that satisfies the `minter` interface
type MinterMock struct {
	MintCoinsFn                    func(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccountFn func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

func (m MinterMock) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	if m.MintCoinsFn == nil {
		panic("not expected to be called")
	}
	return m.MintCoinsFn(ctx, moduleName, amt)
}

func (m MinterMock) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.SendCoinsFromModuleToAccountFn == nil {
		panic("not expected to be called")
	}
	return m.SendCoinsFromModuleToAccountFn(ctx, senderModule, recipientAddr, amt)
}

func NoopMinterMock() *MinterMock {
	return &MinterMock{
		MintCoinsFn: func(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
			return nil
		},
		SendCoinsFromModuleToAccountFn: func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
			return nil
		},
	}
}

func CaptureMintedCoinsFn() (func(ctx sdk.Context, moduleName string, amt sdk.Coins) error, *[]sdk.Coins) {
	var r []sdk.Coins
	return func(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
		r = append(r, amt)
		return nil
	}, &r
}

type capturedSendCoins struct {
	recipientAddr sdk.AccAddress
	coins         sdk.Coins
}

func CaptureSendCoinsFn() (func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error, *[]capturedSendCoins) {
	var r []capturedSendCoins
	return func(ctx sdk.Context, moduleName string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
		r = append(r, capturedSendCoins{recipientAddr: recipientAddr, coins: amt})
		return nil
	}, &r
}

type ConsensusParamsStoreMock struct {
	GetConsensusParamsFn   func(ctx sdk.Context) *abci.ConsensusParams
	StoreConsensusParamsFn func(ctx sdk.Context, cp *abci.ConsensusParams)
}

func NoopConsensusParamsStoreMock() ConsensusParamsStoreMock {
	return ConsensusParamsStoreMock{
		GetConsensusParamsFn: func(ctx sdk.Context) *abci.ConsensusParams {
			return types.ConsensusParamsFixture()
		},
		StoreConsensusParamsFn: func(ctx sdk.Context, cp *abci.ConsensusParams) {},
	}
}

func (m ConsensusParamsStoreMock) GetConsensusParams(ctx sdk.Context) *abci.ConsensusParams {
	if m.GetConsensusParamsFn == nil {
		panic("not expected to be called")
	}
	return m.GetConsensusParamsFn(ctx)
}

func (m ConsensusParamsStoreMock) StoreConsensusParams(ctx sdk.Context, cp *abci.ConsensusParams) {
	if m.StoreConsensusParamsFn == nil {
		panic("not expected to be called")
	}
	m.StoreConsensusParamsFn(ctx, cp)
}
