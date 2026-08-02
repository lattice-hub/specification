# pole-specification

Pole Specification 的官方 Python gRPC 定义包。包内包含权威 Proto 副本，以及
`grpcio-tools` 生成的 Message、类型声明、Client Stub 和 Servicer。

```python
from pole_specification.generated import bootstrap_pb2, bootstrap_pb2_grpc
```

该包由 specification GitHub Release 自动生成和发布，不接受手工修改生成目录。
