<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [confio/globalfee/v1beta1/genesis.proto](#confio/globalfee/v1beta1/genesis.proto)
    - [GenesisState](#confio.globalfee.v1beta1.GenesisState)
    - [Params](#confio.globalfee.v1beta1.Params)
  
- [confio/globalfee/v1beta1/query.proto](#confio/globalfee/v1beta1/query.proto)
    - [QueryMinimumGasPricesRequest](#confio.globalfee.v1beta1.QueryMinimumGasPricesRequest)
    - [QueryMinimumGasPricesResponse](#confio.globalfee.v1beta1.QueryMinimumGasPricesResponse)
  
    - [Query](#confio.globalfee.v1beta1.Query)
  
- [confio/poe/v1beta1/poe.proto](#confio/poe/v1beta1/poe.proto)
    - [Params](#confio.poe.v1beta1.Params)
  
    - [PoEContractType](#confio.poe.v1beta1.PoEContractType)
  
- [confio/poe/v1beta1/genesis.proto](#confio/poe/v1beta1/genesis.proto)
    - [CommunityPoolContractConfig](#confio.poe.v1beta1.CommunityPoolContractConfig)
    - [EngagementContractConfig](#confio.poe.v1beta1.EngagementContractConfig)
    - [GenesisState](#confio.poe.v1beta1.GenesisState)
    - [OversightCommitteeContractConfig](#confio.poe.v1beta1.OversightCommitteeContractConfig)
    - [PoEContract](#confio.poe.v1beta1.PoEContract)
    - [StakeContractConfig](#confio.poe.v1beta1.StakeContractConfig)
    - [TG4Member](#confio.poe.v1beta1.TG4Member)
    - [ValidatorVotingContractConfig](#confio.poe.v1beta1.ValidatorVotingContractConfig)
    - [ValsetContractConfig](#confio.poe.v1beta1.ValsetContractConfig)
    - [VotingRules](#confio.poe.v1beta1.VotingRules)
  
- [confio/poe/v1beta1/query.proto](#confio/poe/v1beta1/query.proto)
    - [QueryContractAddressRequest](#confio.poe.v1beta1.QueryContractAddressRequest)
    - [QueryContractAddressResponse](#confio.poe.v1beta1.QueryContractAddressResponse)
    - [QueryUnbondingPeriodRequest](#confio.poe.v1beta1.QueryUnbondingPeriodRequest)
    - [QueryUnbondingPeriodResponse](#confio.poe.v1beta1.QueryUnbondingPeriodResponse)
    - [QueryValidatorDelegationRequest](#confio.poe.v1beta1.QueryValidatorDelegationRequest)
    - [QueryValidatorDelegationResponse](#confio.poe.v1beta1.QueryValidatorDelegationResponse)
    - [QueryValidatorOutstandingRewardRequest](#confio.poe.v1beta1.QueryValidatorOutstandingRewardRequest)
    - [QueryValidatorOutstandingRewardResponse](#confio.poe.v1beta1.QueryValidatorOutstandingRewardResponse)
    - [QueryValidatorUnbondingDelegationsRequest](#confio.poe.v1beta1.QueryValidatorUnbondingDelegationsRequest)
    - [QueryValidatorUnbondingDelegationsResponse](#confio.poe.v1beta1.QueryValidatorUnbondingDelegationsResponse)
  
    - [Query](#confio.poe.v1beta1.Query)
  
- [confio/poe/v1beta1/tx.proto](#confio/poe/v1beta1/tx.proto)
    - [MsgCreateValidator](#confio.poe.v1beta1.MsgCreateValidator)
    - [MsgCreateValidatorResponse](#confio.poe.v1beta1.MsgCreateValidatorResponse)
    - [MsgDelegate](#confio.poe.v1beta1.MsgDelegate)
    - [MsgDelegateResponse](#confio.poe.v1beta1.MsgDelegateResponse)
    - [MsgUndelegate](#confio.poe.v1beta1.MsgUndelegate)
    - [MsgUndelegateResponse](#confio.poe.v1beta1.MsgUndelegateResponse)
    - [MsgUpdateValidator](#confio.poe.v1beta1.MsgUpdateValidator)
    - [MsgUpdateValidatorResponse](#confio.poe.v1beta1.MsgUpdateValidatorResponse)
  
    - [Msg](#confio.poe.v1beta1.Msg)
  
- [confio/twasm/v1beta1/contract_extension.proto](#confio/twasm/v1beta1/contract_extension.proto)
    - [RegisteredPrivilege](#confio.twasm.v1beta1.RegisteredPrivilege)
    - [TgradeContractDetails](#confio.twasm.v1beta1.TgradeContractDetails)
  
- [confio/twasm/v1beta1/genesis.proto](#confio/twasm/v1beta1/genesis.proto)
    - [GenesisState](#confio.twasm.v1beta1.GenesisState)
  
- [confio/twasm/v1beta1/proposal.proto](#confio/twasm/v1beta1/proposal.proto)
    - [DemotePrivilegedContractProposal](#confio.twasm.v1beta1.DemotePrivilegedContractProposal)
    - [PromoteToPrivilegedContractProposal](#confio.twasm.v1beta1.PromoteToPrivilegedContractProposal)
  
- [confio/twasm/v1beta1/query.proto](#confio/twasm/v1beta1/query.proto)
    - [QueryContractsByPrivilegeTypeRequest](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeRequest)
    - [QueryContractsByPrivilegeTypeResponse](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeResponse)
    - [QueryPrivilegedContractsRequest](#confio.twasm.v1beta1.QueryPrivilegedContractsRequest)
    - [QueryPrivilegedContractsResponse](#confio.twasm.v1beta1.QueryPrivilegedContractsResponse)
  
    - [Query](#confio.twasm.v1beta1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="confio/globalfee/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/globalfee/v1beta1/genesis.proto



<a name="confio.globalfee.v1beta1.GenesisState"></a>

### GenesisState
GenesisState - initial state of module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#confio.globalfee.v1beta1.Params) |  | Params of this module |






<a name="confio.globalfee.v1beta1.Params"></a>

### Params
Params defines the set of module parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | Minimum stores the minimum gas price(s) for all TX on the chain. When multiple coins are defined then they are accepted alternatively. The list must be sorted by denoms asc. No duplicate denoms or zero amount values allowed. For more information see https://docs.cosmos.network/master/modules/auth/01_concepts.html |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/globalfee/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/globalfee/v1beta1/query.proto



<a name="confio.globalfee.v1beta1.QueryMinimumGasPricesRequest"></a>

### QueryMinimumGasPricesRequest
QueryMinimumGasPricesRequest is the request type for the
Query/MinimumGasPrices RPC method.






<a name="confio.globalfee.v1beta1.QueryMinimumGasPricesResponse"></a>

### QueryMinimumGasPricesResponse
QueryMinimumGasPricesResponse is the response type for the
Query/MinimumGasPrices RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `minimum_gas_prices` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.globalfee.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `MinimumGasPrices` | [QueryMinimumGasPricesRequest](#confio.globalfee.v1beta1.QueryMinimumGasPricesRequest) | [QueryMinimumGasPricesResponse](#confio.globalfee.v1beta1.QueryMinimumGasPricesResponse) |  | GET|/tgrade/globalfee/v1beta1/minimum_gas_prices|

 <!-- end services -->



<a name="confio/poe/v1beta1/poe.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/poe.proto



<a name="confio.poe.v1beta1.Params"></a>

### Params
Params defines the parameters for the PoE module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `historical_entries` | [uint32](#uint32) |  | HistoricalEntries is the number of historical entries to persist. |
| `initial_val_engagement_points` | [uint64](#uint64) |  | InitialValEngagementPoints defines the number of engagement for any new validator joining post genesis |
| `min_delegation_amounts` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | MinDelegationAmount defines the minimum amount a post genesis validator needs to self delegate to receive any engagement points. One must be exceeded. No minimum condition set when empty. |





 <!-- end messages -->


<a name="confio.poe.v1beta1.PoEContractType"></a>

### PoEContractType
PoEContractType type of PoE contract

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNDEFINED | 0 |  |
| STAKING | 1 |  |
| VALSET | 2 |  |
| ENGAGEMENT | 3 |  |
| MIXER | 4 |  |
| DISTRIBUTION | 5 |  |
| OVERSIGHT_COMMUNITY | 6 |  |
| OVERSIGHT_COMMUNITY_PROPOSALS | 7 |  |
| COMMUNITY_POOL | 8 |  |
| VALIDATOR_VOTING | 9 |  |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/poe/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/genesis.proto



<a name="confio.poe.v1beta1.CommunityPoolContractConfig"></a>

### CommunityPoolContractConfig
CommunityPoolContractConfig initial setup config for the contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `voting_rules` | [VotingRules](#confio.poe.v1beta1.VotingRules) |  | VotingRules rules for the tally |






<a name="confio.poe.v1beta1.EngagementContractConfig"></a>

### EngagementContractConfig
EngagementContractConfig initial setup config


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `halflife` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |






<a name="confio.poe.v1beta1.GenesisState"></a>

### GenesisState
GenesisState - initial state of module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#confio.poe.v1beta1.Params) |  | params defines all the parameter of the module |
| `seed_contracts` | [bool](#bool) |  | SeedContracts when enabled stores and instantiates the Proof of Engagement contracts on the chain. |
| `gen_txs` | [bytes](#bytes) | repeated | GenTxs defines the genesis transactions to create a validator. |
| `system_admin_address` | [string](#string) |  | SystemAdminAddress single address that is set as admin for the PoE contracts in seed mode. |
| `contracts` | [PoEContract](#confio.poe.v1beta1.PoEContract) | repeated | Contracts Poe contract addresses and types when used with state dump in non seed mode. |
| `engagement` | [TG4Member](#confio.poe.v1beta1.TG4Member) | repeated | Engagement weighted members of the engagement group. Validators should be in here. |
| `stake_contract_config` | [StakeContractConfig](#confio.poe.v1beta1.StakeContractConfig) |  |  |
| `valset_contract_config` | [ValsetContractConfig](#confio.poe.v1beta1.ValsetContractConfig) |  |  |
| `engagement_contract_config` | [EngagementContractConfig](#confio.poe.v1beta1.EngagementContractConfig) |  |  |
| `bond_denom` | [string](#string) |  | BondDenom defines the bondable coin denomination. |
| `oversight_committee_contract_config` | [OversightCommitteeContractConfig](#confio.poe.v1beta1.OversightCommitteeContractConfig) |  |  |
| `community_pool_contract_config` | [CommunityPoolContractConfig](#confio.poe.v1beta1.CommunityPoolContractConfig) |  |  |
| `validator_voting_contract_config` | [ValidatorVotingContractConfig](#confio.poe.v1beta1.ValidatorVotingContractConfig) |  |  |






<a name="confio.poe.v1beta1.OversightCommitteeContractConfig"></a>

### OversightCommitteeContractConfig
OversightCommitteeContractConfig initial setup config for the trusted circle


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `name` | [string](#string) |  | Name of TRUSTED_CIRCLE |
| `escrow_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | EscrowAmount The required escrow amount, in the default denom (utgd) |
| `voting_rules` | [VotingRules](#confio.poe.v1beta1.VotingRules) |  | VotingRules rules for the tally |
| `deny_list_contract_address` | [string](#string) |  | DenyListContractAddress is an optional cw4 contract with list of addresses denied to be part of TrustedCircle |






<a name="confio.poe.v1beta1.PoEContract"></a>

### PoEContract
PoEContract address and type information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_type` | [PoEContractType](#confio.poe.v1beta1.PoEContractType) |  | ContractType type. |
| `address` | [string](#string) |  | Address is the bech32 address string |






<a name="confio.poe.v1beta1.StakeContractConfig"></a>

### StakeContractConfig
StakeContractConfig initial setup config


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_bond` | [uint64](#uint64) |  |  |
| `tokens_per_point` | [uint64](#uint64) |  |  |
| `unbonding_period` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `claim_autoreturn_limit` | [uint32](#uint32) |  |  |






<a name="confio.poe.v1beta1.TG4Member"></a>

### TG4Member
TG4Member member of the Engagement group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address string |
| `points` | [uint64](#uint64) |  |  |






<a name="confio.poe.v1beta1.ValidatorVotingContractConfig"></a>

### ValidatorVotingContractConfig
ValidatorVotingContractConfig CommunityPoolContractConfig


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `voting_rules` | [VotingRules](#confio.poe.v1beta1.VotingRules) |  | VotingRules rules for the tally |






<a name="confio.poe.v1beta1.ValsetContractConfig"></a>

### ValsetContractConfig
ValsetContractConfig initial setup config


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `min_points` | [uint64](#uint64) |  |  |
| `max_validators` | [uint32](#uint32) |  |  |
| `epoch_length` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `epoch_reward` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `scaling` | [uint32](#uint32) |  | Scaling is the factor to multiply cw4-group weights to produce the Tendermint validator power |
| `fee_percentage` | [string](#string) |  | FeePercentage is the percentage of total accumulated fees that is subtracted from tokens minted as rewards. 50% by default. To disable this feature just set it to 0 (which effectively means that fees don't affect the per-epoch reward). |
| `community_pool_reward_ratio` | [string](#string) |  | CommunityPoolRewardRation in percentage |
| `engagement_reward_ratio` | [string](#string) |  | EngagementRewardRatio reward ration in percentage for all |
| `validator_reward_ratio` | [string](#string) |  | ValidatorRewardRation in percentage for all |
| `auto_unjail` | [bool](#bool) |  | AutoUnjail if set to true, we will auto-unjail any validator after their jailtime is over. |
| `double_sign_slash_ratio` | [string](#string) |  | DoubleSignSlashRatio Validators who are caught double signing are jailed forever and their bonded tokens are slashed based on this value. |






<a name="confio.poe.v1beta1.VotingRules"></a>

### VotingRules
VotingRules contains configuration for the tally.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `voting_period` | [uint32](#uint32) |  | VotingPeriod Voting period in days |
| `quorum` | [string](#string) |  | Quorum voting quorum percentage (1-100) |
| `threshold` | [string](#string) |  | Threshold voting threshold percentage (50-100) |
| `allow_end_early` | [bool](#bool) |  | AllowEndEarly If true, and absolute threshold and quorum are met, we can end before voting period finished. (Recommended value: true, unless you have special needs) |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/poe/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/query.proto



<a name="confio.poe.v1beta1.QueryContractAddressRequest"></a>

### QueryContractAddressRequest
QueryContractAddressRequest is the request type for the Query/ContractAddress
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_type` | [PoEContractType](#confio.poe.v1beta1.PoEContractType) |  | ContractType is the type of contract |






<a name="confio.poe.v1beta1.QueryContractAddressResponse"></a>

### QueryContractAddressResponse
QueryContractAddressRequest is the response type for the
Query/ContractAddress RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="confio.poe.v1beta1.QueryUnbondingPeriodRequest"></a>

### QueryUnbondingPeriodRequest
QueryUnbondingPeriodRequest is request type for the Query/UnbondingPeriod RPC
method






<a name="confio.poe.v1beta1.QueryUnbondingPeriodResponse"></a>

### QueryUnbondingPeriodResponse
QueryUnbondingPeriodResponse is response type for the Query/UnbondingPeriod
RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `time` | [google.protobuf.Duration](#google.protobuf.Duration) |  | Time is the time that must pass |






<a name="confio.poe.v1beta1.QueryValidatorDelegationRequest"></a>

### QueryValidatorDelegationRequest
QueryValidatorDelegationRequest is request type for the
Query/ValidatorDelegation RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator_addr` | [string](#string) |  | validator_addr defines the validator address to query for. |






<a name="confio.poe.v1beta1.QueryValidatorDelegationResponse"></a>

### QueryValidatorDelegationResponse
QueryValidatorDelegationResponse is response type for the
Query/ValidatorDelegation RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `balance` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="confio.poe.v1beta1.QueryValidatorOutstandingRewardRequest"></a>

### QueryValidatorOutstandingRewardRequest
QueryValidatorOutstandingRewardRequest is the request type for the
Query/ValidatorOutstandingReward RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator_address` | [string](#string) |  | validator_address defines the validator address to query for. |






<a name="confio.poe.v1beta1.QueryValidatorOutstandingRewardResponse"></a>

### QueryValidatorOutstandingRewardResponse
QueryValidatorOutstandingRewardResponse is the response type for the
Query/ValidatorOutstandingReward RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `reward` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |






<a name="confio.poe.v1beta1.QueryValidatorUnbondingDelegationsRequest"></a>

### QueryValidatorUnbondingDelegationsRequest
QueryValidatorUnbondingDelegationsRequest is required type for the
Query/ValidatorUnbondingDelegations RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `validator_addr` | [string](#string) |  | validator_addr defines the validator address to query for. |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="confio.poe.v1beta1.QueryValidatorUnbondingDelegationsResponse"></a>

### QueryValidatorUnbondingDelegationsResponse
QueryValidatorUnbondingDelegationsResponse is response type for the
Query/ValidatorUnbondingDelegations RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `entries` | [cosmos.staking.v1beta1.UnbondingDelegationEntry](#cosmos.staking.v1beta1.UnbondingDelegationEntry) | repeated | unbonding delegation entries |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.poe.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ContractAddress` | [QueryContractAddressRequest](#confio.poe.v1beta1.QueryContractAddressRequest) | [QueryContractAddressResponse](#confio.poe.v1beta1.QueryContractAddressResponse) | ContractAddress queries the address for one of the PoE contracts | GET|/tgrade/poe/v1beta1/contract/{contract_type}|
| `Validators` | [.cosmos.staking.v1beta1.QueryValidatorsRequest](#cosmos.staking.v1beta1.QueryValidatorsRequest) | [.cosmos.staking.v1beta1.QueryValidatorsResponse](#cosmos.staking.v1beta1.QueryValidatorsResponse) | Validators queries all validators that match the given status. | GET|/tgrade/poe/v1beta1/validators|
| `Validator` | [.cosmos.staking.v1beta1.QueryValidatorRequest](#cosmos.staking.v1beta1.QueryValidatorRequest) | [.cosmos.staking.v1beta1.QueryValidatorResponse](#cosmos.staking.v1beta1.QueryValidatorResponse) | Validator queries validator info for given validator address. | GET|/tgrade/poe/v1beta1/validators/{validator_addr}|
| `UnbondingPeriod` | [QueryUnbondingPeriodRequest](#confio.poe.v1beta1.QueryUnbondingPeriodRequest) | [QueryUnbondingPeriodResponse](#confio.poe.v1beta1.QueryUnbondingPeriodResponse) | Validator queries validator info for given validator address. | GET|/tgrade/poe/v1beta1/unbonding|
| `ValidatorDelegation` | [QueryValidatorDelegationRequest](#confio.poe.v1beta1.QueryValidatorDelegationRequest) | [QueryValidatorDelegationResponse](#confio.poe.v1beta1.QueryValidatorDelegationResponse) | ValidatorDelegation queries self delegated amount for given validator. | GET|/poe/tgrade/v1beta1/validators/{validator_addr}/delegation|
| `ValidatorUnbondingDelegations` | [QueryValidatorUnbondingDelegationsRequest](#confio.poe.v1beta1.QueryValidatorUnbondingDelegationsRequest) | [QueryValidatorUnbondingDelegationsResponse](#confio.poe.v1beta1.QueryValidatorUnbondingDelegationsResponse) | ValidatorUnbondingDelegations queries unbonding delegations of a validator. | GET|/tgrade/poe/v1beta1/validators/{validator_addr}/unbonding_delegations|
| `HistoricalInfo` | [.cosmos.staking.v1beta1.QueryHistoricalInfoRequest](#cosmos.staking.v1beta1.QueryHistoricalInfoRequest) | [.cosmos.staking.v1beta1.QueryHistoricalInfoResponse](#cosmos.staking.v1beta1.QueryHistoricalInfoResponse) | HistoricalInfo queries the historical info for given height. | GET|/tgrade/poe/v1beta1/historical_info/{height}|
| `ValidatorOutstandingReward` | [QueryValidatorOutstandingRewardRequest](#confio.poe.v1beta1.QueryValidatorOutstandingRewardRequest) | [QueryValidatorOutstandingRewardResponse](#confio.poe.v1beta1.QueryValidatorOutstandingRewardResponse) | ValidatorOutstandingRewards queries rewards of a validator address. | GET|/tgrade/poe/v1beta1/validators/{validator_address}/outstanding_reward|

 <!-- end services -->



<a name="confio/poe/v1beta1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/tx.proto



<a name="confio.poe.v1beta1.MsgCreateValidator"></a>

### MsgCreateValidator
MsgCreateValidator defines a PoE message for creating a new validator.
Based on the SDK staking.MsgCreateValidator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `description` | [cosmos.staking.v1beta1.Description](#cosmos.staking.v1beta1.Description) |  | Description meta data |
| `operator_address` | [string](#string) |  | OperatorAddress is the bech32 address string |
| `pubkey` | [google.protobuf.Any](#google.protobuf.Any) |  | Pubkey public key |
| `value` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Value defines the initial staking amount |






<a name="confio.poe.v1beta1.MsgCreateValidatorResponse"></a>

### MsgCreateValidatorResponse
MsgCreateValidatorResponse defines the MsgCreateValidator response type.






<a name="confio.poe.v1beta1.MsgDelegate"></a>

### MsgDelegate
MsgDelegate defines a SDK message for performing a self delegation of coins
by a node operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |
| `vesting_amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="confio.poe.v1beta1.MsgDelegateResponse"></a>

### MsgDelegateResponse
MsgDelegateResponse defines the Msg/Delegate response type.






<a name="confio.poe.v1beta1.MsgUndelegate"></a>

### MsgUndelegate
MsgUndelegate defines a SDK message for performing an undelegation from a
node operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `operator_address` | [string](#string) |  |  |
| `amount` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  |  |






<a name="confio.poe.v1beta1.MsgUndelegateResponse"></a>

### MsgUndelegateResponse
MsgUndelegateResponse defines the Msg/Undelegate response type.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `completion_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |






<a name="confio.poe.v1beta1.MsgUpdateValidator"></a>

### MsgUpdateValidator
MsgCreateValidator defines a PoE message for updating validator metadata


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `description` | [cosmos.staking.v1beta1.Description](#cosmos.staking.v1beta1.Description) |  | New Description meta data |
| `operator_address` | [string](#string) |  | OperatorAddress is the bech32 address string Also know as "signer" in other messages |






<a name="confio.poe.v1beta1.MsgUpdateValidatorResponse"></a>

### MsgUpdateValidatorResponse
MsgUpdateValidatorResponse defines the MsgUpdateValidator response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.poe.v1beta1.Msg"></a>

### Msg
Msg defines the staking Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateValidator` | [MsgCreateValidator](#confio.poe.v1beta1.MsgCreateValidator) | [MsgCreateValidatorResponse](#confio.poe.v1beta1.MsgCreateValidatorResponse) | CreateValidator defines a method for creating a new validator. | |
| `UpdateValidator` | [MsgUpdateValidator](#confio.poe.v1beta1.MsgUpdateValidator) | [MsgUpdateValidatorResponse](#confio.poe.v1beta1.MsgUpdateValidatorResponse) | MsgCreateValidator defines a method for updating validator metadata | |
| `Delegate` | [MsgDelegate](#confio.poe.v1beta1.MsgDelegate) | [MsgDelegateResponse](#confio.poe.v1beta1.MsgDelegateResponse) | Delegate defines a method for performing a self delegation of coins by a node operator | |
| `Undelegate` | [MsgUndelegate](#confio.poe.v1beta1.MsgUndelegate) | [MsgUndelegateResponse](#confio.poe.v1beta1.MsgUndelegateResponse) | Undelegate defines a method for performing an undelegation from a node operator | |

 <!-- end services -->



<a name="confio/twasm/v1beta1/contract_extension.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/twasm/v1beta1/contract_extension.proto



<a name="confio.twasm.v1beta1.RegisteredPrivilege"></a>

### RegisteredPrivilege
RegisteredPrivilege stores position and privilege name


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `position` | [uint32](#uint32) |  |  |
| `privilege_type` | [string](#string) |  |  |






<a name="confio.twasm.v1beta1.TgradeContractDetails"></a>

### TgradeContractDetails
TgradeContractDetails is a custom extension to the wasmd ContractInfo


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `registered_privileges` | [RegisteredPrivilege](#confio.twasm.v1beta1.RegisteredPrivilege) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/twasm/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/twasm/v1beta1/genesis.proto



<a name="confio.twasm.v1beta1.GenesisState"></a>

### GenesisState



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `wasm` | [cosmwasm.wasm.v1.GenesisState](#cosmwasm.wasm.v1.GenesisState) |  |  |
| `privileged_contract_addresses` | [string](#string) | repeated |  |
| `pinned_code_ids` | [uint64](#uint64) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/twasm/v1beta1/proposal.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/twasm/v1beta1/proposal.proto



<a name="confio.twasm.v1beta1.DemotePrivilegedContractProposal"></a>

### DemotePrivilegedContractProposal
PromoteToPrivilegedContractProposal gov proposal content type to remove
"privileges" from a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |






<a name="confio.twasm.v1beta1.PromoteToPrivilegedContractProposal"></a>

### PromoteToPrivilegedContractProposal
PromoteToPrivilegedContractProposal gov proposal content type to add
"privileges" to a contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | Title is a short summary |
| `description` | [string](#string) |  | Description is a human readable text |
| `contract` | [string](#string) |  | Contract is the address of the smart contract |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/twasm/v1beta1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/twasm/v1beta1/query.proto



<a name="confio.twasm.v1beta1.QueryContractsByPrivilegeTypeRequest"></a>

### QueryContractsByPrivilegeTypeRequest
QueryContractsByPrivilegeTypeRequest is the request type for the
Query/ContractsByPrivilegeType RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `privilege_type` | [string](#string) |  |  |






<a name="confio.twasm.v1beta1.QueryContractsByPrivilegeTypeResponse"></a>

### QueryContractsByPrivilegeTypeResponse
QueryContractsByPrivilegeTypeResponse is the response type for the
Query/ContractsByPrivilegeType RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contracts` | [string](#string) | repeated | contracts are a set of contract addresses |






<a name="confio.twasm.v1beta1.QueryPrivilegedContractsRequest"></a>

### QueryPrivilegedContractsRequest
QueryPrivilegedContractsResponse is the request type for the
Query/PrivilegedContracts RPC method






<a name="confio.twasm.v1beta1.QueryPrivilegedContractsResponse"></a>

### QueryPrivilegedContractsResponse
QueryPrivilegedContractsResponse is the response type for the
Query/PrivilegedContracts RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contracts` | [string](#string) | repeated | contracts are a set of contract addresses |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.twasm.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PrivilegedContracts` | [QueryPrivilegedContractsRequest](#confio.twasm.v1beta1.QueryPrivilegedContractsRequest) | [QueryPrivilegedContractsResponse](#confio.twasm.v1beta1.QueryPrivilegedContractsResponse) | PrivilegedContracts returns all privileged contracts | GET|/tgrade/twasm/v1beta1/contracts/privileged|
| `ContractsByPrivilegeType` | [QueryContractsByPrivilegeTypeRequest](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeRequest) | [QueryContractsByPrivilegeTypeResponse](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeResponse) | ContractsByPrivilegeType returns all contracts that have registered for the privilege type | GET|/tgrade/twasm/v1beta1/contracts/privilege/{privilege_type}|

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

