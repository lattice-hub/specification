# Thin SDK 契约变更记录

## Unreleased

- 新增 TrafficContext v1：使用 W3C Baggage 保留成员传播 campaign、lane 和
  bucket；OTel-backed 与语言原生上下文运行时共享同一 wire 契约。
- 明确 `traceparent/tracestate`、TargetService 与 TrafficContext 的职责边界。

## 2.0.0

- 新增 Sidecar Session v1：Thin SDK 通过 gRPC over UDS 建立长连接，Sidecar
  在首帧主动下发协议 listener 端口表。
- 新增 TargetService v1，只保留 namespace 与 service。
- 元信息键改为 `latticehub-target-namespace` 与 `latticehub-target-service`。
- Thrift 统一使用 Apache Thrift 官方 HTTP Transport。
- 废弃 TargetEnvelope v1；保留旧资产用于已发布版本追溯。

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
