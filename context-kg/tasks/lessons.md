# 经验记录

## 2026-07-26：发 tag 不等于完成发布

- 在本仓库中，用户要求“发版”或“走 release”时，不能止于创建和推送 Git tag。
- 必须发布对应的 GitHub Release，因为 `.github/workflows/rust-release.yaml`
  只监听 `release.published`。
- 发布后必须跟踪 Release 工作流到成功终态；失败时直接从日志定位并修复。
- 若用户只明确要求创建 tag，交付时也要说明它不会触发制品发布，并确认是否需要
  GitHub Release，不能默认认为发布已经完成。

## 2026-08-02：Maven namespace 与 Java package 分离

- Maven Central 使用已验证的 `io.github.lattice-hub` 作为 `groupId`；Java package 不能照搬含连字符的 namespace。
- Pole Java Thin SDK 的公共 package 固定为 `io.github.latticehub.client`；Specification 生成类型使用独立的 `io.github.latticehub.pole.specification.*` 层级。

## 2026-08-03：区分 Specification 生成包与 Thin SDK 发布

- 在讨论 Node.js、Python 和 Maven 包发布时，如果上下文指向 spec 仓库，目标是
  `specification` 编译生成的 gRPC/Protobuf 语言产物，不是对应 Thin SDK。
- 配置 Registry、Trusted Publisher 或给出首次发布命令前，必须先确认目标仓库中
  已存在实际生成目录、包元数据、版本来源和发布工作流；不能用 Thin SDK 包代替。

## 2026-08-03：CodeGraph 索引属于本地资产

- `.codegraph/` 是本地代码索引，不是项目源码或发布产物；各仓库必须通过
  `.gitignore` 排除，不能暂存、提交或推送。
