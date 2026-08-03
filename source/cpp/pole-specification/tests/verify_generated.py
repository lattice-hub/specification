from __future__ import annotations

import argparse
import re
from pathlib import Path


SERVICE_PATTERN = re.compile(r"(?m)^\s*service\s+([A-Za-z][A-Za-z0-9_]*)\s*\{")
LEGACY_CPP_METHOD_MAPPINGS = {
    "grpcapi_ratelimiter.proto": {
        "service": "polaris.metric.v2.RateLimitGRPC",
        "wire_method": "Service",
        "cpp_method": "Stream",
    }
}


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--api-dir", type=Path, required=True)
    parser.add_argument("--generated-dir", type=Path, required=True)
    parser.add_argument("--proto-dir", type=Path, required=True)
    return parser.parse_args()


def main() -> None:
    arguments = parse_arguments()
    authoritative_files = sorted(arguments.api_dir.rglob("*.proto"))
    authoritative_names = [path.name for path in authoritative_files]
    if len(authoritative_names) != len(set(authoritative_names)):
        raise RuntimeError("authoritative Proto basenames must be unique")

    flattened_names = sorted(path.name for path in arguments.proto_dir.glob("*.proto"))
    if flattened_names != sorted(authoritative_names):
        raise RuntimeError("flattened C++ Proto set differs from authoritative api/")
    for authoritative_file in authoritative_files:
        flattened_file = arguments.proto_dir / authoritative_file.name
        if flattened_file.read_bytes() != authoritative_file.read_bytes():
            raise RuntimeError(
                f"flattened C++ Proto differs from authoritative source: "
                f"{authoritative_file.name}"
            )

    manifest_names = (arguments.generated_dir / "proto-files.txt").read_text(
        encoding="utf-8"
    ).splitlines()
    if manifest_names != sorted(authoritative_names):
        raise RuntimeError("generated C++ manifest differs from authoritative api/")

    for proto_file in authoritative_files:
        stem = proto_file.stem
        for suffix in (".pb.cc", ".pb.h", ".grpc.pb.cc", ".grpc.pb.h"):
            generated_file = arguments.generated_dir / f"{stem}{suffix}"
            if not generated_file.is_file():
                raise RuntimeError(f"missing generated C++ file: {generated_file.name}")

        grpc_header = (arguments.generated_dir / f"{stem}.grpc.pb.h").read_text(
            encoding="utf-8"
        )
        for service_name in SERVICE_PATTERN.findall(proto_file.read_text(encoding="utf-8")):
            if f"class {service_name} final" not in grpc_header:
                raise RuntimeError(
                    f"missing gRPC service {service_name} in {stem}.grpc.pb.h"
                )

    for name, mapping in LEGACY_CPP_METHOD_MAPPINGS.items():
        stem = Path(name).stem
        grpc_header = (arguments.generated_dir / f"{stem}.grpc.pb.h").read_text(
            encoding="utf-8"
        )
        grpc_source = (arguments.generated_dir / f"{stem}.grpc.pb.cc").read_text(
            encoding="utf-8"
        )
        cpp_method = mapping["cpp_method"]
        wire_method = mapping["wire_method"]
        service = mapping["service"]
        if not re.search(rf"\b{re.escape(cpp_method)}\s*\(", grpc_header):
            raise RuntimeError(f"missing C++ RPC alias {cpp_method} in {stem}.grpc.pb.h")
        wire_path = f"/{service}/{wire_method}"
        cpp_path = f"/{service}/{cpp_method}"
        if wire_path not in grpc_source:
            raise RuntimeError(f"missing preserved gRPC wire path: {wire_path}")
        if cpp_path in grpc_source:
            raise RuntimeError(f"generated C++ alias leaked into gRPC wire path: {cpp_path}")


if __name__ == "__main__":
    main()
