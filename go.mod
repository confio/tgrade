module github.com/confio/tgrade

go 1.15

require (
	github.com/CosmWasm/wasmd v0.16.0-alpha1.0.20210319085201-d9142662c19a
	github.com/CosmWasm/wasmvm v0.14.0-beta1
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/cosmos/cosmos-sdk v0.42.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.20.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.8
	github.com/tendermint/tm-db v0.6.4
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
)

// https://docs.cosmos.network/v0.41/core/grpc_rest.html#grpc-server
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/CosmWasm/wasmd => ../../cosmwasm/wasmd