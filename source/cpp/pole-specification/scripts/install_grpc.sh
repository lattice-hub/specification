#!/usr/bin/env bash

set -euo pipefail

if [[ "$#" -ne 1 ]]; then
    echo "usage: $0 <install-prefix>" >&2
    exit 2
fi

readonly grpc_version="1.64.3"
readonly grpc_sha256="7866067e3b29f7bad0228a091b4325852b8fad5adda4775f97b819cc402d8f37"
readonly abseil_revision="4a2c63365eff8823a5221db86ef490e828306f9d"
readonly abseil_sha256="2926ae3b70cb9a4cd4f6bb73eac2f16b7c02fa709a87a32a89634eaecc3ac208"
readonly protobuf_revision="2434ef2adf0c74149b9d547ac5fb545a1ff8b6b5"
readonly protobuf_sha256="387478260190c540388839a3449c635a69708d92fc38ea6e2364b1196db90ea5"
readonly install_prefix="$1"
readonly parallel_level="${CMAKE_BUILD_PARALLEL_LEVEL:-2}"

if [[ -f "$install_prefix/lib/cmake/grpc/gRPCConfig.cmake" ]]; then
    exit 0
fi

readonly work_directory="$(mktemp -d)"
trap 'rm -rf "$work_directory"' EXIT
readonly grpc_archive="$work_directory/grpc.tar.gz"
readonly grpc_source_directory="$work_directory/grpc-$grpc_version"
readonly abseil_archive="$work_directory/abseil.tar.gz"
readonly abseil_source_directory="$work_directory/abseil-cpp-$abseil_revision"
readonly protobuf_archive="$work_directory/protobuf.tar.gz"
readonly protobuf_source_directory="$work_directory/protobuf-$protobuf_revision"

curl --fail --location --silent --show-error \
    "https://github.com/grpc/grpc/archive/refs/tags/v$grpc_version.tar.gz" \
    --output "$grpc_archive"
curl --fail --location --silent --show-error \
    "https://github.com/abseil/abseil-cpp/archive/$abseil_revision.tar.gz" \
    --output "$abseil_archive"
curl --fail --location --silent --show-error \
    "https://github.com/protocolbuffers/protobuf/archive/$protobuf_revision.tar.gz" \
    --output "$protobuf_archive"
printf '%s  %s\n' "$grpc_sha256" "$grpc_archive" | sha256sum --check
printf '%s  %s\n' "$abseil_sha256" "$abseil_archive" | sha256sum --check
printf '%s  %s\n' "$protobuf_sha256" "$protobuf_archive" | sha256sum --check
tar -xzf "$grpc_archive" -C "$work_directory"
tar -xzf "$abseil_archive" -C "$work_directory"
tar -xzf "$protobuf_archive" -C "$work_directory"

if [[ ! -f "$install_prefix/lib/cmake/absl/abslConfig.cmake" ]]; then
    cmake \
        -S "$abseil_source_directory" \
        -B "$work_directory/abseil-build" \
        -G Ninja \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_CXX_STANDARD=17 \
        -DCMAKE_INSTALL_PREFIX="$install_prefix" \
        -DABSL_ENABLE_INSTALL=ON \
        -DABSL_PROPAGATE_CXX_STD=ON
    cmake --build "$work_directory/abseil-build" --target install --parallel "$parallel_level"
fi

if [[ ! -f "$install_prefix/lib/cmake/protobuf/protobuf-config.cmake" ]]; then
    cmake \
        -S "$protobuf_source_directory" \
        -B "$work_directory/protobuf-build" \
        -G Ninja \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_CXX_STANDARD=17 \
        -DCMAKE_INSTALL_PREFIX="$install_prefix" \
        -DCMAKE_PREFIX_PATH="$install_prefix" \
        -Dprotobuf_ABSL_PROVIDER=package \
        -Dprotobuf_BUILD_SHARED_LIBS=OFF \
        -Dprotobuf_BUILD_TESTS=OFF
    cmake --build "$work_directory/protobuf-build" --target install --parallel "$parallel_level"
fi

cmake \
    -S "$grpc_source_directory" \
    -B "$work_directory/build" \
    -G Ninja \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="$install_prefix" \
    -DCMAKE_PREFIX_PATH="$install_prefix" \
    -DgRPC_BUILD_TESTS=OFF \
    -DgRPC_INSTALL=ON \
    -DgRPC_ABSL_PROVIDER=package \
    -DgRPC_CARES_PROVIDER=package \
    -DgRPC_PROTOBUF_PROVIDER=package \
    -DgRPC_RE2_PROVIDER=package \
    -DgRPC_SSL_PROVIDER=package \
    -DgRPC_ZLIB_PROVIDER=package
cmake --build "$work_directory/build" --target install --parallel "$parallel_level"
