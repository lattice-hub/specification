# 2026-08-04 限流运行时目标服务与 ALPHA.45

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
- 本次只准备契约变更，不创建 tag 或 GitHub Release；外部制品发布仍需单独授权。
