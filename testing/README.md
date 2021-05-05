# Testing

Test framework for system tests. Uses:

* gjson

Server and client side are executed on the host machine

## Execute a single test

```sh
go test -tags system_test -count=1 -v ./testing --run TestSmokeTest  -verbose
```

* Force a binary rebuild before running the test

```sh
go test -tags system_test -count=1 -v ./testing --run TestSmokeTest  -verbose -rebuild
```

Test cli parameters

* `-verbose` verbose output
* `-rebuild` - rebuild artifacts
* `-wait-time` duration - time to wait for chain events (default 30s)
* `-nodes-count` int - number of nodes in the cluster (default 4)

# Port ranges
With *n* nodes:
* `26657` - `26657+n` - RPC
* `1317` - `1317+n` - API
* `9090` - `9090+n` - GRPC
* `16656` - `16656+n` - P2P

For example Node *3* listens on `26660` for RPC calls

## Resources

* [gjson query syntax](https://github.com/tidwall/gjson#path-syntax)

## Disclaimer

This was inspired by the amazing work of the [e-money](https://github.com/e-money) team on their system tests. Thank
you!