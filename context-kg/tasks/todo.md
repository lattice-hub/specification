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
