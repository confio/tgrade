#!/bin/bash
set -o errexit -o nounset -o pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
KEY=node0
echo "-----------------------"
echo "## Add new CosmWasm contract"
RESP=$(tgrade tx wasm store "$DIR/hackatom.wasm.gzip" \
  --from $KEY --gas 1500000 -y --chain-id=testing --node=http://localhost:26657 -b block \
  --keyring-backend=test --keyring-dir=testnet/node0/tgrade)

CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[0].attributes[-1].value')
echo "* Code id: $CODE_ID"
echo "* Download code"
TMPDIR=$(mktemp -t tgradeXXXXXX)
tgrade q wasm code "$CODE_ID" "$TMPDIR"
rm -f "$TMPDIR"
echo "-----------------------"
echo "## List code"
tgrade query wasm list-code --node=http://localhost:26657 --chain-id=testing -o json | jq

echo "-----------------------"
echo "## Create new contract instance"
INIT="{\"verifier\":\"$(tgrade keys show $KEY -a --keyring-backend=test --keyring-dir=testnet/node0/tgrade)\", \"beneficiary\":\"$(tgrade keys show $KEY -a --keyring-backend=test --keyring-dir=testnet/node0/tgrade)\"}"
tgrade tx wasm instantiate "$CODE_ID" "$INIT" --admin=$(tgrade keys show $KEY -a --keyring-backend=test --keyring-dir=testnet/node0/tgrade) \
  --from $KEY --amount="100utgd" --label "local0.1.0" \
  --gas 1000000 -y --chain-id=testing -b block \
  --keyring-backend=test --keyring-dir=testnet/node0/tgrade | jq

CONTRACT=$(tgrade query wasm list-contract-by-code "$CODE_ID" -o json | jq -r '.contract_infos[-1].address')
echo "* Contract address: $CONTRACT"
echo "### Query all"
RESP=$(tgrade query wasm contract-state all "$CONTRACT" -o json)
echo "$RESP" | jq
echo "### Query smart"
tgrade query wasm contract-state smart "$CONTRACT" '{"verifier":{}}' -o json | jq
echo "### Query raw"
QUERY_KEY=$(echo "$RESP" | jq -r ".models[0].key")
tgrade query wasm contract-state raw "$CONTRACT" "$QUERY_KEY" -o json | jq

echo "-----------------------"
echo "## Execute contract $CONTRACT"
MSG='{"release":{}}'
tgrade tx wasm execute "$CONTRACT" "$MSG" \
  --from $KEY \
  --gas 1000000 -y --chain-id=testing -b block \
  --keyring-backend=test --keyring-dir=testnet/node0/tgrade | jq

echo "-----------------------"
echo "## Set new admin"
echo "### Query old admin: $(tgrade q wasm contract $CONTRACT -o json | jq -r '.contract_info.admin')"
echo "### Update contract"
tgrade tx wasm set-contract-admin "$CONTRACT" $(tgrade keys show $KEY -a --keyring-backend=test --keyring-dir=testnet/node0/tgrade ) \
  --from $KEY -y --chain-id=testing -b block --keyring-backend=test --keyring-dir=testnet/node0/tgrade  | jq
echo "### Query new admin: $(tgrade q wasm contract $CONTRACT -o json | jq -r '.admin')"

exit 0

echo "-----------------------"
echo "## Migrate contract"
echo "### Upload new code"
RESP=$(tgrade tx wasm store "$DIR/../../x/wasm/keeper/testdata/burner.wasm" \
  --from $KEY --gas 1000000 -y --chain-id=testing --node=http://localhost:26657 -b block)

BURNER_CODE_ID=$(echo "$RESP" | jq -r '.logs[0].events[0].attributes[-1].value')
echo "### Migrate to code id: $BURNER_CODE_ID"

DEST_ACCOUNT=$(tgrade keys show fred -a)
tgrade tx wasm migrate "$CONTRACT" "$BURNER_CODE_ID" "{\"payout\": \"$DEST_ACCOUNT\"}" --from fred \
  --chain-id=testing -b block -y | jq

echo "### Query destination account: $BURNER_CODE_ID"
tgrade q bank balances "$DEST_ACCOUNT" -o json | jq
echo "### Query contract meta data: $CONTRACT"
tgrade q wasm contract "$CONTRACT" -o json | jq

echo "### Query contract meta history: $CONTRACT"
tgrade q wasm contract-history "$CONTRACT" -o json | jq

echo "-----------------------"
echo "## Clear contract admin"
echo "### Query old admin: $(tgrade q wasm contract $CONTRACT -o json | jq -r '.admin')"
echo "### Update contract"
tgrade tx wasm clear-contract-admin "$CONTRACT" \
  --from fred -y --chain-id=testing -b block | jq
echo "### Query new admin: $(tgrade q wasm contract $CONTRACT -o json | jq -r '.admin')"
