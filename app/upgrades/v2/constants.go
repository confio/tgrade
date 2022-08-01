package v2

import (
	"github.com/confio/tgrade/app/upgrades"
)

// UpgradeName defines the on-chain upgrade name for the Tgrade v2 upgrade.
const UpgradeName = "v2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
