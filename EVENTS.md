# Event System

Events are an essential part of the Cosmos SDK. They are similar to "logs" in Ethereum and allow a blockchain
app to attach key-value pairs to a transaction that can later be used to search for it or extract some information
in human readable form. Events are not written to the application state, nor do they form part of the AppHash,
but mainly intended for client use (and become an essential API for any reactive app or app that searches for txs).

See https://github.com/CosmWasm/wasmd/blob/master/EVENTS.md for more details


### Standard Events in x/poe
```go
// create new validator
sdk.NewEvent(
    "create_validator",
    sdk.NewAttribute("operator", msg.DelegatorAddress),
    sdk.NewAttribute("moniker", msg.Description.Moniker),
    sdk.NewAttribute("pubkey", hex.EncodeToString(pk.Bytes())),
    sdk.NewAttribute("amount", msg.Value.Amount.String()),
)

// update any validator data
sdk.NewEvent(
    "update_validator",
    sdk.NewAttribute("operator", msg.DelegatorAddress),
    sdk.NewAttribute("moniker", msg.Description.Moniker),
),

```

### Standard Events in x/twasm
In twasm we have the concept of (privileged](https://github.com/confio/tgrade/tree/main/x/twasm#privileged) contracts that
add a number of new events to the system:

```go
// when a contract was set as privileged
sdk.NewEvent(
    "set_privileged_contract",
    sdk.NewAttribute("_contract_address", contractAddr.String()),
)

// when a contract had the privileged flag removed
sdk.NewEvent(
    "unset_privileged_contract",
    sdk.NewAttribute("_contract_address", contractAddr.String()),
)

// when a new privilege was given to a contract
sdk.NewEvent(
    "register_privilege",
    sdk.NewAttribute("_contract_address", contractAddr.String()),
    sdk.NewAttribute("privilege_type", privilegeType.String()),
)

// when a privilege was removed for a contract
event := sdk.NewEvent(
    "release_privilege",
    sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
    sdk.NewAttribute("privilege_type", privilegeType.String()),
)
```
We also emit the standard events from [wasmd/x/wasm](https://github.com/CosmWasm/wasmd/blob/master/EVENTS.md#standard-events-in-xwasm)