# Thin SDK 契约变更记录

## 1.0.0

- 定义 TargetEnvelope wire version `1`。
- 定义 namespace、service、protocol、group、service_version、method 和
  original_endpoint 字段。
- 定义 Unicode 规范化、控制字符拒绝和 endpoint 校验规则。
- 定义 Unicode scalar value、UTF-8 百分号 Header 值编码、重复值拒绝和内部
  Header 覆盖规则。
- 定义 Sidecar 接收端拒绝向量和精确发行版本组合的兼容证据模型。
- 增加跨语言一致性向量和兼容矩阵。

当前版本只表示协议定义完成；Sidecar 与 SDK 的兼容发行版仍为空。
