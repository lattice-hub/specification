# Java、Node.js、Python、C++ 与 C# gRPC 生成包

本仓库以 `api/` 中的 Proto 为唯一权威来源，在 CI 和 Release 阶段生成并发布：

| 语言 | Registry | 包名 | 生成方式 |
| --- | --- | --- | --- |
| Java | Maven Central | `io.github.lattice-hub:pole-specification` | `protobuf-maven-plugin` 与 `grpc-java` |
| Node.js | npm | `@lattice-hub/pole-specification` | `@grpc/proto-loader` 与 `proto-loader-gen-types` |
| Python | PyPI | `pole-specification` | `grpcio-tools` |
| C++ | GitHub Release | `pole-specification-cpp-<version>` | `protoc` 与 `grpc_cpp_plugin` |
| C# | NuGet | `LatticeHub.Pole.Specification` | `Grpc.Tools` |

生成文件不提交到 Git。普通分支和 Pull Request 由 `.github/workflows/testing.yml`
重新生成、测试和打包；发布 GitHub Release 后，
Java 由 `.github/workflows/java-release.yml` 发布，Node.js 与 Python 由
`.github/workflows/generated-release.yml` 发布，C++ 由
`.github/workflows/cpp-release.yml` 生成并附加 CMake 源码包，C# 由
`.github/workflows/csharp-release.yml` 发布到 NuGet，均使用 Release tag 设置包版本。
Python 生成器固定版本，且包声明的 `grpcio`、`protobuf` 最低版本与生成代码一致。

## 本地验证

```shell
docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace/source/java/pole-specification \
  maven:3.9.11-eclipse-temurin-17 \
  mvn --batch-mode --no-transfer-progress clean verify
```

```shell
cd source/node
npm ci
npm test
npm pack --dry-run
```

```shell
cd source/python
python -m pip install ".[dev]"
python scripts/generate.py
PYTHONPATH=src python -m unittest discover -s tests
python -m build
```

```shell
cmake -S source/cpp/pole-specification -B build/cpp -G Ninja
cmake --build build/cpp
ctest --test-dir build/cpp --output-on-failure
```

```shell
dotnet test source/csharp/LatticeHub.Pole.Specification.Tests/LatticeHub.Pole.Specification.Tests.csproj
dotnet pack source/csharp/LatticeHub.Pole.Specification/LatticeHub.Pole.Specification.csproj
```

## Registry 配置

发布工作流使用 GitHub Actions OIDC，不需要保存长期 npm 或 PyPI Token。

### Maven Central

Java 发布使用 GitHub `maven-central` Environment，并读取以下 Secrets：

- `MAVEN_CENTRAL_USERNAME`
- `MAVEN_CENTRAL_TOKEN`
- `MAVEN_GPG_PRIVATE_KEY`
- `MAVEN_GPG_PASSPHRASE`

Sonatype namespace 为 `io.github.lattice-hub`，发布坐标为
`io.github.lattice-hub:pole-specification`。

### npm

首次手工发布 `@lattice-hub/pole-specification` 后，在 npm 包设置中配置
Trusted Publisher：

- Organization：`lattice-hub`
- Repository：`specification`
- Workflow：`generated-release.yml`
- Environment：`npm`
- Allowed actions：`npm publish`

### PyPI

在 PyPI 创建 `pole-specification` 的 Pending Publisher，或在首次发布后配置
Trusted Publisher：

- Owner：`lattice-hub`
- Repository：`specification`
- Workflow：`generated-release.yml`
- Environment：`pypi`

GitHub 仓库需要同名的 `npm` 和 `pypi` Environments。可以按需为环境增加审批人；
不需要配置 `NPM_TOKEN` 或 `PYPI_API_TOKEN`。

### NuGet

C# 发布使用 GitHub `nuget` Environment，并读取 `NUGET_API_KEY` Secret。
API Key 需要允许向 NuGet.org 推送 `LatticeHub.Pole.Specification`。

## 版本映射

Node.js 使用去掉前导 `v` 的 Release tag，例如 `v0.1.0-ALPHA.41` 发布为
`0.1.0-ALPHA.41`。Python 按 PEP 440 规范化同一版本，发布为 `0.1.0a41`。
C++ 使用去掉前导 `v` 的版本生成 `.tar.gz` 和 `.zip` GitHub Release
附件，并同时发布 `SHA256SUMS`。C++ 生成和消费的最低 gRPC 版本为
1.64，最低 Protobuf 版本为 26.1（CMake package 版本 5.26.1）；CI 固定使用
gRPC 1.64.3 和 Protobuf 26.1。
C# 使用去掉前导 `v` 的 Release tag 作为 NuGet 包版本。
