package v3

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/confio/tgrade/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the Tgrade v3 upgrade.
const UpgradeName = "v3"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
