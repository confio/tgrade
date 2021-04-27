# Testing

Test framework for system tests. Uses:

* Docker
* Docker-compose
* gjson

Server and client side are executed in Docker containers

## Execute a single test

```sh
go test -tags system_test -count=1 -v ./testing --run TestSmokeTest  --verbose
```

* Force a binary rebuild before running the test

```sh
go test -tags system_test -count=1 -v ./testing --run TestSmokeTest  --verbose -rebuild
```

## Resources

* [gjson query syntax](https://github.com/tidwall/gjson#path-syntax)

## Disclaimer

This was inspired by the amazing work of the [e-money](https://github.com/e-money) team on their system tests. Thank
you!