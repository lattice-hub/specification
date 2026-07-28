# 协议兼容性

## 当前兼容基线

`v0.1.0-ALPHA.38` 以 `v0.1.0-ALPHA.37` 的 Protobuf wire 布局为兼容基线。
新增字段必须使用未占用的字段号；删除的字段名或字段号必须使用 `reserved` 保留。

`TrafficMirror` 和 `TrafficMock` 在 `v0.1.0-ALPHA.31` 引入了
`caller`/`callee` 模型，并调整了顶层字段布局。由于旧 `target_service = 4`
与新 `caller = 4` 具有相同的 wire 类型、但语义不同，schema 无法自动区分
`v0.1.0-ALPHA.30` 及更早版本的 payload。

因此：

- `.31` 及之后的 payload 可以在当前兼容基线上演进；
- `.30` 及更早的 Mirror/Mock payload 必须先通过应用层迁移；
- 未来若需要同时在线支持两套语义，应引入新 message 名称或带
  `schema_version` 的 envelope，不能再次复用字段号。

## 变更规则

- 同一字段号不得改变类型或业务语义。
- singular message 扩展为 repeated message 时，即使 wire 可以读取，也必须考虑
  JSON 字段名和生成代码 API 的兼容性。
- 生成代码必须由 `api/v1` 下的源 Proto 统一生成，不手工修改。

## 配置模板兼容扩展

配置模板能力保留既有 `ConfigFile`、`ConfigFileRelease`、
`ConfigFileReleaseHistory` 和 `ConfigFilePublishInfo` 的
`config_type = 200` wire 布局及 `CONFIG_FILE = 0`、`CONFIG_TEMPLATE = 1`
枚举值。

旧 `placeholder_value_map = 201` 保留并标记 deprecated；新实现使用新增的
`template_binding = 202` 和独立的模板、Namespace Value Release。发现响应使用新增的
`render_snapshot = 8`，因此旧 payload 仍然可读，旧客户端会忽略无法识别的新字段。

## Namespace 类型兼容扩展

`Namespace.kind = 10` 区分业务运行环境与 Pole 内部系统空间：

- `NAMESPACE_KIND_BUSINESS = 0` 是 wire 默认值；旧 payload 和旧 JSON 未携带
  `kind` 时，新客户端天然按业务空间解释。
- `NAMESPACE_KIND_SYSTEM = 1` 仅用于 Pole 控制面管理的内部系统空间。
- 旧客户端会忽略新增字段；服务端必须对历史 `pole-system` 记录显式返回
  `SYSTEM`，不能允许普通创建请求通过该字段伪造系统空间。
