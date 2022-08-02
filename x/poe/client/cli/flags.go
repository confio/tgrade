package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagPubKey          = "pubkey"
	FlagAmount          = "amount"
	FlagVestingAmount   = "vesting-amount"
	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"
	FlagNodeID          = "node-id"
	FlagIP              = "ip"
	flagAddress         = "address"
	flagEngagement      = "engagement"
	flagDistribution    = "distribution"
)

// FlagSetAmounts Returns the FlagSet for amount related operations.
func FlagSetAmounts() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagAmount, "", "Amount of liquid coins to bond")
	fs.String(FlagVestingAmount, "", "Amount of vesting coins to bond")
	return fs
}

// FlagSetPublicKey Returns the flagset for Public Key related operations.
func FlagSetPublicKey() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagPubKey, "", "The validator's Protobuf JSON encoded public key")
	return fs
}

func flagSetDescriptionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")

	return fs
}
