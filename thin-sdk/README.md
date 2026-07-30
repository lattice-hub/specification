# Pole Thin SDK 契约

本目录是 Pole Thin SDK 与 Pole Sidecar 本地接入协议的单一事实源。

## 当前版本

- TargetEnvelope wire version：`1`
- 契约版本：`1.0.0`
- 状态：协议已定义，尚无完成端到端验证的 Sidecar 或 SDK 正式发行版

## 资产

- [TargetEnvelope v1 语义](target-envelope/v1/README.md)
- [TargetEnvelope v1 JSON Schema](target-envelope/v1/schema.json)
- [TargetEnvelope v1 一致性向量](target-envelope/v1/conformance.json)
- [兼容矩阵](compatibility.json)
- [兼容矩阵 Schema](compatibility.schema.json)
- [变更记录](CHANGELOG.md)

各语言 SDK 必须 vendoring 合入 `develop` 的 `schema.json`、`conformance.json`
和 `SHA256SUMS`，并在语言原生测试中执行全部向量。`contract/VERSION` 必须记录
契约版本和 specification 的完整 Git commit SHA，构成不可变来源定位。SDK 不应
通过 Git submodule 或网络在构建期读取本仓库。

契约 `1.0.0` 合入后使用 annotated Git tag `thin-sdk-contract-v1.0.0` 固定到
唯一 specification commit；不创建 GitHub Release，避免触发其它语言发布流程。

## 版本关系

TargetEnvelope wire version 与 SDK 包版本独立：

- wire version 只在线路语义不兼容时提升；
- Java、Node.js、Python、Go SDK 各自遵循独立 SemVer；
- v1 的字段和 Header 集合冻结；
- 新增、删除、重命名字段或改变既有语义必须引入新的 wire version；
- 契约 SemVer 用于文档、向量或约束的发布标识，wire version 用于线路协商。

`compatibility.json` 的每个 `verified_combinations` 条目必须同时记录 Sidecar
版本、SDK 语言与版本，以及 specification、Sidecar、SDK 的 commit 和 CI 证据。
