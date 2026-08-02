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

## 2026-08-03：npm OIDC 不负责发布后的 dist-tag 管理

- npm Trusted Publisher 的 OIDC 权限用于 `npm publish` 或 `npm stage publish`，
  `npm dist-tag` 仍要求维护者交互式 2FA；不能设计依赖发布后自动改 tag 的流程。
- 若项目当前只有预发布版本但希望默认安装得到最新版本，应在发布时直接更新
  `latest`，不要先发布到其它 tag 后再依赖 CI 修改 `latest`。npm 11
  对预发布版本要求显式传入 `--tag`，因此该策略必须使用
  `npm publish --tag latest`，不能依赖 `npm publish` 的默认值。

## 2026-08-03：多语言发布不能只按当前目录推断完整范围

- 用户要求检查 specification 的“每个语言”发布时，必须同时核对当前树、历史删除、
  Maven/npm/PyPI/crates Registry 和消费者依赖；当前树缺少工程不等于该语言不在目标范围。
- Java Specification 的正式坐标是 `io.github.lattice-hub:pole-specification`，生成
  package 使用 `io.github.latticehub.pole.specification.*`，二者不能混淆。

## 2026-08-03：Thin SDK 语言范围必须与 Specification 生成链路对齐

- 新增 C++ Thin SDK 时，Specification 必须同步从权威 `api/` 产出 C++
  Protobuf Message 和 gRPC Stub；不能让 Thin SDK 手工复制 Proto 或自行维护生成副本。
- C++ 没有本项目已配置的通用 Registry，因此先以带 CMake target 的 GitHub
  Release 源码包作为正式产物，并在 CI 中验证生成、编译和打包。
- 全量 C++ 编译会暴露其它语言没有暴露的生成器命名冲突；已发布的 RPC
  wire path 不能为了生成方便直接改名。若官方生成器无法处理，允许仅对 C++ API
  使用方法别名，但必须自动校验生成描述符仍使用原始 wire path，且发布包保留未修改
  的权威 Proto。
