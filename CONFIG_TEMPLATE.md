# SDK 本地渲染与可选协议转换规范

本文定义配置中心 v1 的 `pole-mustache-v1` 引擎，配合 [统一配置协议](./CONFIG.md)
及 `ConfigDefinition`、`RenderSnapshot`、`ValueType` 使用。模板和 values 随一份 ConfigRelease 原子冻结。
原生 SDK 与控制台分别在本地执行运行时渲染和预览；服务端只在对应协议适配器和选中发布的
Apollo/Nacos/Consul 兼容转换都已显式开启时渲染。SDK、控制台和兼容适配器必须遵循同一引擎语义。
原生读取只分发模板和有效 values，不生成最终字节或预期输出摘要。
`literal` 是恒等字节模式，不适用以下模板语法，也不解释正文中的占位符。

## 参数与取值

`Parameter(name, type, required)` 只声明约束，默认值唯一来源是 `ConfigDefinition.default_values`。
参数名必须唯一，type 必须是 STRING、BOOLEAN、INTEGER、DECIMAL 对应的 `ValueType`；
UNSPECIFIED 和未知枚举值均非法。任何 default_values 或 rule.values 的 key 都必须事先声明，
Scalar.type 必须与 Parameter.type 完全一致，不进行隐式类型转换。

按请求 labels 匹配 ValueRule，选择唯一最小 priority，将该规则的 values 覆盖到 default_values 上。
不叠加其他命中规则；最高优先级并列拒绝。客户端标签仅用于选择，不自动成为模板变量。
required=true 的参数必须在最终 values 中存在；非必填且未被模板引用的参数可以缺失，
任何被占位符引用的参数都必须有值。空字符串是显式值，不能当作缺失。
服务端保存、预览及发布只对完整定义进行静态校验：默认分支及每条单规则覆盖分支的
语法、声明、类型、缺失值和输入上限均需满足约束，不能依赖“某个标签永远不会出现”绕过校验。
静态校验不执行占位符替换，也不检查最终输出文档格式；不通过发布前预渲染所有分支实现校验。

## 模板语法

模板只能使用 triple-mustache 原始变量，例如 `{{{database.host}}}`。
参数名是由 `.` 分隔的一个或多个非空段，每段匹配 `[A-Za-z_][A-Za-z0-9_-]*`；
占位符内部不接受额外空白，名称查找精确且区分大小写。

引擎保留模板的全部 UTF-8 字节，包括 BOM、空格和 CRLF/LF；只将占位符替换为标量的规范文本。
未声明的占位符、缺失值、重复参数、重复 JSON key、类型不符和非法标量必须报错。

不支持 double-mustache、HTML escape、section、循环、partial、继承、lambda、helper、
函数、pipeline、delimiter change、递归或动态模板引用；不读取文件、网络、环境变量、时间或随机数。
任何不属于合法占位符的 `{{` 或 `}}` 都是语法错误。

## 标量的规范文本

所有 Scalar.value 都通过字符串传输，禁止先经浮点数转换再还原。
JSON 中 type 使用 `VALUE_TYPE_STRING`、`VALUE_TYPE_BOOLEAN`、`VALUE_TYPE_INTEGER`、
`VALUE_TYPE_DECIMAL`。输出规则如下：

| 类型 | 规范值及输出 |
| --- | --- |
| STRING | 有效 UTF-8 字符串，逐字节原样输出，允许空值 |
| BOOLEAN | 仅 `true` 或 `false` |
| INTEGER | 匹配 `-?(0\|[1-9][0-9]*)`，拒绝 `-0`；不限制为 JavaScript Number 或 int64 |
| DECIMAL | 匹配 `-?(0\|[1-9][0-9]*)(\.[0-9]*[1-9])?`，拒绝 `-0` |

以上正则中的 `|` 表示分支。整数和小数的长度受统一协议的输入总量上限约束。
拒绝前导 `+`、前导零、科学计数法和小数尾随零，例如 `1.25` 有效，`1.250`、`01.25`、
`.25`、`1e3` 无效。模板引擎不额外加引号或进行 JSON/YAML 转义，作者须提供符合目标格式的正文和值。

## 输出、失败与测试向量

模板正文及字符串值必须为 UTF-8。引擎不得执行 Unicode、换行或空白归一化，
本地内容摘要由实际渲染器按最终字节计算，原生协议不要求服务端预先渲染出预期摘要。
模板结果为 UTF-8；literal 可以是任意二进制字节，均不转码。模板模式的 format 支持 text、json、yaml/yml、xml、
properties、html，ASCII 大小写不敏感，空值按 text 校验。text 不增加结构校验；其余格式校验
渲染后文档，其中 XML 要求恰好一个根元素。SDK、控制台预览和兼容适配器必须使用相同的格式校验策略。
服务端静态校验拒绝未知模板格式；实际输出格式校验在渲染方完成。
literal 的 format 仅为表示元数据，不改变任意原始字节的交付。

静态校验失败不能保存或发布成功；成功发布不保证模板替换后的文档有效。
SDK 必须检查引擎、参数、输入和输出上限，校验渲染后的文档；全部成功后才原子安装内容与接纳游标。
未知引擎、渲染或输出格式失败时保留同上下文允许的上一成功值，不推进版本/游标，不回落另一发布。
兼容适配器渲染失败使用公共 Code.ConfigRenderFailed 并映射外部协议错误；原生读取不产生这个渲染错误。
预览使用 Response(data=RenderSnapshot)，由控制台本地渲染，不产生运行时发布身份或观察游标。

[共享测试向量](./CONFIG_TEMPLATE_TEST_VECTORS.json) 的每项 request 是 PreviewRequest 的 protobuf JSON：
content 为 base64、type 为完整枚举名，默认值只出现在 definition.defaultValues。
valid=true 时 expectedContent 为预期精确字节的 base64，expectedSha256 为其摘要；
valid=false 时 reason 说明必须拒绝的原因，不增加一套运行时错误码。
SDK、控制台渲染器和已启用的兼容适配器应运行这些向量；先按 request 的定义和标签取得有效输入，再渲染。
spec 仓库只校验向量结构、消息编解码与摘要，不提供渲染器实现，也不要求 Preview HTTP 返回渲染结果。
