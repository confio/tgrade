package poe

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	//go:embed contract/version.txt
	contractVersion []byte
)

// ClearEmbeddedContracts release memory
func ClearEmbeddedContracts() {
	tg4Engagement = nil
	tg4Stake = nil
	tg4Mixer = nil
	tgValset = nil
}

type poeKeeper interface {
	keeper.ContractSource
	SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
}

// bootstrapPoEContracts stores and instantiates all PoE contracts:
//
// * [tg4-group](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-group) - engagement group with weighted
//  members
//* [tg4-stake](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-stake) - validator group weighted by
//  staked amount
//* [valset](https://github.com/confio/tgrade-contracts/tree/main/contracts/tgrade-valset) - privileged contract to map a
//  trusted cw4 contract to the Tendermint validator set running the chain
//* [mixer](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-mixer) - calculates the combined value of
//  stake and engagement points. Source for the valset contract.
func bootstrapPoEContracts(ctx sdk.Context, k wasmtypes.ContractOpsKeeper, tk twasmKeeper, poeKeeper poeKeeper, gs types.GenesisState) error {
	tg4EngagementInitMsg := contract.TG4EngagementInitMsg{
		Admin:    gs.SystemAdminAddress,
		Members:  make([]contract.TG4Member, len(gs.Engagement)),
		Preauths: 1,
		Token:    gs.BondDenom,
		// TODO: allow us to configure halflife in Genesis
		// now hardcoded as 180 days = 180 * 86400s
		Halflife: 15552000,
	}
	for i, v := range gs.Engagement {
		tg4EngagementInitMsg.Members[i] = contract.TG4Member{
			Addr:   v.Address,
			Weight: v.Weight,
		}
	}
	systemAdmin, err := sdk.AccAddressFromBech32(gs.SystemAdminAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "system admin")
	}
	creator := systemAdmin
	engagementCodeID, err := k.Create(ctx, creator, tg4Engagement, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 engagement contract")
	}
	engagementContractAddr, _, err := k.Instantiate(ctx, engagementCodeID, creator, systemAdmin, mustMarshalJson(tg4EngagementInitMsg), "engagement", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 engagement")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeEngagement, engagementContractAddr)
	if err := k.PinCode(ctx, engagementCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 engagement contract")
	}

	var claimLimit = uint64(gs.StakeContractConfig.ClaimAutoreturnLimit)
	stakeCodeID, err := k.Create(ctx, creator, tg4Stake, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 stake contract")
	}
	tg4StakerInitMsg := newStakeInitMsg(gs, claimLimit)
	stakersContractAddr, _, err := k.Instantiate(ctx, stakeCodeID, creator, systemAdmin, mustMarshalJson(tg4StakerInitMsg), "stakers", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 stake")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeStaking, stakersContractAddr)
	if err := tk.SetPrivileged(ctx, stakersContractAddr); err != nil {
		return sdkerrors.Wrap(err, "grant privileges to staker contract")
	}

	tg4MixerInitMsg := contract.TG4MixerInitMsg{
		LeftGroup:  engagementContractAddr.String(),
		RightGroup: stakersContractAddr.String(),
		// TODO: allow to configure the other types.
		// We need to analyze benchmarks and discuss first.
		// This maintains same behavior
		FunctionType: contract.MixerFunction{
			GeometricMean: &struct{}{},
		},
	}
	mixerCodeID, err := k.Create(ctx, creator, tg4Mixer, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 mixer contract")
	}
	mixerContractAddr, _, err := k.Instantiate(ctx, mixerCodeID, creator, systemAdmin, mustMarshalJson(tg4MixerInitMsg), "poe", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 mixer")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeMixer, mixerContractAddr)

	if err := k.PinCode(ctx, mixerCodeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 mixer contract")
	}

	valSetCodeID, err := k.Create(ctx, creator, tgValset, &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store valset contract")
	}

	valsetInitMsg := newValsetInitMsg(mixerContractAddr, gs, engagementContractAddr, engagementCodeID)
	valsetJson := mustMarshalJson(valsetInitMsg)
	valsetContractAddr, _, err := k.Instantiate(ctx, valSetCodeID, creator, systemAdmin, valsetJson, "valset", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate valset")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeValset, valsetContractAddr)

	if err := tk.SetPrivileged(ctx, valsetContractAddr); err != nil {
		return sdkerrors.Wrap(err, "grant privileges to valset contract")
	}
	return nil
}

func newStakeInitMsg(gs types.GenesisState, claimLimit uint64) contract.TG4StakeInitMsg {
	return contract.TG4StakeInitMsg{
		Admin:           gs.SystemAdminAddress,
		Denom:           gs.BondDenom,
		MinBond:         gs.StakeContractConfig.MinBond,
		TokensPerWeight: gs.StakeContractConfig.TokensPerWeight,
		UnbondingPeriod: uint64(gs.StakeContractConfig.UnbondingPeriod.Seconds()),
		AutoReturnLimit: &claimLimit,
		Preauths:        uint64(gs.StakeContractConfig.PreAuths),
	}
}

// TODO: needs tg4-engagement code id, address
func newValsetInitMsg(mixerContractAddr sdk.AccAddress, gs types.GenesisState, engagementAddr sdk.AccAddress, engagementID uint64) contract.ValsetInitMsg {
	return contract.ValsetInitMsg{
		Membership:    mixerContractAddr.String(),
		MinWeight:     gs.ValsetContractConfig.MinWeight,
		MaxValidators: gs.ValsetContractConfig.MaxValidators,
		EpochLength:   uint64(gs.ValsetContractConfig.EpochLength.Seconds()),
		EpochReward:   gs.ValsetContractConfig.EpochReward,
		InitialKeys:   []contract.Validator{},
		Scaling:       gs.ValsetContractConfig.Scaling,
		FeePercentage: contract.DecimalFromPercentage(gs.ValsetContractConfig.FeePercentage),
		// TODO: set AutoJail from genesis
		// TODO: set ValidatorsRewardRatio from genesis (hardcode to 50% here)
		ValidatorsRewardRatio: contract.DecimalFromPercentage(sdk.NewDec(50)),
		DistributionContract:  engagementAddr.String(),
		RewardsCodeId:         engagementID,
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
