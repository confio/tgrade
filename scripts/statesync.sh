#!/bin/bash

# Print every command.
set -ux

# Initialize chain.
tgrade init test --chain-id tgrade-mainnet-1

# Get "trust_hash" and "trust_height".
INTERVAL=2000
SNAP_RPC="https://rpc.mainnet-1.tgrade.confio.run:443"
LATEST_HEIGHT=$(curl -s $SNAP_RPC/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL))
TRUST_HASH=$(curl -s "$SNAP_RPC/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"

# Export state sync variables.
export TGRADE_STATESYNC_ENABLE=true
export TGRADE_STATESYNC_RPC_SERVERS="$SNAP_RPC,https://tgrade-mainnet-rpc.allthatnode.com:26657"
export TGRADE_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export TGRADE_STATESYNC_TRUST_HASH=$TRUST_HASH

# Start chain.
tgrade start