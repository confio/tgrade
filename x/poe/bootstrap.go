package poe

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

var (
	//go:embed contract/tg4_engagement.wasm
	tg4Engagement []byte
	//go:embed contract/tg4_stake.wasm
	tg4Stake []byte
	//go:embed contract/tg4_mixer.wasm
	tg4Mixer []byte
	//go:embed contract/tgrade_valset.wasm
	tgValset []byte
	//go:embed contract/tgrade_trusted_circle.wasm
	tgTrustedCircles []byte
	//go:embed contract/tgrade_oc_proposals.wasm
	tgOCGovProposalsCircles []byte
	//go:embed contract/tgrade_community_pool.wasm
	tgCommunityPool []byte
	//go:embed contract/tgrade_validator_voting.wasm
	tgValidatorVoting []byte
	//go:embed contract/version.txt
	contractVersion []byte
)

// ClearEmbeddedContracts release memory
func ClearEmbeddedContracts() {
	tg4Engagement = nil
	tg4Stake = nil
	tg4Mixer = nil
	tgValset = nil
	tgTrustedCircles = nil
	tgOCGovProposalsCircles = nil
	tgCommunityPool = nil
	tgValidatorVoting = nil
}

type poeKeeper interface {
	keeper.ContractSource
	SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
	ValsetContract(ctx sdk.Context) keeper.ValsetContract
	EngagementContract(ctx sdk.Context) keeper.EngagementContract
}

// bootstrapPoEContracts stores and instantiates all PoE contracts:
// See https://github.com/confio/tgrade-contracts/blob/main/docs/Architecture.md#multi-level-governance for an overview
func bootstrapPoEContracts(ctx sdk.Context, k wasmtypes.ContractOpsKeeper, tk twasmKeeper, poeKeeper poeKeeper, gs types.GenesisState) error {
	systemAdminAddr, err := sdk.AccAddressFromBech32(gs.SystemAdminAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "system admin")
	}

	// setup engagement contract
	tg4EngagementInitMsg := newEngagementInitMsg(gs, systemAdminAddr)
	engagementCodeID, err := k.Create(ctx, systemAdminAddr, tg4Engagement, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 engagement contract")
	}
	engagementContractAddr, _, err := k.Instantiate(ctx, engagementCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(tg4EngagementInitMsg), "engagement", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 engagement")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeEngagement, engagementContractAddr)
	if err := k.PinCode(ctx, engagementCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 engagement contract")
	}
	logger := keeper.ModuleLogger(ctx)
	logger.Info("engagement group contract", "address", engagementContractAddr, "code_id", engagementCodeID)

	// setup trusted circle for oversight community
	ocCodeID, err := k.Create(ctx, systemAdminAddr, tgTrustedCircles, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg trusted circle contract")
	}
	ocInitMsg := newOCInitMsg(gs)
	deposit := sdk.NewCoins(gs.OversightCommitteeContractConfig.EscrowAmount)
	ocContractAddr, _, err := k.Instantiate(ctx, ocCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(ocInitMsg), "oversight_committee", deposit)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg trusted circle contract")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeOversightCommunity, ocContractAddr)
	if err := k.PinCode(ctx, ocCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg trusted circle contract")
	}
	logger.Info("oversight community contract", "address", ocContractAddr, "code_id", ocCodeID)

	// setup stake contract
	stakeCodeID, err := k.Create(ctx, systemAdminAddr, tg4Stake, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 stake contract")
	}
	tg4StakeInitMsg := newStakeInitMsg(gs, systemAdminAddr)
	stakeContractAddr, _, err := k.Instantiate(ctx, stakeCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(tg4StakeInitMsg), "stakers", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 stake")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeStaking, stakeContractAddr)
	if err := tk.SetPrivileged(ctx, stakeContractAddr); err != nil {
		return sdkerrors.Wrap(err, "grant privileges to stake contract")
	}
	logger.Info("stake contract", "address", stakeContractAddr, "code_id", stakeCodeID)

	// setup mixer contract
	tg4MixerInitMsg := contract.TG4MixerInitMsg{
		LeftGroup:        engagementContractAddr.String(),
		RightGroup:       stakeContractAddr.String(),
		PreAuthsSlashing: 1,
		// TODO: allow to configure the other types.
		// We need to analyze benchmarks and discuss first.
		// This maintains same behavior
		FunctionType: contract.MixerFunction{
			GeometricMean: &struct{}{},
		},
	}
	mixerCodeID, err := k.Create(ctx, systemAdminAddr, tg4Mixer, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 mixer contract")
	}
	mixerContractAddr, _, err := k.Instantiate(ctx, mixerCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(tg4MixerInitMsg), "poe", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 mixer")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeMixer, mixerContractAddr)
	if err := k.PinCode(ctx, mixerCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 mixer contract")
	}
	logger.Info("mixer contract", "address", mixerContractAddr, "code_id", mixerCodeID)

	// setup valset contract
	valSetCodeID, err := k.Create(ctx, systemAdminAddr, tgValset, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store valset contract")
	}

	valsetInitMsg := newValsetInitMsg(gs, systemAdminAddr, mixerContractAddr, engagementContractAddr, engagementCodeID)
	valsetJSON := mustMarshalJson(valsetInitMsg)
	valsetContractAddr, _, err := k.Instantiate(ctx, valSetCodeID, systemAdminAddr, systemAdminAddr, valsetJSON, "valset", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate valset")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeValset, valsetContractAddr)

	// setup distribution contract address
	valsetCfg, err := poeKeeper.ValsetContract(ctx).QueryConfig(ctx)
	if err != nil {
		return sdkerrors.Wrap(err, "query valset config")
	}

	distrAddr, err := sdk.AccAddressFromBech32(valsetCfg.RewardsContract)
	if err != nil {
		return sdkerrors.Wrap(err, "distribution contract address")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeDistribution, distrAddr)

	if err := tk.SetPrivileged(ctx, valsetContractAddr); err != nil {
		return sdkerrors.Wrap(err, "grant privileges to valset contract")
	}
	logger.Info("valset contract", "address", valsetContractAddr, "code_id", valSetCodeID)

	// setup oversight community gov proposals contract
	ocGovCodeID, err := k.Create(ctx, systemAdminAddr, tgOCGovProposalsCircles, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg oc gov proposals contract: ")
	}
	ocGovInitMsg := newOCGovProposalsInitMsg(gs, ocContractAddr, engagementContractAddr, valsetContractAddr)
	ocGovProposalsContractAddr, _, err := k.Instantiate(ctx, ocGovCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(ocGovInitMsg), "oversight_committee gov proposals", deposit)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg oc gov proposals contract")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeOversightCommunityGovProposals, ocGovProposalsContractAddr)
	if err := k.PinCode(ctx, ocGovCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg oc gov proposals contract")
	}
	logger.Info("oversight community gov proposal contract", "address", ocGovProposalsContractAddr, "code_id", ocGovCodeID)

	err = poeKeeper.EngagementContract(ctx).UpdateAdmin(ctx, ocGovProposalsContractAddr, systemAdminAddr)
	if err != nil {
		return sdkerrors.Wrap(err, "set new engagement contract admin")
	}

	err = poeKeeper.ValsetContract(ctx).UpdateAdmin(ctx, ocGovProposalsContractAddr, systemAdminAddr)
	if err != nil {
		return sdkerrors.Wrap(err, "set new valset contract admin")
	}

	// setup community pool
	//
	communityPoolCodeID, err := k.Create(ctx, systemAdminAddr, tgCommunityPool, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store community pool contract")
	}
	communityPoolInitMsg := contract.CommunityPoolInitMsg{
		VotingRules:  toContractVotingRules(gs.CommunityPoolContractConfig.VotingRules),
		GroupAddress: engagementContractAddr.String(),
		DELME:        "to-be-removed",
	}
	communityPoolContractAddr, _, err := k.Instantiate(ctx, communityPoolCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(communityPoolInitMsg), "stakers", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate community pool")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeCommunityPool, communityPoolContractAddr)
	if err := k.PinCode(ctx, communityPoolCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin community pool contract")
	}
	logger.Info("community pool contract", "address", communityPoolContractAddr, "code_id", communityPoolCodeID)

	// setup validator voting contract
	//
	validatorVotingCodeID, err := k.Create(ctx, systemAdminAddr, tgValidatorVoting, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store validator voting contract")
	}
	validatorVotingInitMsg := contract.ValidatorVotingInitMsg{
		VotingRules:  toContractVotingRules(gs.ValidatorVotingContractConfig.VotingRules),
		GroupAddress: distrAddr.String(),
		DELME:        "to-be-removed",
	}
	validatorVotingContractAddr, _, err := k.Instantiate(ctx, validatorVotingCodeID, systemAdminAddr, systemAdminAddr, mustMarshalJson(validatorVotingInitMsg), "stakers", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate validator voting")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeValidatorVoting, validatorVotingContractAddr)

	// todo: contract requires GovProposalExecutor ConsensusParamChanger permissions which are not implemented
	// if err := tk.SetPrivileged(ctx, validatorVotingContractAddr); err != nil {
	//	return sdkerrors.Wrap(err, "grant privileges to validator voting contract")
	// }
	if err := k.PinCode(ctx, validatorVotingCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin validator voting contract")
	}

	logger.Info("validator voting contract", "address", validatorVotingContractAddr, "code_id", validatorVotingCodeID)

	if err := setAllPoEContractsInstanceMigrators(ctx, k, poeKeeper, systemAdminAddr, validatorVotingContractAddr); err != nil {
		return sdkerrors.Wrap(err, "set new instance admin")
	}
	return nil
}

// set new migrator for all PoE contracts
func setAllPoEContractsInstanceMigrators(ctx sdk.Context, k wasmtypes.ContractOpsKeeper, poeKeeper poeKeeper, oldAdminAddr, newAdminAddr sdk.AccAddress) error {
	// set new admin for all contracts
	for name, v := range types.PoEContractType_value {
		contractType := types.PoEContractType(v)
		if contractType == types.PoEContractTypeUndefined {
			continue
		}
		if contractType == types.PoEContractTypeDistribution {
			continue // disabled due to https://github.com/confio/tgrade-contracts/issues/353
		}
		addr, err := poeKeeper.GetPoEContractAddress(ctx, contractType)
		if err != nil {
			return sdkerrors.Wrapf(err, "failed to find contract address for %s", name)
		}
		if err := k.UpdateContractAdmin(ctx, addr, oldAdminAddr, newAdminAddr); err != nil {
			return sdkerrors.Wrapf(err, "%s contract", name)
		}
	}
	return nil
}

// build instantiate message for the trusted circle contract that contains the oversight committee
func newOCInitMsg(gs types.GenesisState) contract.TrustedCircleInitMsg {
	cfg := gs.OversightCommitteeContractConfig
	return contract.TrustedCircleInitMsg{
		Name:                      cfg.Name,
		EscrowAmount:              cfg.EscrowAmount.Amount,
		VotingPeriod:              cfg.VotingRules.VotingPeriod,
		Quorum:                    *contract.DecimalFromPercentage(cfg.VotingRules.Quorum),
		Threshold:                 *contract.DecimalFromPercentage(cfg.VotingRules.Threshold),
		AllowEndEarly:             cfg.VotingRules.AllowEndEarly,
		InitialMembers:            []string{}, // sender is added to OC by default in the contract
		DenyList:                  cfg.DenyListContractAddress,
		EditTrustedCircleDisabled: true, // product requirement for OC
	}
}

// build instantiate message for OC Proposals contract
func newOCGovProposalsInitMsg(gs types.GenesisState, ocContract, engagementContract, valsetContract sdk.AccAddress) contract.OCProposalsInitMsg {
	cfg := gs.OversightCommitteeContractConfig
	return contract.OCProposalsInitMsg{
		GroupContractAddress:      ocContract.String(),
		ValsetContractAddress:     valsetContract.String(),
		EngagementContractAddress: engagementContract.String(),
		VotingRules:               toContractVotingRules(cfg.VotingRules),
	}
}

func newEngagementInitMsg(gs types.GenesisState, adminAddr sdk.AccAddress) contract.TG4EngagementInitMsg {
	tg4EngagementInitMsg := contract.TG4EngagementInitMsg{
		Admin:            adminAddr.String(),
		Members:          make([]contract.TG4Member, len(gs.Engagement)),
		PreAuthsHooks:    1,
		PreAuthsSlashing: 1,
		Token:            gs.BondDenom,
		Halflife:         uint64(gs.EngagmentContractConfig.Halflife.Seconds()),
	}
	for i, v := range gs.Engagement {
		tg4EngagementInitMsg.Members[i] = contract.TG4Member{
			Addr:   v.Address,
			Weight: v.Weight,
		}
	}
	return tg4EngagementInitMsg
}

func newStakeInitMsg(gs types.GenesisState, adminAddr sdk.AccAddress) contract.TG4StakeInitMsg {
	var claimLimit = uint64(gs.StakeContractConfig.ClaimAutoreturnLimit)
	return contract.TG4StakeInitMsg{
		Admin:            adminAddr.String(),
		Denom:            gs.BondDenom,
		MinBond:          gs.StakeContractConfig.MinBond,
		TokensPerWeight:  gs.StakeContractConfig.TokensPerWeight,
		UnbondingPeriod:  uint64(gs.StakeContractConfig.UnbondingPeriod.Seconds()),
		AutoReturnLimit:  &claimLimit,
		PreAuthsHooks:    1,
		PreAuthsSlashing: 1,
	}
}

func newValsetInitMsg(
	gs types.GenesisState,
	admin sdk.AccAddress,
	mixerContractAddr sdk.AccAddress,
	engagementAddr sdk.AccAddress,
	engagementCodeID uint64,
) contract.ValsetInitMsg {
	config := gs.ValsetContractConfig
	return contract.ValsetInitMsg{
		Admin:                 admin.String(),
		Membership:            mixerContractAddr.String(),
		MinWeight:             config.MinWeight,
		MaxValidators:         config.MaxValidators,
		EpochLength:           uint64(config.EpochLength.Seconds()),
		EpochReward:           config.EpochReward,
		InitialKeys:           []contract.Validator{},
		Scaling:               config.Scaling,
		FeePercentage:         contract.DecimalFromPercentage(config.FeePercentage),
		AutoUnjail:            config.AutoUnjail,
		ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(int64(config.ValidatorsRewardRatio))),
		DistributionContract:  engagementAddr.String(),
		RewardsCodeID:         engagementCodeID,
	}
}

// verifyPoEContracts verifies all PoE contracts are setup as expected
func verifyPoEContracts(ctx sdk.Context, k wasmtypes.ContractOpsKeeper, tk twasmKeeper, poeKeeper poeKeeper, gs types.GenesisState) error {
	return errors.New("not supported, yet")
	// all poe contracts pinned
	// valset privileged
	// valset has registered for endblock valset update privilege
	// admin set matches genesis system admin address for engagement and staking contract
}

// mustMarshalJson with stdlib json
func mustMarshalJson(s interface{}) []byte {
	jsonBz, err := json.Marshal(s)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal json: %s", err))
	}
	return jsonBz
}

// map to contract object
func toContractVotingRules(votingRules types.VotingRules) contract.VotingRules {
	return contract.VotingRules{
		VotingPeriod:  votingRules.VotingPeriod,
		Quorum:        *contract.DecimalFromPercentage(votingRules.Quorum),
		Threshold:     *contract.DecimalFromPercentage(votingRules.Threshold),
		AllowEndEarly: votingRules.AllowEndEarly,
	}
}
