# 配置模板规范

本文定义配置中心跨语言模板协议 `pole-mustache-v1`。其中的 MUST、MUST NOT、
SHOULD 按 RFC 2119 的规范性含义解释。

所有服务端参考实现和 SDK MUST 运行仓库根目录
[`CONFIG_TEMPLATE_TEST_VECTORS.json`](./CONFIG_TEMPLATE_TEST_VECTORS.json) 中的共享向量。

## 资源与运行时职责

- `ConfigFile.config_type = CONFIG_FILE` 时，`content` 是客户端直接使用的普通文本。
- `ConfigFile.config_type = CONFIG_TEMPLATE` 时，配置文件必须携带
  `ConfigTemplateBinding`，并显式固定 `template_release_id`。
- 模板和 Namespace Value 分别发布为不可变的 `ConfigTemplateRelease` 与
  `NamespaceTemplateValueRelease`。
- Value 聚合身份是 `namespace + template_id`。服务端根据客户端标签只返回唯一命中的
  Value Release，SDK MUST NOT 解释灰度规则。
- 服务端通过 `RenderSnapshot` 原子下发模板、命中的 Value 和组合 revision。SDK MUST
  在本地渲染并校验哈希，成功后才能原子替换 last-known-good。
- `PreviewConfigTemplate` 的结果只用于预览、发布校验和参考哈希，MUST NOT 作为客户端
  运行时权威配置。
- `RenderPreview.code/info` 使用统一 API 错误码表达鉴权、请求参数和系统错误；
  `diagnostics` 只表达模板语法、Value、类型和目标格式诊断，MUST NOT 承载权限错误。

## `pole-mustache-v1`

引擎能力标识固定为：

```text
name = pole-mustache
version = v1
```

### 语法

模板只允许 triple-mustache 原始变量：

```text
{{{database.host}}}
{{{database.port}}}
```

参数名由一个或多个 `.` 分隔的非空段组成。每段 MUST 匹配：

```text
[A-Za-z_][A-Za-z0-9_-]*
```

实现 MUST：

- 对参数名进行精确、区分大小写的查找；
- 保留模板的全部原始字节，包括空白、换行和 UTF-8 字符；
- 将每个占位符替换为对应标量的规范文本；
- 拒绝缺失的必填参数、未在 Schema 声明的占位符和 Value；
- 在缺失非必填参数时使用 Schema 的 `default_value`；没有默认值仍视为错误；
- 拒绝重复 Schema 参数名、重复 map key 和类型不匹配。

实现 MUST NOT 支持或隐式解释：

- double-mustache、HTML escape；
- section、inverted section、循环；
- partial、继承、lambda、helper、函数、pipeline 或方法调用；
- delimiter change、递归和动态模板引用；
- 文件、网络、环境变量、系统时间或随机数。

模板中出现任何不属于合法 triple-mustache 的 `{{` 或 `}}` MUST 报语法错误。

### 标量序列化

| 类型 | 传输字段 | 输出规则 |
|---|---|---|
| string | `string_value` | UTF-8 字符串原样输出 |
| boolean | `boolean_value` | 仅输出 `true` 或 `false` |
| integer | `integer_value` | 十进制输出，不带前导 `+` 或前导零 |
| decimal | `decimal_value` | 规范十进制字符串，不允许科学计数法 |

decimal MUST 匹配：

```text
-?(0|[1-9][0-9]*)(\.[0-9]*[1-9])?
```

此外 `-0` MUST 被拒绝。该规则不允许尾随小数零；例如 `1.25` 合法，`1.250`、
`01.25`、`.25` 和 `1e3` 非法。

### 输出与哈希

- 输入模板和 string Value MUST 是有效 UTF-8。
- 实现 MUST NOT 执行 Unicode、换行或空白归一化。
- `rendered_sha256` 是最终 UTF-8 字节的 SHA-256 小写十六进制值。
- SDK MUST 校验渲染后内容的目标 `format`，并将本地哈希与
  `expected_rendered_sha256` 比较。
- 格式校验、哈希或渲染失败时，SDK MUST 拒绝新快照并保留 last-known-good。

示例：

```yaml
# template
host: {{{database.host}}}
port: {{{database.port}}}
enabled: {{{feature.enabled}}}
ratio: {{{traffic.ratio}}}
```

使用 `数据库.local`、`3306`、`true`、`1.25` 渲染后的精确内容是：

```text
host: 数据库.local
port: 3306
enabled: true
ratio: 1.25
```

末尾包含一个 LF，其 SHA-256 为：

```text
4f35ba156d46d0236a0dfdad17e65692e84d5376c444154ddca9c18fe617ca10
```

## 能力协商

客户端通过 `ConfigDiscoverFilter.supported_template_engines` 声明其支持的引擎和版本。
服务端仅能向声明支持目标版本的客户端返回模板快照。旧客户端或不支持目标版本的客户端
必须收到明确的不兼容错误；服务端 MUST NOT 把未渲染模板作为普通配置静默下发。

## 组合 Revision

`RenderSnapshot.revision` MUST 至少覆盖以下身份：

```text
template_binding.binding_release_id
template_release.id
value_release.id
template_release.engine.name
template_release.engine.version
```

revision 的具体哈希编码由服务端版本化实现，但对相同输入 MUST 稳定；上述任一输入变化
都 MUST 产生新的客户端可见 revision。
