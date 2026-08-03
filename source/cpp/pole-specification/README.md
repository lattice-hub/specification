# Pole Specification C++

Pole Specification 的官方 C++ Protobuf Message 与 gRPC Stub 源码包。

GitHub Release 会附加以下资产：

- `pole-specification-cpp-<version>.tar.gz`
- `pole-specification-cpp-<version>.zip`
- `SHA256SUMS`

源码包已经包含从仓库 `api/` 生成的 `.pb.cc`、`.pb.h`、`.grpc.pb.cc` 与
`.grpc.pb.h`，使用方只需要提供 Protobuf 26.1（CMake package 版本 5.26.1）、
gRPC C++ 1.64 或更高版本和 CMake：

```cmake
find_package(PoleSpecification CONFIG REQUIRED)
target_link_libraries(your_target PRIVATE PoleSpecification::pole-specification)
```

`RateLimitGRPC.Service` 是已发布的历史 RPC 路径，名称与 gRPC C++ 默认生成的
`Service` 基类冲突。C++ API 因此将该方法暴露为
`RateLimitGRPC::Stub::Stream(...)`，但生成脚本会把方法描述符严格还原为
`/polaris.metric.v2.RateLimitGRPC/Service`。仓库测试同时校验 C++ 方法别名、原始
wire path 和发布包内未修改的权威 Proto，协议兼容性保持不变。

仓库内验证：

```shell
cmake -S source/cpp/pole-specification -B build/cpp -G Ninja
cmake --build build/cpp
ctest --test-dir build/cpp --output-on-failure
```

CI 使用仓库脚本构建并缓存固定的 gRPC C++ 1.64.3 工具链，
Abseil 和 Protobuf 使用该 gRPC tag 锁定的官方 revision：

```shell
bash source/cpp/pole-specification/scripts/install_grpc.sh build/cpp-deps
cmake -S source/cpp/pole-specification -B build/cpp \
  -DCMAKE_PREFIX_PATH="$PWD/build/cpp-deps"
```
