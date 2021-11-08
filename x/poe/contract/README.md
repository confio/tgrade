# Contracts
Tgrade smart contracts. See https://github.com/confio/tgrade-contracts for source code and details.

![Arch with Gov](https://github.com/confio/tgrade-contracts/blob/main/docs/Architecture.md)
[Architecture](https://github.com/confio/tgrade-contracts/blob/main/docs/Architecture.md)

## Developers
### Test strategy
**Contract interactions (query/ updates)** should be covered by **integration tests** that talk to the real contract(s)
We need to ensure that things work as expected but also to provide ["consumer driven contracts"](https://martinfowler.com/articles/consumerDrivenContracts.html) 
in code that couple both worlds to be stable.
Though there are use cases that would also be covered by system tests like create-validator or end block callback. In this
case we have enough confidence. 

Pure Go code should be unit tested only


### Download new artifacts
**Requires a github access token**

Example to download contract for tag `v0.1.3`
```sh
./download_releases.sh v0.1.3
```

### OSX
### Preparation
1. Create access token on [github](https://github.com/settings/tokens) with access to private repos!

#### Store key in OS keychain
1. Copy token to clipboard
1. Add to your OS keychain
```sh
security add-generic-password -a "$USER" -s 'github_api_key' -w "$(pbpaste)"
```
The script will read the token from the OS keychain

### Others
Set your api key via environment `GITHUB_API_TOKEN` variable before running the download script