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
	"time"
)

var (
	//go:embed contract/tg4_group.wasm
	tg4Group []byte
	//go:embed contract/tg4_stake.wasm
	tg4Stake []byte
	//go:embed contract/tg4_mixer.wasm
	tg4Mixer []byte
	//go:embed contract/tgrade_valset.wasm
	tgValset []byte
	//go:embed contract/version.txt
	contractVersion []byte
)

// clearEmbeddedContracts release memory
func clearEmbeddedContracts() {
	tg4Group = nil
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
	tg4EngagementInitMsg := contract.TG4GroupInitMsg{
		Admin:    gs.SystemAdminAddress,
		Members:  make([]contract.TG4Member, len(gs.Engagement)),
		Preauths: 1,
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
	codeID, err := k.Create(ctx, creator, tg4Group, "https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-group", "cosmwasm/workspace-optimizer:0.11.3", &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 group contract")
	}
	engagementContractAddr, _, err := k.Instantiate(ctx, codeID, creator, systemAdmin, mustMarshalJson(tg4EngagementInitMsg), "engagement", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 group")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeEngagement, engagementContractAddr)
	if err := k.PinCode(ctx, codeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 group contract")
	}

	tg4StakerInitMsg := contract.TG4StakeInitMsg{
		Admin:           gs.SystemAdminAddress,
		Denom:           contract.Denom{Native: gs.BondDenom},
		MinBond:         "1",
		TokensPerWeight: "1",
		UnbondingPeriod: contract.UnbodingPeriod{
			TimeInSec: uint64(21 * 24 * time.Hour.Seconds()),
		},
		Preauths: 1,
	}
	codeID, err = k.Create(ctx, creator, tg4Stake, "https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-stake", "cosmwasm/workspace-optimizer:0.11.3", &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 stake contract")
	}
	stakersContractAddr, _, err := k.Instantiate(ctx, codeID, creator, systemAdmin, mustMarshalJson(tg4StakerInitMsg), "stakers", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 stake")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeStaking, stakersContractAddr)
	if err := k.PinCode(ctx, codeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 stake contract")
	}

	tg4MixerInitMsg := contract.TG4MixerInitMsg{
		LeftGroup:  engagementContractAddr.String(),
		RightGroup: stakersContractAddr.String(),
	}
	codeID, err = k.Create(ctx, creator, tg4Mixer, "https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-mixer", "cosmwasm/workspace-optimizer:0.11.3", &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store tg4 mixer contract")
	}
	mixerContractAddr, _, err := k.Instantiate(ctx, codeID, creator, systemAdmin, mustMarshalJson(tg4MixerInitMsg), "poe", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate tg4 mixer")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeMixer, mixerContractAddr)

	if err := k.PinCode(ctx, codeID); err != nil {
		return sdkerrors.Wrap(err, "pin tg4 mixer contract")
	}

	valsetInitMsg := contract.ValsetInitMsg{
		Membership:    mixerContractAddr.String(),
		MinWeight:     1,
		MaxValidators: 100,
		EpochLength:   1,
		EpochReward:   sdk.NewCoin(gs.BondDenom, sdk.OneInt()),
		InitialKeys:   []contract.ValsetInitKey{},
	}
	codeID, err = k.Create(ctx, creator, tgValset, "https://github.com/confio/tgrade-contracts/tree/main/contracts/tgrade-valset", "cosmwasm/workspace-optimizer:0.11.3", &wasmtypes.AllowEverybody)
	if err != nil {
		return sdkerrors.Wrap(err, "store valset contract")
	}
	valsetContractAddr, _, err := k.Instantiate(ctx, codeID, creator, systemAdmin, mustMarshalJson(valsetInitMsg), "valset", nil)
	if err != nil {
		return sdkerrors.Wrap(err, "instantiate valset")
	}
	poeKeeper.SetPoEContractAddress(ctx, types.PoEContractTypeValset, valsetContractAddr)

	if err := tk.SetPrivileged(ctx, valsetContractAddr); err != nil {
		return sdkerrors.Wrap(err, "grant privileges to valset contract")
	}
	return nil
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
