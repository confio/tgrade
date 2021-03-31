FROM golang:1.16-alpine3.12 AS go-builder

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v0.14.0-beta1/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep b69cf9ffbdfb2f1bd1e6f730ecee1eb0d06a1473840a24709aec7a7df5907d45

WORKDIR /code
# Speed up build by caching Go dependencies as a separate step
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . /code/

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build

# --------------------------------------------------------
FROM alpine:3.12

COPY --from=go-builder /code/build/tgrade /usr/bin/tgrade

WORKDIR /opt

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD ["/usr/bin/tgrade", "version"]