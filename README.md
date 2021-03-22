# Tgrade


## Development


### Local testnet
* Install
```sh
make install
```
* Setup network
```sh
# --v 1 = single node
tgrade testnet --chain-id=testing --output-dir=$(pwd)/testnet --v=1 --keyring-backend=test --commit-timeout=1500ms --minimum-gas-prices=""
```
* Start a validator node
```sh
tgrade start --home=./testnet/node0/tgrade
```

* Use test keyring
```sh
tgrade keys list --keyring-backend=test --keyring-dir=testnet/node0/tgrade
```
