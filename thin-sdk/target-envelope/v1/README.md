# TargetEnvelope v1

> **已废弃**：该契约错误地把协议、方法和 endpoint 混入目标服务身份，并使用
> `x-pole-*` 键。新实现必须使用 Sidecar Session v1 与 TargetService v1；本目录
> 仅为已发布 `thin-sdk-contract-v1.0.0` 保留，不得继续扩展。

## 目标

Thin SDK 在语言框架已经识别逻辑目标后、连接本地 Pole Sidecar 前构造
TargetEnvelope。Sidecar 使用该信封执行服务级发现、路由和治理。

TargetEnvelope 解决“业务调用目标是谁”，不承载控制面规则，也不替代业务协议。

## 信任边界

- Sidecar 只能信任来自本地专用 listener 或 UDS 的 TargetEnvelope。
- SDK 校验不是安全边界；Sidecar 必须重复执行版本、来源和字段校验。
- Sidecar 转发请求前必须删除全部 `x-pole-*` 内部 Header。
- 协议原生身份与 TargetEnvelope 冲突时必须拒绝，或应用显式迁移策略；
  不得静默覆盖。

## 字段

| 字段 | 必填 | 含义 |
| --- | --- | --- |
| `namespace` | 是 | Pole 命名空间 |
| `service` | 是 | Pole 逻辑服务名 |
| `protocol` | 否 | `http`、`grpc`、`dubbo`、`thrift` 或扩展值 |
| `group` | 否 | 协议服务分组 |
| `service_version` | 否 | 目标服务版本 |
| `method` | 否 | RPC 方法或 HTTP 方法 |
| `original_endpoint` | 否 | 迁移阶段框架已经选择的原始 endpoint |

v1 的字段和 Header 集合冻结。任何字段增删、改名或语义变化都必须使用新的
envelope version。

## 规范化

1. 输入必须是 Unicode scalar value 序列，拒绝孤立 UTF-16 surrogate；随后检查
   Unicode General Category `Cc`，存在控制字符时拒绝。
2. 去除首尾 Unicode 15.1 `White_Space`：
   `U+0009..U+000D`、`U+0020`、`U+0085`、`U+00A0`、`U+1680`、
   `U+2000..U+200A`、`U+2028`、`U+2029`、`U+202F`、`U+205F`、
   `U+3000`。由于步骤 1 优先，属于 `Cc` 的空白字符仍然拒绝。
3. `U+FEFF` 不属于上述集合，不能作为空白去除。
4. `namespace` 和 `service` 规范化后不能为空。
5. 空的可选字段视为未提供，不生成 Header。

## original_endpoint

- 只接受非空 `host:port` 或 `[ipv6]:port`。
- port 使用无前导零的 canonical ASCII 十进制，数值范围为 `1..65535`。
- 非括号 host 不得包含冒号、`%`、Unicode `White_Space` 或
  `/`、`\`、`[`、`]`、`@`、`?`、`#`。
- 括号 host 必须符合 RFC 4291 IPv6 文本表示；接受 IPv4-embedded IPv6，
  不接受括号 IPv4 和 RFC 4007 zone identifier。
- host 保留调用方原值，不执行大小写、IDN、IPv4 或尾点规范化。

## Header 值编码

规范化字段必须先编码为 UTF-8，再逐字节转换成 ASCII Header value：

- `0x20..0x7E` 中除 `%`、`,` 外的字节原样保留；
- 其他字节以及 `%`、`,` 写成大写十六进制 `%HH`；
- Sidecar 只接受 ASCII `0x20..0x7E`、合法的 `%HH`、合法 UTF-8 和重新编码后
  完全相同的 canonical value。

因此 ASCII 常用值保持可读，Unicode、`%` 和 `,` 在不同 HTTP 库中仍有一致线路
表示。Sidecar 必须拒绝内部 Header value 中的原始逗号；即使 HTTP 栈已将重复
field occurrence 合并为逗号分隔值，也不会与合法字段值混淆。

## HTTP Header 编码

SDK 的有序编码结果按下列规范顺序生成；HTTP 在线传输和 map 类型不以 Header
顺序作为语义。

| 顺序 | Header | 字段 |
| --- | --- | --- |
| 1 | `x-pole-target-envelope-version` | 固定为 `1` |
| 2 | `x-pole-target-namespace` | `namespace` |
| 3 | `x-pole-target-service` | `service` |
| 4 | `x-pole-target-protocol` | `protocol` |
| 5 | `x-pole-target-group` | `group` |
| 6 | `x-pole-target-service-version` | `service_version` |
| 7 | `x-pole-target-method` | `method` |
| 8 | `x-pole-original-endpoint` | `original_endpoint` |

“v1 内部 Header”只表示表中的八个名称。SDK 合并调用方 Header 时，必须按
大小写不敏感方式删除这些名称的全部旧值，再写入一个规范名称和值。

Sidecar 必须在可访问原始 field occurrence 时先检查重复项，并要求 version、
namespace、service 恰好出现一次，可选内部 Header 至多出现一次；重复值、包含
原始逗号的内部值或未知的 `x-pole-target-*` Header 必须拒绝。转发到目标服务前，
Sidecar 必须删除所有 `x-pole-*` Header。

一致性向量中的 `expected_headers` 是测试夹具的有序表示：非内部 base Header
按输入顺序保留，随后追加上述规范顺序的内部 Header；线路接收方不得依赖顺序。

## Schema 与一致性

[schema.json](schema.json) 只用于规范化后的逻辑对象：空可选字段已经删除，
endpoint 已去除首尾空白。它固定字段结构和可静态表达的约束；构造器原始输入、
`original_endpoint` 的完整 IPv6 解析与 Header 编解码必须执行本规范和
[conformance.json](conformance.json)，不能只依赖通用 JSON Schema validator。

一致性文件的 `sidecar_receive` 部分定义 Sidecar 接收端行为；
`language_specific_invalid` 要求 SDK 使用语言原生方式构造 JSON 无法无损表达的
非法字符串。

本文与测试向量冲突时，视为 specification 契约缺陷，必须在本仓修复并提升契约
SemVer；SDK 不得自行选择其中一份解释。

## 兼容性

- v1 SDK 只生成 envelope version `1`。
- Sidecar 遇到未知 envelope version 必须拒绝。
- v1 字段与 Header 集合冻结，任何变化都提升 envelope version。
- 已验证组合只以 `thin-sdk/compatibility.json` 中的精确发行版本和证据为准。
