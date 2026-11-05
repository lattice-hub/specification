---
type: Task Record
---

# 2026-08-04 限流运行时目标服务与 ALPHA.45

## 原生 SDK 渲染与公共响应收敛（2026-10-03）

- [x] 核对公共 Response/Code、spec 与控制面调用链，保留共享工作区已有修改。
- [x] 原生 Discover 分发 RenderSnapshot；明确定义 SDK 本地渲染及失败接纳规则。
- [x] 删除配置专用响应/错误体系，复用 Response、BatchQueryResponse 和公共 Code。
- [x] 定义 Apollo/Nacos/Consul 兼容转换显式开关；保存、发布、原生读取和服务端预览不执行渲染。
- [x] 同步规范、Go/Rust 生成代码和跨语言契约测试，执行检查并记录结果。

### 本轮范围与纠正规则

- 修改范围为 specification 的协议、生成产物、文档和合同测试。control-plane 当前仍引用本轮前已从 spec 删除的 ReleaseMember/PromoteRequest；运行时迁移、SDK 渲染实现和三个兼容适配器接入单独验收，不以协议通过代表实现完成。
- 统一发布只决定模板、values 和规则的版本原子性，不决定渲染位置。原生 SDK 本地渲染，控制台本地预览；服务端仅在用户显式开启兼容转换后渲染对应协议请求。
- 业务快照属于 Response.data，不因此另建响应和错误码体系。不提交、推送或部署。

### 本轮 Review

- 原生 Discover 保持逐请求/逐响应，统一使用 v1.Response(data=ConfigResult)。RenderSnapshot 只含原文、引擎、格式、参数和当前有效 values，SDK 成功渲染后才接纳快照版本与游标；不携带服务端预渲染摘要。
- 删除配置私有 Error/ErrorCode、StateResponse、PreviewResponse 和最终内容专用 ResolvedConfig。分页字段直接并入管理投影 ConfigState，Response.data 直接承载该状态；普通列表/批量写入沿用已有公共响应，不增加对应新接口。
- ConfigDefinition.compatibility 的 apollo/nacos/consul 缺省 false，与发布冻结。明确部署适配器和选中发布双条件、关闭后不复用渲染缓存、不回落其他已开启的发布；原生入口和控制台预览始终不触发服务端渲染。
- 公共 Code 仅补 ConfigCursorReset、ConfigRenderFailed、ConfigUnavailable。Go/Rust 生成代码和 wire/Any/默认关闭合同同步；Rust 公共 Response 生成位置因依赖顺序移动，其字段未变，其他模块并行修改保留。
- 验证通过：go test ./...、go vet ./...、cargo test --all、cargo check --all、cargo fmt --all -- --check、Node 生成/构建/测试、Python 生成及3项测试、定向脚本语法、Rust Proto 副本/文档链接检查、git diff --check。
- 本轮未修改 control-plane、SDK 或实际兼容转换实现；没有执行运行时行为验收、提交、推送或部署。协议开关的默认值/编解码已验证，不等于对应适配器已按开关执行。

## TrafficContext v1 传播契约（2026-08-06）

- [x] 核对 TargetService v1、Sidecar TrafficContext 与 OTel 传播边界
- [x] 定义 W3C Baggage 保留键、字段约束和失败语义
- [x] 增加 Schema、conformance vectors 与校验和
- [x] 更新 Thin SDK 契约索引与兼容性声明
- [x] 运行格式、Go 与 Rust 验证并记录 Review

### TrafficContext v1 Review

- 冻结 `latticehub-traffic-context` 1.0.0：W3C Baggage 保留 version/campaign/lane/bucket，TargetService 与 Trace wire 契约保持独立。
- JSON Schema 通过 `x-latticehub-maxUtf8Bytes` 声明 128-byte 约束，conformance 同时提供 43 个中文字符的超限拒绝向量，避免按字符数误判。
- 有效向量覆盖完整标签、Unicode/保留字符和仅 lane 的部分标签；接收端向量覆盖缺 version、重复、未知 version、bucket 越界、非 canonical 与 UTF-8 byte 越界。
- 双轴审查发现的 Schema 字节语义、可选字段、空 carrier 与 W3C 180 成员边界均已修复。
- 运行时语义允许语言习惯的 closeable `attach` 或 callback `runWith`；Node.js 标准 OTel
  bridge 明确使用 `context.with`，禁止把 SDK 私有存储冒充为 OTel active Context。
- Java、Go、Node.js、Python、C++、C# 已 vendoring 四项资产并执行全部 TrafficContext
  valid 与 Sidecar receive valid/invalid 向量；精确发布来源仍需本次 Specification 合入后写入
  各仓 `contract/VERSION`，当前不得伪造未提交内容的 commit SHA。
- `go test ./...`、`cargo fmt --all -- --check`、`cargo test --all`、`cargo check --all` 与 `git diff --check` 全部通过。

- [x] 为 `RateLimit` 增加正式 `callee` 作用域
- [x] 重新生成并验证多语言协议产物
- [ ] 通过 PR 合入 `develop`
- [ ] 发布 `v0.1.0-ALPHA.45`
- [ ] 验证 control-plane 可按 tag 消费

## Review

- 待完成。

# 仓库发布流程固化与 `.38` Release

- [x] 在仓库根目录新增本地 `AGENTS.md` 发布规范
- [x] 记录“tag 后必须走 GitHub Release”的纠正经验
- [x] 审查、提交并通过 PR 合入 `develop`
- [x] 发布 `v0.1.0-ALPHA.38` GitHub Release
- [x] 等待并验证 Rust 发布流水线终态

## Review

- 仓库级规范通过 PR `#9` 合入 `develop`，合并提交为 `a83ece1`。
- 规范只新增在仓库根目录 `AGENTS.md`，没有修改全局规则。
- Standards 与 Spec 双轴审查无剩余阻塞，PR 和合并后 CI 全部通过。
- `v0.1.0-ALPHA.38` GitHub prerelease 已发布。
- `Release-Rust` 工作流 `30192552431` 成功，`cargo publish` 步骤通过。
- 已通过 `cargo info pole-specification@0.1.0-ALPHA.38` 从 crates.io
  下载并确认版本。
- 发布工作流仍有旧版 Actions、`set-output` 和 `cargo login <token>` 的弃用告警；
  不影响本次发布，后续应单独升级。

## Thin SDK TargetEnvelope v1 契约（2026-07-31）

- [x] 核对现有兼容性和共享向量模式
- [x] 定义 TargetEnvelope v1 语义与信任边界
- [x] 增加 JSON Schema 和机器一致性向量
- [x] 增加 SDK/Sidecar 兼容矩阵
- [x] 增加 Go 资产自检测试
- [x] 运行 Go、Rust 和格式验证
- [x] 完成双轴代码审查
- [x] 修复审查发现并重新验证
- [ ] 通过 PR 合入 develop

## Review

- 首轮双轴审查发现 Schema 与语义不一致、v1 演进规则冲突、Unicode Header
  缺少线路编码、endpoint 与重复 Header 规则不完整。
- 已暂停四语言 SDK 接入，先修正 specification，避免实现分叉。
- 修正后冻结 v1 字段集合，定义 Unicode scalar 与 ASCII 百分号线路编码，
  增加 Sidecar 接收端拒绝向量、精确兼容组合及 Go 参考执行器。
- 最终复审确认端口 canonical 规则、解码结果断言、normalized Schema 和兼容矩阵
  Schema 已闭环，无剩余 P1/P2。
- `go test ./...`、`cargo fmt --all -- --check`、`cargo test --all`、
  `cargo check --all`、JSON Schema normalized vectors 与 SHA256 校验均通过。

## Java 生成包 Namespace 迁移（2026-08-02）

- [x] 将全部 Proto `java_package` 从 `io.pole.specification.*` 迁移到 `io.github.latticehub.pole.specification.*`。
- [x] 重新生成 Go descriptor，并同步 Rust Proto 输入。
- [x] 同步 Sidecar 与四语言 Thin SDK 的 bootstrap 契约副本。
- [x] 完成 Go、Rust、Java 与契约校验。

### Review

- Maven namespace 与 Java 源码 package 分离；Specification 生成类型统一迁移到 `io.github.latticehub.pole.specification.*`。
- Java Thin SDK 公共 API 使用 `io.github.latticehub.client`，不增加重复的 `pole` 层级。
- Java package option 不改变 Protobuf wire contract；Go/Rust descriptor 与各消费者 vendored bootstrap 已重新生成或同步。

## 2026-08-03 Node.js gRPC npm 产物发布核查

- [x] 核对 `develop` 与全部本地、远端分支
- [x] 扫描 Node.js 包元数据、生成目录和 npm workflow
- [x] 区分 Specification 生成包与 Node.js Thin SDK

### Review

- 当前 `specification` 只包含 Go 生成代码和 Rust crate；不存在 Node.js
  `package.json`、生成脚本、生成产物或 npm 发布 workflow。
- 因此当前没有可安全执行的 npm 首次发布命令；必须先定义包名、生成技术栈、
  版本来源和发布内容，再实现并验证发布链路。

## 2026-08-03 Node.js 与 Python gRPC 生成包自动化

- [x] 定义 npm 与 PyPI 包结构和版本映射
- [x] 实现 Node.js Proto、类型和加载入口生成
- [x] 实现 Python Message、类型和 gRPC Stub 生成
- [x] 增加分支与 Pull Request 持续验证
- [x] 增加 GitHub Release OIDC 发布工作流
- [x] 完成仓库级验证并推送

### Review

- npm 包 `@lattice-hub/pole-specification` 生成 26 个权威 Proto 的动态加载入口和
  TypeScript 类型；`npm test` 与 `npm pack --dry-run` 通过。
- PyPI 包 `pole-specification` 生成 Message、`.pyi`、Client Stub 与 Servicer；
  单元测试、sdist、wheel 构建及 wheel 独立导入通过。
- Python 生成器固定为 `grpcio-tools==1.80.0`，运行时下限同步为
  `grpcio>=1.80.0`、`protobuf>=6.31.1`，避免生成代码与声明依赖不一致。
- `actionlint` 对新增及修改的工作流检查通过；GitHub 已创建 `npm`、`pypi`
  Environments，发布使用 OIDC，无长期 Registry Token。
- 测试工作流的 Checkout 与 Go Action 已升级，消除 GitHub 托管 runner 的
  Node.js 20 弃用告警。
- Go 全量测试、Rust fmt/test/check 与 `git diff --check` 均通过。

## 2026-08-03 Specification v0.1.0-ALPHA.41 多语言发布

- [x] 核对最新 tag、Release 与制品发布入口
- [x] 现代化 Rust Release 工作流并验证
- [x] 提交并推送发布前修复
- [x] 创建 annotated tag 与 GitHub prerelease
- [x] 跟踪 Rust、Node.js、Python 发布工作流
- [x] 验证 Go tag、crates.io、npm、PyPI 可消费

### Review

- 发布前修复通过 PR `#14` 合入 `develop`，合并提交为 `18c78fb`；合并后
  Testing 工作流 `30768671079` 的 Go、Rust、Node.js、Python 全部通过。
- annotated tag 与 GitHub prerelease `v0.1.0-ALPHA.41` 已发布。
- Rust 工作流 `30768728739` 成功，crates.io 可下载
  `pole-specification 0.1.0-ALPHA.41`。
- Generated Packages 工作流 `30768728734` 成功；npm 可下载
  `@lattice-hub/pole-specification@0.1.0-ALPHA.41`，PyPI wheel 与 sdist
  `pole-specification 0.1.0a41` 均可下载且 wheel 导入通过。
- `GOWORK=off go list -m` 可解析
  `github.com/pole-io/specification@v0.1.0-ALPHA.41`。
- npm `alpha` 已指向本次版本；`latest` 从 bootstrap 切换到本次版本需要维护者
  在本机完成一次 2FA。后续 Release 直接使用 `npm publish` 自动更新 `latest`。

## 2026-08-03 Java Specification 生成包恢复

- [x] 追查 Java 工程删除历史与 Maven Central 状态
- [x] 明确新坐标和 Java package namespace
- [x] 恢复权威 Proto 驱动的 Maven 生成工程
- [x] 增加逐 Proto 与逐 gRPC Service 完整性测试
- [x] 接入 CI 与 Maven Central Release
- [x] 触发首个 Java 制品发布
- [x] 修复 npm 预发布 dist-tag 参数
- [x] 发布并验证统一的多语言版本

### Review

- PR `#16` 已恢复 Java 生成包和 Maven Central 发布链路，并以合并提交
  `9d4bfd5` 进入 `develop`。
- 合并后 Testing 工作流 `30769452097` 全部通过。
- `v0.1.0-ALPHA.42` 首次触发 Java 发布；Rust 和 PyPI 已成功，npm 11
  因预发布版本未显式指定 dist-tag 而拒绝发布，需修复后发布新的统一版本。
- npm 修复通过 PR `#17` 合入 `develop`，合并提交为 `d33ecaf`；合并后
  Testing 工作流 `30769797731` 的 Go、Rust、Java、Node.js 和 Python
  全部通过。
- annotated tag 和 GitHub prerelease `v0.1.0-ALPHA.43` 已发布；Java 工作流
  `30769845673`、Rust 工作流 `30769845695` 和 Generated Packages 工作流
  `30769845687` 全部成功。
- Maven Central 可下载
  `io.github.lattice-hub:pole-specification:0.1.0-ALPHA.43`，JAR 包含
  `io.github.latticehub.pole.specification.api.v1.sidecar.SidecarSessionServiceGrpc`。
- `GOWORK=off go list -m`、`cargo info`、npm 全新安装和 PyPI 全新虚拟环境安装
  均验证了 `ALPHA.43`；npm `latest` 已指向 `0.1.0-ALPHA.43`。

## 2026-08-03 C++ Specification 生成包

- [x] 核对现有多语言生成和发布边界
- [x] 定义 CMake 包与 GitHub Release 资产形式
- [x] 实现全量 Protobuf 与 gRPC C++ 生成
- [x] 增加完整性、编译与安装验证
- [x] 接入 Pull Request CI 与 Release 附件发布
- [x] 通过 PR 合入 `develop`

### Review

- C++ CMake 工程从权威 `api/` 生成 26 个 Proto 对应的 104 个 Message/gRPC
  源文件，不提交生成副本；Release 产出 `.tar.gz`、`.zip` 和 `SHA256SUMS`。
- 固定使用 gRPC 1.64.3、Protobuf 26.1 及其官方 Abseil revision，工具链下载
  校验 SHA-256，并由 GitHub Actions cache 复用。
- `RateLimitGRPC.Service` 仅在 C++ API 中映射为 `Stream`，自动校验生成描述符
  仍使用 `/polaris.metric.v2.RateLimitGRPC/Service`，且打包 Proto 与权威源逐字节一致。
- Ubuntu 24.04 ARM64 容器已通过全量编译、2 个 CTest、CMake 安装、独立消费者
  `find_package` 编译运行、压缩包校验及解压后二次构建。
- PR #19 的 Go、Rust、Node.js、Python、Java 与新增 C++ job 全部通过；C++ job
  在 Ubuntu x86_64 上完成固定工具链构建并保存 cache，本提交随 PR 合入 `develop`。

## 2026-08-04 Specification v0.1.0-ALPHA.44 多语言发布

- [x] 核验 ALPHA.43 Java、Node.js、Python 发布状态
- [x] 确认 `develop` 集成提交与发布前 CI
- [x] 创建 annotated tag 与 GitHub prerelease
- [x] 跟踪 Rust、Java、Node.js、Python、C++ 工作流
- [x] 验证各 Registry 可消费与 C++ Release 资产

### Review

- 发布准备通过 PR `#20` 合入 `develop`，合并提交为 `78a6d3e`；合并后 Testing
  工作流 `30831280406` 的 Go、Rust、Java、Node.js、Python 和 C++ 全部成功。
- annotated tag 和 GitHub prerelease `v0.1.0-ALPHA.44` 已发布，tag 指向已验证的
  `78a6d3e` 集成提交。
- Rust 工作流 `30831516151`、Java 工作流 `30831516501`、Generated Packages
  工作流 `30831516914` 和 C++ 工作流 `30831515866` 全部成功。
- Maven Central 可下载 `io.github.lattice-hub:pole-specification:0.1.0-ALPHA.44`，
  JAR 包含 `SidecarSessionServiceGrpc`；npm `latest` 指向 `0.1.0-ALPHA.44`，全新
  npm 工程可加载 26 个 Proto 和 `SidecarSessionService`。
- PyPI wheel 与 sdist `0.1.0a44` 可下载，全新虚拟环境可导入 Message 和 Stub；
  crates.io `0.1.0-ALPHA.44` 与 Go module tag 均可解析。
- GitHub Release 包含 C++ `.tar.gz`、`.zip` 和 `SHA256SUMS`，校验和通过；归档内
  版本为 `0.1.0-ALPHA.44`，包含 26 个权威 Proto 和 104 个生成源码文件。

## 2026-08-06 Thin SDK 控制会话 v3

- [x] 保留兼容的 `OpenSession` server stream
- [x] 增加 `OpenControlSession` 双向控制流
- [x] 定义本地服务注册、注销与状态事件
- [x] 明确 listener 仅代表本地出站端口
- [x] 扩展 C++ 与 C# Thin SDK 兼容矩阵
- [x] 更新 Go、Rust 生成代码与跨语言契约测试
- [x] 验证 Go、Rust、Node.js、Python、Java、C#
- [x] 修正 C++ 生成 API 测试并触发固定工具链复验

### Review

- 控制会话首个客户端事件必须为 `ClientHello`，Sidecar 首个服务端事件仍为完整
  `ListenerSnapshot`；后续客户端可声明或撤销本地业务服务。
- `LocalServiceRegistration.local_port` 是 Pod 内业务进程端口，不是注册中心发布
  端口；注册中心端点由 Sidecar 映射为 Pod IP 与对应协议 ingress listener。
- 本地已通过 Go、Rust、Node.js、Python、Java Docker Maven 与 C# 测试；本机
  Protobuf 35 与仓库锁定的 Protobuf 26.1 不兼容，C++ 留给 PR CI 固定工具链验证。
- 首轮 C++ CI 已成功生成和编译新 Stub，但历史测试仍断言 service 只有一个方法；
  已改为校验两个方法以及 `OpenControlSession` 的双向流属性并触发复验。
- 用户授权发布后发现 `v0.1.0-ALPHA.47` 已指向更早的提交；遵循不可移动已推送
  tag 的发布规则，改用 `v0.1.0-ALPHA.48`，tag 指向已验证的合并提交
  `2642bc29c0a512f4da84ec4eb862b1e1ceee9833`。
- GitHub prerelease `v0.1.0-ALPHA.48` 已发布；Rust `31039309898`、Java
  `31039309969`、C++ `31039309990`、Generated Packages `31039310044` 与 C#
  `31039310045` 工作流全部成功。
- Go module、crates.io、npm、PyPI、Maven Central、NuGet 与 GitHub Release C++
  归档均完成外部下载验证；npm/Python 还验证了 `OpenControlSession` 为双向流。

## 2026-10-02 配置中心统一 v1 契约

- [x] 定义统一 ConfigDefinition/ConfigRelease/ResolvedConfig，九个管理 HTTP 动作、HTTP resolve/watch 与唯一客户端 Discover 读取流
- [x] 固定 CAS、多灰度、服务端渲染和观察游标语义
- [x] 生成 Go/Rust 类型与客户端服务 Stub，保留已有 v1 并行改动
- [x] 增加有界 StateResponse、历史元数据分页与按 ID 单版本读取
- [x] 补齐原生 HTTP/gRPC 映射、权限边界及初版容量限制
- [x] 完成契约测试与定向交叉审查

### Review

- 新合同见 [配置中心统一 v1](../../CONFIG.md) 与 [权威 Proto](../../api/v1/config/config.proto)。
  发布冻结完整定义，运行时仅返回最终字节，原生 SDK 不取得模板、默认 values 或规则。
- 管理动作只通过 HTTP；config.proto 不声明 service。客户端统一复用 v1.ConfigGRPC，仅保留
  Discover 双向流，直接使用 ResolveRequest/ResolvedConfig，一请求一完整响应且无主动 push；
  不保留独立 Resolve/Watch 或 Create/Update/Publish/Preview 管理 gRPC 方法及旧发现消息转换。
- GetState 的 page_size/page_token/release_id 复用同一管理 HTTP state 接口；StateResponse 的 members 只返回
  sequence DESC 的历史元数据页，active_release_ids 保留全量生效 ID，selected_release 按 ID
  单独返回完整历史，避免管理响应累积全部历史正文。所有写入响应返回默认第一页。
- token 绑定 key/incarnation/release_set_revision 与 before_sequence；集合变化冲突要求重查，
  纯草稿修改不使历史页 token 失效。文档固定定义总量 1 MiB、正文/渲染 512 KiB、页面元数据
  256 KiB、最多 128 个 active gray 和 HTTP 请求 4 MiB 等初版资源限制，不宣称三种外部协议兼容。
- Go 全量测试、统一配置生成包与契约测试的 go vet、Rust 全量 test/check、cargo fmt 检查、
  git diff --check 与 Proto 副本一致性检查通过。Go/Rust 共享非 UTF-8/uint64 上界和分页固定 wire 向量。
- 全量 go vet 的旧 v1 测试在 config_template_contract_test.go:232 存在 copylock 基线，未改动。
  本次验证证明生成协议与 wire 合同；control-plane 持久事务、分页投影和 HTTP 长轮询由其核心测试另行验收。
- 未执行提交、推送、发布或部署；Console、原生 SDK 接入及 Apollo/Nacos/Consul 适配不属于此 spec 阶段。


## 2026-10-02 配置 spec 收拢与生成一致性

- [x] 移除旧配置公开模型，统一服务端渲染语法与测试向量
- [x] 引入 ValueType、PreviewResponse、ReleaseSummary，明确 CopyReleaseToDraft
- [x] 补充管理 HTTP 幂等键合同，保留 Discover 逐次请求/响应语义
- [x] 修复 C# 导入与 Java 包名，统一 Pole 服务命名及生成产物
- [x] 运行跨语言生成、契约测试与格式检查，核对无关工作区改动

### Review

- 范围仅为 specification；control-plane/SDK 的新合同接入与幂等实现单独验收。
- 本次不保留旧配置模型兼容；限流协议只调整产品命名为 pole.metric.v2，不改变其独立协议版本。
- 移除 config_file.proto、对应 Go/Rust 产物及旧 wire 测试；配置只公开统一发布和服务端渲染合同。
- 新增 ValueType、PreviewResponse、ReleaseSummary；ConfigState.releases 是摘要页，
  CopyReleaseToDraft 支持复制 normal/gray 历史定义。管理 HTTP 写入要求 Idempotency-Key；
  该项目前是规范，持久幂等仍需 control-plane 实现。Discover 的路径和逐次响应语义保持不变。
- 模板规范和 12 个共享向量改用统一 PreviewRequest JSON；覆盖单规则覆盖、歧义、精确大整数、
  空正文及 literal 二进制。spec 测试验证消息/字节/摘要，不声称已经运行服务端引擎。
- Go 全量 test/vet、Rust 全量 test/check/fmt、Node 生成/编译/测试、Python 生成及 3 项测试、
  C# 实际生成/构建及 3 项测试通过。26 份 Proto 的描述符、Java 源码和 C++ 消息生成通过；
  Java/C++ 完整编译未执行，本机缺少 JDK/Maven 和 C++ gRPC 插件。
- Rust 生成输入增加 .proto 过滤及稳定排序；核对所有 Rust Proto 副本与 api 一致，
  auth Proto/Go 产物与现有 skill 契约测试的初始 SHA-256 保持不变。
- git diff --check 通过。知识库 lint 仍有既有 lessons.md 缺少 frontmatter 的 1 个错误和
  缺少根/tasks 索引的 2 个警告，本次 task 记录没有新增诊断。
- 此次含不兼容的配置类型/字段和限流 RPC package 调整，消费者需按新 spec 接入；
  未验证 control-plane、SDK 或第三方协议适配，不执行提交、推送、发布或部署。

## 2026-10-03 原生配置客户端统一 Discover

- [x] 删除 WatchRequest 与原生 HTTP resolve/watch，统一 DiscoverRequest
- [x] 明确 Discover 逐次请求/响应、不等待变化、不主动推送的语义
- [x] 同步 Go/Rust 生成产物，验证跨语言契约与工作区保护

### Review

- 本条以当前 SDK 本地渲染、公共 Response/Code 的配置合同为准；此前服务端渲染记录为历史。
- 原生客户端只通过 `/v1.ConfigGRPC/Discover` 读取；管理操作仍通过 HTTP。
- Apollo/Nacos/Consul 的监听和等待语义由可选适配层处理，不引入原生 Watch。
- Go 全量 test/vet、Rust 全量 test/check/fmt、Node.js 生成/编译/测试、Python 生成及
  3 项测试、C# 实际生成/构建及 3 项测试通过；生成服务仅公开 Discover。
- Rust 配置 Proto 副本与权威文件一致；auth、公共 Code 及已有 skill 契约测试的
  8 份文件 SHA-256 保持不变，git diff --check 通过。
- 范围仅为 specification；实现仓库接入、提交、推送、发布和部署不在本次范围。
