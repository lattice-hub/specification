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
