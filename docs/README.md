# API 文档

本目录包含项目的 API 文档，使用 protoc-gen-doc 插件从 protobuf 文件自动生成。

## 文档导航

### API 参考文档
- [API 接口文档](./index.html) - 完整的 API 接口说明
- [错误码列表](./guide/zh-CN/api/errors-code/apiserver/v1/errors_code.md) - 所有可用的错误码

### 开发指南
- [错误处理快速开始](./guide/zh-CN/errors-quickstart.md) - 5分钟快速上手错误处理
- [错误处理使用指南](./guide/zh-CN/errors-usage.md) - 完整的错误处理机制说明
- [错误处理最佳实践](./guide/zh-CN/errors-best-practices.md) - 错误处理的最佳实践和规范
- [验证使用指南](./validation-usage.md) - 请求验证机制说明

# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [apiserver/v1/apiserver.proto](#apiserver_v1_apiserver-proto)
    - [GetUserRequest](#apiserver-v1-GetUserRequest)
    - [GetUserResponse](#apiserver-v1-GetUserResponse)
  
    - [ApiServer](#apiserver-v1-ApiServer)
  
- [apiserver/v1/errors.proto](#apiserver_v1_errors-proto)
    - [ErrorReason](#apiserver-v1-ErrorReason)
  
- [Scalar Value Types](#scalar-value-types)



<a name="apiserver_v1_apiserver-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## apiserver/v1/apiserver.proto



<a name="apiserver-v1-GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="apiserver-v1-GetUserResponse"></a>

### GetUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| name | [string](#string) |  |  |
| email | [string](#string) |  |  |





 

 

 


<a name="apiserver-v1-ApiServer"></a>

### ApiServer


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetUser | [GetUserRequest](#apiserver-v1-GetUserRequest) | [GetUserResponse](#apiserver-v1-GetUserResponse) |  |

 



<a name="apiserver_v1_errors-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## apiserver/v1/errors.proto


 


<a name="apiserver-v1-ErrorReason"></a>

### ErrorReason


| Name | Number | Description |
| ---- | ------ | ----------- |
| UserLoginFailed | 0 | 用户登录失败，身份验证未通过 |
| UserAlreadyExists | 1 | 用户已存在，无法创建用户 |
| UserNotFound | 2 | 用户未找到，可能是用户不存在或输入的用户标识有误 |
| UserCreateFailed | 3 | 创建用户失败，可能是由于服务器或其他问题导致的创建过程中的错误 |
| UserOperationForbidden | 4 | 用户操作被禁止，可能是由于权限不足或其他安全限制导致的 |
| SecretReachMaxCount | 5 | 密钥达到最大数量限制，无法继续创建新密钥 |
| SecretNotFound | 6 | 密钥未找到，可能是由于密钥不存在或输入的密钥标识有误 |
| SecretCreateFailed | 7 | 创建密钥失败，可能是由于服务器或其他问题导致的创建过程中的错误 |


 

 

 



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

