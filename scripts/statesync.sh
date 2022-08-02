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
tgrade start --p2p.seeds="0c3b7d5a4253216de01b8642261d4e1e76aee9d8@45.76.202.195:26656,7d08b16e568d8fcee1a6a4850197054aa469bf71@seed.tgrade.stakewith.us:54456" --p2p.persistent_peers="0a63421f67d02e7fb823ea6d6ceb8acf758df24d@142.132.226.137:26656,4a319eead699418e974e8eed47c2de6332c3f825@167.235.255.9:26656,6918efd409684d64694cac485dbcc27dfeea4f38@49.12.240.203:26656"