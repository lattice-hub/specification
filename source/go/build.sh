#!/bin/bash
# Tencent is pleased to support the open source community by making Polaris available.
#
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
#
# Licensed under the BSD 3-Clause License (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# https://opensource.org/licenses/BSD-3-Clause
#
# Unless required by applicable law or agreed to in writing, software distributed
# under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
# CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.

set -e

# 统一配置定向生成入口：保留工作区已有 v1 生成文件及并行改动。
if [[ "${1:-}" == "config" ]]; then
    spec_root=$(cd "$(dirname "$0")/../.." && pwd)
    spec_protoc_dir="${spec_root}/source/protoc/protoc-$(uname -s | tr 'A-Z' 'a-z')-$(uname -m | tr 'A-Z' 'a-z')"
    mkdir -p "${spec_root}/source/go/api/v1/config" "${spec_root}/source/go/api/v1/config_manage"
    # 配置错误复用公共 Code，定向同步新增的配置码，不重建其他模型。
    "${spec_protoc_dir}/bin/protoc" \
        --plugin=protoc-gen-go="${spec_protoc_dir}/bin/protoc-gen-go" \
        --go_out=plugins=grpc,paths=source_relative:"${spec_root}/source/go/api/v1/model" \
        --proto_path="${spec_root}/api/v1/model" "${spec_root}/api/v1/model/code.proto"
    "${spec_protoc_dir}/bin/protoc" \
        --plugin=protoc-gen-go="${spec_protoc_dir}/bin/protoc-gen-go" \
        --go_out=plugins=grpc,paths=source_relative:"${spec_root}/source/go/api/v1/config" \
        --proto_path="${spec_root}/api/v1/config" "${spec_root}/api/v1/config/config.proto"
    "${spec_protoc_dir}/bin/protoc" \
        --plugin=protoc-gen-go="${spec_protoc_dir}/bin/protoc-gen-go" \
        --go_out=plugins=grpc,paths=source_relative:"${spec_root}/source/go/api/v1/config_manage" \
        --proto_path="${spec_root}/api/v1/config_manage" \
        --proto_path="${spec_root}/api/v1/config" \
        --proto_path="${spec_root}/api/v1/model" \
        --proto_path="${spec_protoc_dir}/include" \
        "${spec_root}/api/v1/config_manage/grpc_config_api.proto"
    exit 0
fi

if [[ $# -gt 0 ]]; then
    printf 'Usage: bash build.sh [config]\n' >&2
    exit 2
fi

CURRENT_OS=$(uname -s | tr 'A-Z' 'a-z')
CURRENT_ARCH=$(uname -m | tr 'A-Z' 'a-z')

pushd ../..
workdir=$(pwd)
popd
model_dir=${workdir}/api/v1/model
service_manage_dir=${workdir}/api/v1/service_manage
traffic_manage_dir=${workdir}/api/v1/traffic_manage
fault_tolerance_dir=${workdir}/api/v1/fault_tolerance
config_manage_dir=${workdir}/api/v1/config_manage
security_dir=${workdir}/api/v1/security
ai_dir=${workdir}/api/v1/ai
sidecar_dir=${workdir}/api/v1/sidecar
ratelimiter_dir=${workdir}/api/v1/traffic_manage/ratelimiter
out_dir=${workdir}/source/go

proto_files_model="types.proto response.proto namespace.proto code.proto"
proto_files_service_manage="client.proto service.proto grpcapi.proto heartbeat.proto contract.proto"
proto_files_traffic_manage="router.proto ratelimit.proto lane.proto lossless.proto mirror.proto mock.proto"
proto_files_fault_tolerance="circuitbreaker.proto fault_detector.proto"
proto_files_config_manage="grpc_config_api.proto"
proto_files_security="auth.proto traffic_security.proto workload_identity.proto"
proto_files_ai="mcp.proto"
proto_files_sidecar="bootstrap.proto"
proto_files_ratelimiter="ratelimiter.proto grpcapi_ratelimiter.proto"

if [[ "$CURRENT_OS" == "linux" || "$CURRENT_OS" == "darwin" ]]; then
    protoc_dir=${workdir}/source/protoc/protoc-${CURRENT_OS}-${CURRENT_ARCH}
    pushd "${protoc_dir}"/bin
    chmod +x *
    popd
    rm -rf "${out_dir}/api/v1"
    mkdir -p "${out_dir}/api/v1"
    # generate model
    pushd "${model_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path=. ${proto_files_model}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/model" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/model"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${fault_tolerance_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path="${model_dir}" \
        --proto_path=. ${proto_files_fault_tolerance}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/fault_tolerance" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/fault_tolerance"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${traffic_manage_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path="${model_dir}" \
        --proto_path=. ${proto_files_traffic_manage}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/traffic_manage" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/traffic_manage"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${config_manage_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path="${model_dir}" \
        --proto_path="${workdir}/api/v1/config" \
        --proto_path=. ${proto_files_config_manage}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/config_manage" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/config_manage"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${security_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path="${model_dir}" \
        --proto_path="${traffic_manage_dir}" \
        --proto_path=. ${proto_files_security}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/security" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/security"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${ai_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path=. ${proto_files_ai}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/ai" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/ai"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${sidecar_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path=. ${proto_files_sidecar}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/sidecar" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/sidecar"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${service_manage_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path=. \
        --proto_path="${model_dir}" \
        --proto_path="${traffic_manage_dir}" \
        --proto_path="${fault_tolerance_dir}" \
        --proto_path="${security_dir}" \
        --proto_path=. ${proto_files_service_manage}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/service_manage" "${out_dir}/api/v1"
    pushd "${out_dir}/api/v1/service_manage"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    pushd "${ratelimiter_dir}"
    "${protoc_dir}"/bin/protoc \
        --plugin=protoc-gen-go="${protoc_dir}"/bin/protoc-gen-go \
        --go_out=plugins=grpc:"${out_dir}" \
        --proto_path="${protoc_dir}"/include \
        --proto_path=. ${proto_files_ratelimiter}
    mv "${out_dir}/github.com/pole-io/specification/source/go/api/v1/traffic_manage/ratelimiter" "${out_dir}/api/v1/traffic_manage"
    pushd "${out_dir}/api/v1/traffic_manage/ratelimiter"
    "${protoc_dir}"/bin/protoc-go-inject-tag -input="*.pb.go"
    popd
    popd

    rm -rf "${out_dir}/github.com"
    bash "${out_dir}/build.sh" config
else
    docker run --rm -it -v "$(dirname $(pwd))":/app --workdir /app/source/go debian:buster bash build.sh
fi
