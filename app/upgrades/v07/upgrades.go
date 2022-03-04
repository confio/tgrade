package v07

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	poekeeper "github.com/confio/tgrade/x/poe/keeper"
	poetypes "github.com/confio/tgrade/x/poe/types"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
)

func CreateUpgradeHandler(
	wasmKeeper twasmkeeper.Keeper,
	poeKeeper poekeeper.ContractSource,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Migrating to set tfi factory migrator address")
		govVotingContractAddr, err := poeKeeper.GetPoEContractAddress(ctx, poetypes.PoEContractTypeValidatorVoting)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "validator voting contract address")
		}
		tfiFactoryAddr, err := sdk.AccAddressFromBech32(plan.Info) // using the info field to transport the address
		if err != nil {
			return nil, sdkerrors.Wrap(err, "tfi contract address")
		}

		govPermission := wasmkeeper.NewGovPermissionKeeper(&wasmKeeper.Keeper)
		err = govPermission.UpdateContractAdmin(ctx, tfiFactoryAddr, govVotingContractAddr, govVotingContractAddr)
		return vm, err
	}
}
