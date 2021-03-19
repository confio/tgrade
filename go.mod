module github.com/confio/tgrade

go 1.16

require (
	github.com/CosmWasm/wasmd v0.16.0-alpha1.0.20210319085201-d9142662c19a
	github.com/cosmos/cosmos-sdk v0.42.1
	github.com/gorilla/mux v1.8.0
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.20.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.8
	github.com/tendermint/tm-db v0.6.4
)

// https://docs.cosmos.network/v0.41/core/grpc_rest.html#grpc-server
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
