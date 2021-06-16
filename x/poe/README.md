# Proof of Engagement (PoE)

This module contains the Proof of Engagement (PoE) contracts and integration points. See
the [Whitepaper](https://github.com/confio/ProofOfEngagement) for more details about PoE.

This module provides first class support for PoE:

* Bootstrap and contract instantiation
* Query and CLI support
* Genesis import/ export
* Integration tests

### Contracts

* [tg4-group](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-group) - engagement group with weighted
  members
* [tg4-stake](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-stake) - validator group weighted by
  staked amount
* [valset](https://github.com/confio/tgrade-contracts/tree/main/contracts/tgrade-valset) - privileged contract to map a
  trusted cw4 contract to the Tendermint validator set running the chain
* [mixer](https://github.com/confio/tgrade-contracts/tree/main/contracts/tg4-mixer) - calculates the combined value of
  stake and engagement points. Source for the valset contract.

### Command line interface (CLI)

* Commands

```sh
  tgrade tx poe -h
```

* Query

```sh
  tgrade query poe -h
```

### Disclaimer

This module uses code that was part on
the [Cosmos-sdk genutil](https://github.com/cosmos/cosmos-sdk/tree/v0.42.5/x/genutil) module.

Credits and big thank you go to the original authors