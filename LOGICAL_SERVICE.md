# 逻辑服务与环境服务规范

本文定义控制面逻辑服务与运行时环境服务的关系。其中 MUST、MUST NOT、SHOULD 按
RFC 2119 的规范性含义解释。

## 两层身份

- `LogicalService.id` 是控制面稳定身份，用于跨 Namespace 聚合产品信息。
- `Service.id` 与 `namespace + name` 仍是环境服务的运行时身份。
- SDK 注册、发现、契约和治理请求 MUST NOT 携带或依赖 `LogicalService.id`。
- 同一逻辑服务在不同 Namespace 中 MAY 使用不同的运行时服务名。

## 显式关联

- 管理端 MUST 通过 `ServiceEnvironmentBinding` 显式关联现有环境服务。
- 服务端 MUST 以 `service_id` 作为权威关联键，并从现有 `Service` 解析 Namespace
  与运行时服务名；不得信任客户端重复提交的名称快照。
- 一个环境服务 MUST NOT 同时属于多个逻辑服务。
- 一个逻辑服务在同一 Namespace MUST NOT 同时关联多个运行时服务。
- 有效逻辑服务名称 MUST 全局唯一；并发创建与改名也必须由持久化约束拒绝重复。
- 名称相同只能作为管理端候选建议，MUST NOT 自动建立关联。
- 关联冲突使用统一 `Code_DataConflict`；权限失败使用统一
  `Code_NotAllowedAccess`，不得用业务诊断字符串代替错误码。

## 生命周期

- 删除逻辑服务 MUST NOT 删除其环境服务。
- 存在环境关联时，服务端 SHOULD 拒绝删除逻辑服务，要求管理员先显式解除关联。
- 删除环境服务 MUST NOT 隐式删除逻辑服务；控制面 MAY 保留缺失绑定用于审计和修复。
- 未关联环境服务仍 MUST 保持完整的注册、发现和治理能力。

## 查询与权限

- 逻辑服务根和关联属于管理面能力；服务端 MUST 使用独立管理函数权限保护创建、修改、
  删除、查询、关联和解除关联，不能把环境服务的一般写权限等同于逻辑服务管理权限。
- 逻辑服务的环境数量、实例数量和健康数量 MUST 只统计调用方有权读取的环境服务。
- 环境列表 MUST 对每个现有 `Service` 继续执行原有 Service/Namespace 权限判断。
- 管理接口可以使用 `LogicalService`、`ServiceEnvironmentBinding`、
  `BindServiceEnvironmentRequest` 和 `UnbindServiceEnvironmentRequest`。
- 客户端 gRPC 发现接口和现有 `Service` 消息不得因逻辑服务能力增加新字段。
