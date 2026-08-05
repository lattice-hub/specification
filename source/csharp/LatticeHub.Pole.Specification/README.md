# Pole Specification for C#

`LatticeHub.Pole.Specification` 是 Pole Specification 的官方 C# Protobuf Message 与
gRPC Client/Server Stub 包。项目通过 `Grpc.Tools` 在构建时直接从仓库权威 `api/`
目录生成代码，不维护第二份 Proto。

```xml
<PackageReference Include="LatticeHub.Pole.Specification" Version="0.1.0" />
```

Sidecar 启动会话示例类型位于 `Pole.Sidecar.V1` 命名空间：

```csharp
using Pole.Sidecar.V1;

var hello = new ClientHello
{
    SdkLanguage = "csharp",
    SdkVersion = "0.1.0"
};
hello.SupportedProtocols.Add(Protocol.Grpc);
```

仓库内验证：

```shell
dotnet test source/csharp/LatticeHub.Pole.Specification.Tests/LatticeHub.Pole.Specification.Tests.csproj
dotnet pack source/csharp/LatticeHub.Pole.Specification/LatticeHub.Pole.Specification.csproj
```
