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
