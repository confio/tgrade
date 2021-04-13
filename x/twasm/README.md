# twasm

Extended version of [CosmWasm/wasmd/x/wasm](https://github.com/CosmWasm/wasmd/tree/d9142662c19a151f34ff4a66d69007124051bd28/x/wasm) module.



## Concepts
Tgrade is different than wasmd is.

### Privileged
How nice would it be to build chain logic in wasm with all the flexibility that [wasmd](https://github.com/CosmWasm/wasmd) brings you today.
Now imagine that you build the chain and can do what you want. This inspired us to create a new set of contracts that we call
"privileged". They can interact with the native go modules or receive callbacks from the chain to do certain business logic.
This power comes with some responsibility that we restrict to "privileged" contracts only.
Technically it is a marker persisted as a secondary index that points to the contract and a set of 
[predefined callbacks](./types/callbacks.go) the contract can register.



