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
