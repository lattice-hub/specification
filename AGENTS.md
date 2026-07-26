# AGENTS.md（仓库级）

本文件只适用于 `pole-io/specification` 仓库，并补充用户目录下的全局规则。

## 发布流程

### 发布完成的定义

本仓库的“发版”必须同时满足以下条件：

1. 发布改动已通过 PR 合入 `develop`。
2. 合并提交上的 Go、Rust 和格式检查全部成功。
3. annotated tag 已创建并推送，且指向已验证的 `develop` 集成提交。
4. 对应 tag 的 GitHub Release 已发布。
5. `Release-Rust` 工作流已运行并成功结束。

仅创建或推送 Git tag 不算完成发布。`.github/workflows/rust-release.yaml` 监听
`release.published`，只有发布 GitHub Release 才会触发 Rust crate 发布。

### 标准步骤

1. 同步远端分支和 tags，确认目标 tag 尚不存在。
2. 确认待发布提交已包含所有预期历史，并且是 `origin/develop` 的已验证提交。
3. 运行：
   - `go test ./...`
   - `cargo fmt --all -- --check`
   - `cargo test --all`
   - `cargo check --all`
4. 创建 PR 合入 `develop`，等待 PR CI 和合并后的 `develop` CI 全部通过。
5. 在最终 `develop` 集成提交上创建 annotated tag，并推送该 tag。
6. 使用该 tag 发布 GitHub Release；ALPHA 版本标记为 prerelease。
7. 跟踪 `Release-Rust` 工作流到终态，检查 crate 发布步骤成功。
8. 最终交付中分别报告 PR、合并提交、tag、GitHub Release 和发布工作流链接。

### 安全约束

- 已推送的 tag 不得移动、覆盖或删除；需要修复时递增版本号。
- tag 必须指向集成提交，不能指向尚未合入的功能分支。
- 发布工作流失败时，不得把任务标记为完成；应先查看日志并修复根因。
- 不得在没有用户明确授权的情况下发布 GitHub Release，因为它会触发外部制品发布。
- 如果用户只要求创建 tag，必须明确说明 tag 不会触发制品发布，并询问是否继续发布
  GitHub Release。
