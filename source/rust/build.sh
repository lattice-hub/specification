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

CURRENT_OS=$(uname -s | tr 'A-Z' 'a-z')
CURRENT_ARCH=$(uname -m | tr 'A-Z' 'a-z')

pushd ../..
workdir=$(pwd)
popd

model_dir=${workdir}/api/v1/model
service_manage_dir=${workdir}/api/v1/service_manage
traffic_manage_dir=${workdir}/api/v1/traffic_manage
ratelimiter_dir=${workdir}/api/v1/traffic_manage/ratelimiter
fault_tolerance_dir=${workdir}/api/v1/fault_tolerance
config_manage_dir=${workdir}/api/v1/config_manage
security_dir=${workdir}/api/v1/security
ai_dir=${workdir}/api/v1/ai
sidecar_dir=${workdir}/api/v1/sidecar

rust_root_dir=${workdir}/source/rust/pole-specification
protoc_dir=${workdir}/source/protoc/protoc-${CURRENT_OS}-${CURRENT_ARCH}

export PROTOC=${protoc_dir}/bin/protoc

cp ${model_dir}/*.proto ${rust_root_dir}/proto/
cp ${service_manage_dir}/*.proto ${rust_root_dir}/proto/
cp ${traffic_manage_dir}/*.proto ${rust_root_dir}/proto/
cp ${ratelimiter_dir}/*.proto ${rust_root_dir}/proto/
cp ${fault_tolerance_dir}/*.proto ${rust_root_dir}/proto/
cp ${config_manage_dir}/*.proto ${rust_root_dir}/proto/
cp ${security_dir}/*.proto ${rust_root_dir}/proto/
cp ${ai_dir}/*.proto ${rust_root_dir}/proto/
cp ${sidecar_dir}/*.proto ${rust_root_dir}/proto/

pushd ${rust_root_dir}
cargo build --release
popd
