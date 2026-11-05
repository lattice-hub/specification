# 配置中心统一协议 v1

权威消息定义为 [config.proto](api/v1/config/config.proto)，package 为 `pole.config.v1`。
模板、values 和规则一起发布，原生 SDK 本地渲染。控制面只管理定义、发布、生效关系，
并按客户端上下文分发一致快照；保存、发布和原生读取不执行实际渲染。
本版直接归入 v1，不维护旧 Pole 分拆模板/Value/绑定发布或旧 SDK 的兼容协议。

## 消息分层与公共响应

业务对象与响应封装分开：HTTP 管理和原生客户端统一使用
[Response](api/v1/model/response.proto) 的 `code + info + Any data`，
错误使用 [Code](api/v1/model/code.proto)。不再定义配置专用 Error、ErrorCode、
StateResponse、PreviewResponse 或把错误混入业务存在状态。

| 业务数据 | 职责 |
| --- | --- |
| ConfigKey | environment/group/name 稳定定位 |
| ConfigDefinition / ConfigDraft | 完整定义及可编辑草稿版本 |
| ConfigRelease / ReleaseSummary | 不可变发布及当前生效状态摘要 |
| ConfigState | 管理查询投影、历史分页及指定版本详情 |
| RenderSnapshot | 本地渲染所需的原文、引擎、格式、参数和有效 values |
| ConfigResult | 成功运行时读取的存在状态、发布身份、快照和观察游标 |

普通列表查询复用 BatchQueryResponse；真正的批量写入复用 BatchWriteResponse。
当前 GetState 是带草稿、生效 ID 和历史页的聚合读取，因此使用
`Response(data=ConfigState)`。next_page_token 属于这个聚合数据，
不伪装成 BatchQueryResponse 已有的 amount/size，也不因此再造响应外壳。
当前协议没有新增独立的配置/历史列表接口。

Response.code=ExecuteSuccess 才允许消费 data，失败时 data 缺省，info 不泄露配置值。
Any 使用标准 type URL，例如 `type.googleapis.com/pole.config.v1.ConfigResult`。
JSON 按标准 protobuf JSON 编解码，data 带 `@type`，不能把 Any 当作无类型 JSON 对象。

## 定义与统一发布

ConfigKey 的三个字段必填；服务端派生 canonical identity，不支持重命名。
ConfigState.incarnation 标识存在期，下线保留草稿和历史，删除重建不得复用旧存在期。

ConfigDefinition 冻结 content、engine、format、parameters、default_values、value_rules
及 compatibility。完整定义随 ConfigRelease 原子发布，禁止从其他发布动态拼接 values。

- `literal` 为恒等字节模式，保留空内容、非 UTF-8、BOM 和 CRLF；参数、values、规则必须为空。
- `pole-mustache-v1` 遵守 [本地渲染规范](CONFIG_TEMPLATE.md)，标签不隐式成为模板变量。
- Parameter 只有 name/type/required，默认值唯一来自 default_values。
- Parameter.type 与 Scalar.type 共用 ValueType；零值和未知值拒绝。
  标量使用字符串无损传输，不能经过浮点数再还原。
- Audience 用于灰度和 ValueRule。旧 `labels` 保持精确相等 AND；新 `conditions` 按 key 唯一的条件 AND；两者不得混用。缺少标签时任何条件（包括负条件）均不匹配，`*` 为普通字面字符。空选择器只用于 normal/default，不能用于 gray 或 ValueRule；标签不替代授权。
- `EXACT`/`NOT_EQUALS` 按字符串字面相等/不相等；`REGEX` 使用 Go RE2 搜索语义；`IN`/`NOT_IN` 以逗号分隔字面成员，不隐式去空格，空成员表示空字符串；`RANGE` 为 `left~right` 有符号 int64 闭区间，包含端点。
- 新条件最多 64 条，每条 key 非空且不超过 128 字节，value 不超过 4096 字节；零值或未知 type、重复 key、无效正则或无效/倒序范围均拒绝。旧 labels 和请求 labels 的原有限制不变。
- 默认 values 加上唯一最小 priority 的命中 ValueRule，仅覆盖其声明字段，不叠加其他规则。
  最高优先级并列拒绝；required 参数和占位符引用必须有值。
- 所有 priority 为非负 int32；normal 的 priority=0、audience 为空。

保存/发布进行大小、语法、参数声明、标量类型、规则歧义和输入完整性的静态校验。
默认分支及每条单规则覆盖分支均须满足参数约束，但不执行模板替换、不生成最终字节，
也不计算预期输出摘要。渲染后格式校验由 SDK/控制台或已启用的兼容适配器执行。
发布成功不代表所有客户端已渲染或安装成功。

最多一个 active normal，允许多个 active gray，也允许尚无 normal。
命中灰度按 priority ASC、sequence DESC 唯一选择；sequence 在该配置环境内单调分配。
没有灰度命中才回落 normal，没有 normal 则返回 ABSENT。
选中发布的取值、引擎或渲染失败不能转投次选发布。
灰度选择同优先级新发布优先，与取值规则同优先级歧义拒绝，是两个不同作用域。
回滚产生新 sequence，可能改变重叠灰度的流量归属；其他版本仍然保留。

## 管理 HTTP 动作

所有 expected_draft_revision / expected_release_set_revision 为精确 CAS，
零表示初始草稿/无发布，不能表示跳过比较。事务必须原子复核版本、固定发布定义、
更新生效关系与持久观察位置，失败整体回滚。

| HTTP 路径 | 请求 | 作用 | Response.data |
| --- | --- | --- | --- |
| /config/v1/files | ListConfigFilesRequest | 按环境查询配置目录 | ConfigFileList |
| /config/v1/file-template | GetFileTemplateRequest | 查询共享正文和环境 | ConfigFileTemplate |
| /config/v1/file-template-draft | SaveFileTemplateRequest | 精确 CAS 保存共享正文 | ConfigFileTemplate |
| /config/v1/draft | SaveDraftRequest | 保存整份定义，仅推进草稿版本 | ConfigState |
| /config/v1/state | GetStateRequest | 草稿、历史摘要页、生效 ID 和可选版本正文 | ConfigState |
| /config/v1/preview | PreviewRequest | 静态校验并按 labels 选值，控制台本地渲染 | RenderSnapshot |
| /config/v1/publish | PublishRequest | 冻结草稿并发布 | ConfigState |
| /config/v1/promote-gray | PromoteGrayRequest | 原子将指定生效灰度转正式并停止源灰度 | ConfigState |
| /config/v1/stop | StopRequest | 停止指定 active gray | ConfigState |
| /config/v1/stop-all-gray | StopAllGrayRequest | 停止该配置全部 active gray，保留 normal | ConfigState |
| /config/v1/copy-release-to-draft | CopyReleaseToDraftRequest | 指定历史完整定义复制到草稿，复核两项 CAS | ConfigState |
| /config/v1/rollback | RollbackRequest | 复制历史定义和投放条件，产生新发布 | ConfigState |
| /config/v1/offline | OfflineRequest | 停止该配置全部 active 发布，保留草稿和历史 | ConfigState |

normal 发布只替换 normal；gray 默认追加。replace_release_id 只替换指定同 kind active 发布。
Publish.stop_release_id 仅允许 normal 发布原子停止明确指定的 active gray，
不能隐式停止全部灰度。复制历史到草稿不停止或激活版本，需要另行发布。
Rollback 的 source.kind 决定发布类型，受众和优先级也来自 source；
gray 必须指定同 kind active 替换目标，normal 无替换目标时只允许当前无 active normal。
复制和回滚也复制该历史版本的 compatibility，管理界面须展示对应协议开关后再发布。

Preview 返回 RenderSnapshot，没有 release_id、snapshot_revision 或 observation_cursor，
不能安装为 SDK 生效结果。控制台在本地执行模板替换及输出格式校验，
不通过“预览”接口绕过服务端渲染的开关限制。

### 管理分页与资源上限

ConfigState 仅是对外管理投影，不能用作不完整的内部存储聚合。
ConfigState.draft 包含完整草稿，releases 为 ReleaseSummary 摘要页，
selected_release 仅在 GetStateRequest.release_id 指定时携带单份完整历史定义。
active_release_ids 是全部 active normal/gray ID，不受分页截断影响。
写操作返回默认第一页，不附带 selected_release。

历史摘要按 sequence DESC；page_size=0 默认20，最大100，
单页摘要 protobuf 编码最多256 KiB，允许提前截页；next_page_token 指向最后实际返回的 sequence。
token 绑定 key/incarnation/release_set_revision/before_sequence，
同一读取快照校验；存在期或发布集合变化返回 DataConflict，从首页重查。
纯草稿变化不使历史 token 失效，每页草稿仍来自该次一致快照。
release_id 只查当前已授权 key 的历史，缺失返回 NotFoundResource。
分页 token 不替代运行时游标、CAS 或权限。

| 项目 | 上限 |
| --- | --- |
| key 三项 | 各为非空 UTF-8，最多128 bytes，无前后空白、NUL、CR、LF |
| labels / audience | 最多64项；key非空最多128 bytes，value最多1024 bytes，UTF-8 且无 NUL/CR/LF |
| 完整 ConfigDefinition | protobuf 编码最多1 MiB，包含全部规则和输入 |
| 原文及本地最终内容 | 各最多512 KiB；渲染器必须在分配膨胀前检查输出上界 |
| parameters / default_values / value_rules | 各最多128项，每条 rule.values 最多128项 |
| 每个选值分支的输入 | 原文、参数名、值总量最多512 KiB |
| active 发布 | 最多1 normal + 128 gray |
| HTTP 请求 | 整个 JSON 文档最多4 MiB |

这些是原生协议上限，不代表外部协议的兼容 profile 已验收。
完整草稿和单份 selected_release 各自最多包含一份定义，历史分页不累积正文。

## 原生 Discover

管理只走 HTTP。客户端 gRPC 保留唯一 `v1.ConfigGRPC/Discover`：
`stream DiscoverRequest -> stream v1.Response`。
每帧重新鉴权，按 key/labels 选择发布和有效 values，在同一快照中返回
`Response(data=ConfigResult)`；服务端不执行模板替换。
没有请求就不发送，按请求顺序一请求一完整响应，版本相同也返回完整快照。
SDK 通过在 Discover 流上继续发送请求检查变化，自行决定轮询间隔和退避。
服务端收到请求即读取当前快照，不等待后续变化；不注册订阅，不主动 push。
EOF 正常结束，传输错误关闭流；业务失败返回公共 code/info 后允许后续请求继续。

原生客户端读取统一使用 Discover；不提供独立 Resolve/Watch 消息、RPC 或 HTTP 读取入口。
网络超时和流关闭不表示配置删除，只有成功返回 ConfigPresence.ABSENT 才表示缺失。

RenderSnapshot.content 是模板原文或 literal 字节；engine、format、parameters 和有效 values
足够 SDK 离线完成渲染。不下发其他受众的 rules/values、历史、草稿或兼容开关。
SDK 不向配置中心回传本地渲染正文，服务器也不为原生响应计算最终 content_sha256。
原生消费权限必然允许读取选中模板及有效 values，但不授予完整定义管理权限；
秘密不能靠模板占位符隐藏，也不能靠客户端自报标签实现权限隔离。

## 存在状态、版本与 SDK 接纳

ConfigPresence 只包含 UNSPECIFIED/PRESENT/ABSENT；UNSPECIFIED 不能成功接纳。

| 成功业务状态 | 数据及行为 |
| --- | --- |
| PRESENT | 必须含 RenderSnapshot、release_id、snapshot_revision 和 observation_cursor；空原文仍为有效存在 |
| ABSENT | 无 snapshot/release_id；带缺失状态版本和游标，撤下当前配置 |

snapshot_revision 是 opaque 标识，绑定 key、存在期、选中发布、
模板、引擎、参数、有效 values 和输出格式，不可排序，不是最终字节摘要。
不包含观察游标或不相关灰度内容，无关灰度变化不应改变当前快照标识。

Cursor(scope, epoch, position) 表示完整观察的持久提交位置，仅同 scope/epoch 可比较。
结果与游标来自同一读取快照。min_cursor 是下限，落后实例补齐或返回 ConfigUnavailable；
世代变化或游标不可恢复返回 ConfigCursorReset，不得静默降低下限。

SDK 收到 PRESENT 后先验证引擎版本、参数和输入边界，再本地渲染并校验输出格式，
全部成功后才原子安装内容、snapshot_revision 和接纳游标。
不支持引擎、渲染或格式校验失败时不推进版本/游标，不切换其他发布；
仅允许按相同身份、配置、标签上下文的既定策略使用上一成功值。
重复快照可复用已成功的本地结果，更新该一致读取的观察下限。
ABSENT 成功接纳时撤下旧值；明确鉴权拒绝时清除该上下文可用缓存。
ConfigCursorReset 使旧读取流程失效，重新读取建立基线。

本地 context_generation/request_seq 淘汰过期请求和应用回调，
标签 A→B→A、停止灰度后迟到响应、删除重建、旧拒绝/缺失响应不得覆盖新状态。
min_cursor 只携带已成功接纳的游标；快照收到了不代表本地配置已生效。

## 公共错误码

| 公共 Code | HTTP | 语义 |
| --- | --- | --- |
| ExecuteSuccess | 200 | data 为对应业务消息；ABSENT 也是成功 |
| InvalidParameter | 400 | 非法请求、静态校验或选值失败 |
| DataConflict | 409 | CAS 或历史页状态冲突 |
| NotFoundResource | 404 | 管理查询目标不存在 |
| Unauthorized / NotAllowedAccess | 401 | 认证或访问身份拒绝 |
| AuthTokenForbidden / OperationRoleForbidden | 403 | 权限拒绝 |
| ConfigUnavailable | 503 | 无法满足一致读取、依赖故障 |
| ConfigCursorReset | 409 | 读取世代失效，需要重建基线 |
| ClientAPINotOpen | 400 | 对应配置协议转换未开启 |
| ConfigRenderFailed | 422 | 仅已启用兼容适配器的服务端渲染失败 |

info 不承载敏感正文；错误时 data 缺省，不附上旧快照伪装成功。
SDK 本地渲染错误是本地消费状态，不伪造服务器返回码。
兼容适配器将公共失败映射为外部协议自己的错误，不能将失败映射为空内容或删除。

## 可选 Apollo / Nacos / Consul 转换

ConfigDefinition.compatibility 分别提供 apollo/nacos/consul 布尔开关，缺省全部 false。
开关与定义一起冻结，按选中发布判断；发布变更由正常 CAS/审计追踪。

兼容请求必须同时满足：
1. 部署显式启用对应协议适配器。
2. 已授权请求选中的发布显式开启该协议转换。

不满足时不执行渲染，拒绝对应兼容请求，不回落到另一个已开启兼容的 normal/gray。
配置下线或无生效版本时返回外部协议的明确缺失语义，不渲染；
关闭适配器时禁用入口，关闭发布的转换能力时不能继续从旧渲染缓存成功交付。

满足条件后，适配层取得同一选版/选值快照，使用与 SDK 一致的引擎语义生成最终内容，
再映射外部协议的版本、校验和及等待语义。兼容开关不改变原生链路和管理 Preview。
不要在发布时预渲染所有客户端标签组合，也不要启动无条件的全量渲染任务。
缓存至少按配置身份/存在期、snapshot_revision、协议及输出表示隔离；
每次命中前仍须重新授权并检查当前选中发布的开关，缓存不能替代关闭、下线或权限变化。

本合同限定渲染边界，不代表三个适配器的标签来源、认证、错误及版本映射已实现或通过兼容验收。

## 生成与验收边界

Go 包：`source/go/api/v1/config`；客户端服务包：`source/go/api/v1/config_manage`。
Rust 消息：`pole::config::v1`；客户端：`v1::config_grpc_client::ConfigGrpcClient`。
Java 消息类：`io.github.latticehub.pole.specification.api.v1.config.ConfigProto`；
服务描述符：`io.github.latticehub.pole.specification.api.v1.config.manage.PoleConfigGRPCService`。

```sh
bash source/go/build.sh config
bash source/rust/build.sh config
go test ./...
cargo fmt --all -- --check
PROTOC="$PWD/source/protoc/protoc-darwin-arm64/bin/protoc" cargo test --all
PROTOC="$PWD/source/protoc/protoc-darwin-arm64/bin/protoc" cargo check --all
```

其他平台按 AGENTS.md 选择内置 protoc。Go/Rust 共享 wire 向量验证二进制原文、
uint64、无损标量、公共 Response/Any、分页和默认关闭的兼容开关。
这些是协议与生成包检查，不代替 control-plane、原生 SDK、控制台本地预览或兼容适配器实现验收。

### 按参数绑定的条件组

`ConfigDefinition.condition_groups` 和 `parameter_bindings` 是可选的新选值模型，任一非空时不得同时设置 `default_values` 或 `value_rules`。两个字段均为空时保留原有单条 `ValueRule` 胜出的行为。条件组的 ID 和名称在定义内唯一，ID 使用 ASCII 字母、数字、`_`、`.`、`-`；每组有非空 `Audience`。每个参数最多一个 `ParameterBindings`，绑定按数组顺序检查，取第一个命中组的值；没有命中时使用有 presence 的 `fallback`，未设置则不返回该参数。空字符串是有效兜底。被正文引用或声明为 required 的参数必须提供兜底，其他参数可不提供；绑定值与兜底必须符合参数类型。条件组和绑定随发布完整冻结；Preview 与 Discover 使用相同的选值语义，客户端仍只接收已选出的 `RenderSnapshot`。

`ConfigFileTemplate` 表示共享正文和环境目录，其 `definition` 只允许正文、引擎、格式、参数声明和兼容开关，不含任一值模型。各环境草稿和发布仍独立，revision 以及发布 CAS 不跨环境组成事务。
