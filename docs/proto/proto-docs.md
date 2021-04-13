<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [confio/twasm/v1beta1/contract_extension.proto](#confio/twasm/v1beta1/contract_extension.proto)
    - [RegisteredCallback](#confio.twasm.v1beta1.RegisteredCallback)
    - [TgradeContractDetails](#confio.twasm.v1beta1.TgradeContractDetails)
  
- [confio/twasm/v1beta1/genesis.proto](#confio/twasm/v1beta1/genesis.proto)
    - [GenesisState](#confio.twasm.v1beta1.GenesisState)
  
- [confio/twasm/v1beta1/proposal.proto](#confio/twasm/v1beta1/proposal.proto)
    - [DemotePrivilegedContractProposal](#confio.twasm.v1beta1.DemotePrivilegedContractProposal)
    - [PromoteToPrivilegedContractProposal](#confio.twasm.v1beta1.PromoteToPrivilegedContractProposal)
  
- [confio/twasm/v1beta1/query.proto](#confio/twasm/v1beta1/query.proto)
    - [QueryContractsByCallbackTypeRequest](#confio.twasm.v1beta1.QueryContractsByCallbackTypeRequest)
    - [QueryContractsByCallbackTypeResponse](#confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse)
    - [QueryContractsByCallbackTypeResponse.ContractPosition](#confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse.ContractPosition)
    - [QueryPrivilegedContractsRequest](#confio.twasm.v1beta1.QueryPrivilegedContractsRequest)
    - [QueryPrivilegedContractsResponse](#confio.twasm.v1beta1.QueryPrivilegedContractsResponse)
  
    - [Query](#confio.twasm.v1beta1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="confio/twasm/v1beta1/contract_extension.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## confio/twasm/v1beta1/contract_extension.proto



<a name="confio.twasm.v1beta1.RegisteredCallback"></a>

### RegisteredCallback



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `position` | [uint32](#uint32) |  |  |
| `callback_type` | [uint32](#uint32) |  |  |






<a name="confio.twasm.v1beta1.TgradeContractDetails"></a>

### TgradeContractDetails



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `registered_callbacks` | [RegisteredCallback](#confio.twasm.v1beta1.RegisteredCallback) | repeated |  |





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



<a name="confio.twasm.v1beta1.QueryContractsByCallbackTypeRequest"></a>

### QueryContractsByCallbackTypeRequest
QueryContractsByCallbackTypeRequest is the request type for the
Query/ContractsByCallbackType RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `callback_type` | [string](#string) |  |  |






<a name="confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse"></a>

### QueryContractsByCallbackTypeResponse
QueryContractsByCallbackTypeResponse is the response type for the
Query/ContractsByCallbackType RPC method


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contracts` | [QueryContractsByCallbackTypeResponse.ContractPosition](#confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse.ContractPosition) | repeated | addresses set of contract addresses |






<a name="confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse.ContractPosition"></a>

### QueryContractsByCallbackTypeResponse.ContractPosition



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `addresses` | [string](#string) |  |  |






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
| `addresses` | [string](#string) | repeated | addresses set of contract addresses |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="confio.twasm.v1beta1.Query"></a>

### Query
Query provides defines the gRPC querier service

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `PrivilegedContracts` | [QueryPrivilegedContractsRequest](#confio.twasm.v1beta1.QueryPrivilegedContractsRequest) | [QueryPrivilegedContractsResponse](#confio.twasm.v1beta1.QueryPrivilegedContractsResponse) | PrivilegedContracts returns all privileged contracts | GET|/twasm/v1beta1/contracts/privileged|
| `ContractsByCallbackType` | [QueryContractsByCallbackTypeRequest](#confio.twasm.v1beta1.QueryContractsByCallbackTypeRequest) | [QueryContractsByCallbackTypeResponse](#confio.twasm.v1beta1.QueryContractsByCallbackTypeResponse) | ContractsByCallbackType returns all contracts that have registered for the callback type | GET|/twasm/v1beta1/contracts/callback/{callback_type}|

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

