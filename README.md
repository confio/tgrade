# Tgrade


## Development

### Local testnet
* Install
```sh
make install
```
* Setup network
```sh
tgrade testnet --chain-id=testing --output-dir=$(pwd)/testnet --v=2 --keyring-backend=test --commit-timeout=1500ms --minimum-gas-prices=""
```
* Start a validator node
```sh
tgrade start --home=./testnet/node0/tgrade
```

## License

Copyright (c) 2021 Confio GmbH. All Rights Reserved.

This repository contains confidential and proprietary information of Confio GmbH,
and is protected under U.S. and international copyright and other intellectual property laws.

It is not offered under an open source license.

Please contact hello (at) confio (dot) gmbh for licensing related questions.
