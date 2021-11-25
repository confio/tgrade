#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck > /dev/null && shellcheck "$0"

if [ $# -ne 1 ]; then
  echo "Usage: ./download_releases.sh RELEASE_TAG"
  exit 1
fi

tag="$1"
# load token from OS keychain when not set via ENV
GITHUB_API_TOKEN=${GITHUB_API_TOKEN:-"$(security find-generic-password -a "$USER" -s "github_api_key" -w)"}

rm -f version.txt
for contract in tg4_engagement tgrade_valset tg4_mixer tg4_stake tgrade_trusted_circle tgrade_gov_reflect tgrade_oc_proposals tgrade_community_pool tgrade_validator_voting; do
  echo "Download $contract"
  list_asset_url="https://api.github.com/repos/confio/tgrade-contracts/releases/tags/${tag}"
  # get url for artifact with name==$artifact
  asset_url=$(curl -H "Accept: application/vnd.github.v3+json" -H "Authorization: token $GITHUB_API_TOKEN" "${list_asset_url}" | jq -r ".assets[] | select(.name==\"${contract}.wasm\") | .url")
  rm -f "./${contract}.wasm"
  # download the artifact
  curl -LJO -H 'Accept: application/octet-stream' -H "Authorization: token $GITHUB_API_TOKEN" "$asset_url"
done

rm -f version.txt
echo "$tag" >version.txt
