module github.com/confio/tgrade

go 1.15

require (
	github.com/CosmWasm/wasmd v0.16.0-alpha1.0.20210416113915-aabc575ff493
	github.com/CosmWasm/wasmvm v0.14.0-beta3
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/kr/text v0.2.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.21.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.9
	github.com/tendermint/tm-db v0.6.4
	github.com/tidwall/gjson v1.7.3
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v2 v2.4.0
)

// https://docs.cosmos.network/v0.41/core/grpc_rest.html#grpc-server
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
