# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [pkg/proto/agentrpc.proto](#pkg_proto_agentrpc-proto)
    - [ReloadRequest](#mcing-ReloadRequest)
    - [ReloadResponse](#mcing-ReloadResponse)
    - [SaveAllFlushRequest](#mcing-SaveAllFlushRequest)
    - [SaveAllFlushResponse](#mcing-SaveAllFlushResponse)
    - [SaveOffRequest](#mcing-SaveOffRequest)
    - [SaveOffResponse](#mcing-SaveOffResponse)
    - [SaveOnRequest](#mcing-SaveOnRequest)
    - [SaveOnResponse](#mcing-SaveOnResponse)
    - [SyncOpsRequest](#mcing-SyncOpsRequest)
    - [SyncOpsResponse](#mcing-SyncOpsResponse)
    - [SyncWhitelistRequest](#mcing-SyncWhitelistRequest)
    - [SyncWhitelistResponse](#mcing-SyncWhitelistResponse)
  
    - [Agent](#mcing-Agent)
  
- [Scalar Value Types](#scalar-value-types)



<a name="pkg_proto_agentrpc-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pkg/proto/agentrpc.proto



<a name="mcing-ReloadRequest"></a>

### ReloadRequest
ReloadRequest is the request message to execute `/reload` via rcon.






<a name="mcing-ReloadResponse"></a>

### ReloadResponse
ReloadResponse is the response message of Reload






<a name="mcing-SaveAllFlushRequest"></a>

### SaveAllFlushRequest







<a name="mcing-SaveAllFlushResponse"></a>

### SaveAllFlushResponse







<a name="mcing-SaveOffRequest"></a>

### SaveOffRequest







<a name="mcing-SaveOffResponse"></a>

### SaveOffResponse







<a name="mcing-SaveOnRequest"></a>

### SaveOnRequest







<a name="mcing-SaveOnResponse"></a>

### SaveOnResponse







<a name="mcing-SyncOpsRequest"></a>

### SyncOpsRequest
SyncWhitelistRequest is the request message to exec /whitelist via rcon


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [string](#string) | repeated |  |






<a name="mcing-SyncOpsResponse"></a>

### SyncOpsResponse
SyncOpsResponse is the response message of SyncOps






<a name="mcing-SyncWhitelistRequest"></a>

### SyncWhitelistRequest
SyncWhitelistRequest is the request message to exec /whitelist via rcon


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | [bool](#bool) |  |  |
| users | [string](#string) | repeated |  |






<a name="mcing-SyncWhitelistResponse"></a>

### SyncWhitelistResponse
SyncWhitelistResponse is the response message of SyncWhitelist





 

 

 


<a name="mcing-Agent"></a>

### Agent
Agent provides services for MCing.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Reload | [ReloadRequest](#mcing-ReloadRequest) | [ReloadResponse](#mcing-ReloadResponse) |  |
| SyncWhitelist | [SyncWhitelistRequest](#mcing-SyncWhitelistRequest) | [SyncWhitelistResponse](#mcing-SyncWhitelistResponse) |  |
| SyncOps | [SyncOpsRequest](#mcing-SyncOpsRequest) | [SyncOpsResponse](#mcing-SyncOpsResponse) |  |
| SaveOff | [SaveOffRequest](#mcing-SaveOffRequest) | [SaveOffResponse](#mcing-SaveOffResponse) |  |
| SaveAllFlush | [SaveAllFlushRequest](#mcing-SaveAllFlushRequest) | [SaveAllFlushResponse](#mcing-SaveAllFlushResponse) |  |
| SaveOn | [SaveOnRequest](#mcing-SaveOnRequest) | [SaveOnResponse](#mcing-SaveOnResponse) |  |

 



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

