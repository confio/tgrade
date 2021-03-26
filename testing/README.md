# Testing

Test framework for system tests. Uses:

* Docker-compose
* gjson
## Execute a single test
```sh
go test -count=1 -v ./testing --run TestSmokeTest
```
* Force a rebuild
```sh
go test -count=1 -v ./testing --run TestSmokeTest -rebuild
```

## Disclaimer
This was inspired by the amazing work of the [e-money](https://github.com/e-money) team. Thank you!