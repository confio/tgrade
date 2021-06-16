<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [confio/globalfee/v1beta1/genesis.proto](#confio/globalfee/v1beta1/genesis.proto)
    - [GenesisState](#confio.globalfee.v1beta1.GenesisState)
    - [Params](#confio.globalfee.v1beta1.Params)
  
- [confio/poe/v1beta1/poe.proto](#confio/poe/v1beta1/poe.proto)
    - [PoEContractType](#confio.poe.v1beta1.PoEContractType)
  
- [confio/poe/v1beta1/genesis.proto](#confio/poe/v1beta1/genesis.proto)
    - [GenesisState](#confio.poe.v1beta1.GenesisState)
    - [PoEContract](#confio.poe.v1beta1.PoEContract)
    - [TG4Member](#confio.poe.v1beta1.TG4Member)
  
- [confio/poe/v1beta1/query.proto](#confio/poe/v1beta1/query.proto)
    - [QueryContractAddressRequest](#confio.poe.v1beta1.QueryContractAddressRequest)
    - [QueryContractAddressResponse](#confio.poe.v1beta1.QueryContractAddressResponse)
  
    - [Query](#confio.poe.v1beta1.Query)
  
- [confio/poe/v1beta1/tx.proto](#confio/poe/v1beta1/tx.proto)
    - [MsgCreateValidator](#confio.poe.v1beta1.MsgCreateValidator)
    - [MsgCreateValidatorResponse](#confio.poe.v1beta1.MsgCreateValidatorResponse)
  
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



<a name="confio/poe/v1beta1/poe.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/poe.proto


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


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="confio/poe/v1beta1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/poe/v1beta1/genesis.proto



<a name="confio.poe.v1beta1.GenesisState"></a>

### GenesisState
GenesisState - initial state of module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `seed_contracts` | [bool](#bool) |  | SeedContracts when enabled stores and instantiates the Proof of Engagement contracts on the chain. |
| `gen_txs` | [bytes](#bytes) | repeated | GenTxs defines the genesis transactions to create a validator. |
| `system_admin_address` | [string](#string) |  | SystemAdminAddress single address that is set as admin for the PoE contracts in seed mode. |
| `contracts` | [PoEContract](#confio.poe.v1beta1.PoEContract) | repeated | Contracts Poe contract addresses and types when used with state dump in non seed mode. |
| `engagement` | [TG4Member](#confio.poe.v1beta1.TG4Member) | repeated | Engagement weighted members of the engagement group. Validators should be in here. |
| `bond_denom` | [string](#string) |  | BondDenom defines the bondable coin denomination. |






<a name="confio.poe.v1beta1.PoEContract"></a>

### PoEContract
PoEContract address and type information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_type` | [PoEContractType](#confio.poe.v1beta1.PoEContractType) |  | ContractType type. |
| `address` | [string](#string) |  | Address is the bech32 address string |






<a name="confio.poe.v1beta1.TG4Member"></a>

### TG4Member
TG4Member member of the Engagement group.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | Address is the bech32 address string |
| `weight` | [uint64](#uint64) |  |  |





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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.poe.v1beta1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ContractAddress` | [QueryContractAddressRequest](#confio.poe.v1beta1.QueryContractAddressRequest) | [QueryContractAddressResponse](#confio.poe.v1beta1.QueryContractAddressResponse) |  | GET|/poe/v1beta1/contract/{contract_type}|

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
| `delegator_address` | [string](#string) |  | DelegatorAddress is the bech32 address string |
| `pubkey` | [google.protobuf.Any](#google.protobuf.Any) |  | Pubkey public key |
| `value` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Value defines the initial staking amount |






<a name="confio.poe.v1beta1.MsgCreateValidatorResponse"></a>

### MsgCreateValidatorResponse
MsgCreateValidatorResponse defines the Msg/CreateValidator response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.poe.v1beta1.Msg"></a>

### Msg
Msg defines the staking Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateValidator` | [MsgCreateValidator](#confio.poe.v1beta1.MsgCreateValidator) | [MsgCreateValidatorResponse](#confio.poe.v1beta1.MsgCreateValidatorResponse) | CreateValidator defines a method for creating a new validator. | |

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
| `wasm` | [cosmwasm.wasm.v1beta1.GenesisState](#cosmwasm.wasm.v1beta1.GenesisState) |  |  |
| `privileged_contract_addresses` | [string](#string) | repeated |  |





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
| `PrivilegedContracts` | [QueryPrivilegedContractsRequest](#confio.twasm.v1beta1.QueryPrivilegedContractsRequest) | [QueryPrivilegedContractsResponse](#confio.twasm.v1beta1.QueryPrivilegedContractsResponse) | PrivilegedContracts returns all privileged contracts | GET|/twasm/v1beta1/contracts/privileged|
| `ContractsByPrivilegeType` | [QueryContractsByPrivilegeTypeRequest](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeRequest) | [QueryContractsByPrivilegeTypeResponse](#confio.twasm.v1beta1.QueryContractsByPrivilegeTypeResponse) | ContractsByPrivilegeType returns all contracts that have registered for the privilege type | GET|/twasm/v1beta1/contracts/privilege/{privilege_type}|

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

