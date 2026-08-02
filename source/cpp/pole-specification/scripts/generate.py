from __future__ import annotations

import argparse
import re
import shutil
import subprocess
import tempfile
from pathlib import Path


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
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--proto-dir", type=Path, required=True)
    parser.add_argument("--protoc", type=Path, required=True)
    parser.add_argument("--grpc-cpp-plugin", type=Path, required=True)
    parser.add_argument("--include-dir", type=Path, action="append", default=[])
    return parser.parse_args()


def main() -> None:
    arguments = parse_arguments()
    proto_files = sorted(arguments.api_dir.rglob("*.proto"))
    if not proto_files:
        raise RuntimeError(f"no Proto files found under {arguments.api_dir}")

    names: set[str] = set()
    for proto_file in proto_files:
        if proto_file.name in names:
            raise RuntimeError(f"duplicate Proto basename: {proto_file.name}")
        names.add(proto_file.name)

    shutil.rmtree(arguments.output_dir, ignore_errors=True)
    shutil.rmtree(arguments.proto_dir, ignore_errors=True)
    arguments.output_dir.mkdir(parents=True)
    arguments.proto_dir.mkdir(parents=True)

    for proto_file in proto_files:
        shutil.copy2(proto_file, arguments.proto_dir / proto_file.name)

    message_command = [
        str(arguments.protoc),
        f"--proto_path={arguments.proto_dir}",
        *[f"--proto_path={include_dir}" for include_dir in arguments.include_dir],
        f"--cpp_out={arguments.output_dir}",
        *sorted(names),
    ]
    subprocess.run(message_command, cwd=arguments.proto_dir, check=True)

    with tempfile.TemporaryDirectory() as temporary_directory:
        grpc_proto_dir = Path(temporary_directory) / "proto"
        shutil.copytree(arguments.proto_dir, grpc_proto_dir)

        for name, mapping in LEGACY_CPP_METHOD_MAPPINGS.items():
            proto_path = grpc_proto_dir / name
            source = proto_path.read_text(encoding="utf-8")
            pattern = re.compile(
                rf"(?m)^(\s*)rpc\s+{re.escape(mapping['wire_method'])}(\s*\()"
            )
            adapted_source, replacements = pattern.subn(
                rf"\1rpc {mapping['cpp_method']}\2", source
            )
            if replacements != 1:
                raise RuntimeError(
                    f"expected exactly one legacy RPC mapping in {name}, got {replacements}"
                )
            proto_path.write_text(adapted_source, encoding="utf-8")

        for name in sorted(names):
            grpc_command = [
                str(arguments.protoc),
                f"--proto_path={grpc_proto_dir}",
                *[
                    f"--proto_path={include_dir}"
                    for include_dir in arguments.include_dir
                ],
                f"--grpc_out={arguments.output_dir}",
                f"--plugin=protoc-gen-grpc={arguments.grpc_cpp_plugin}",
                name,
            ]
            subprocess.run(grpc_command, cwd=grpc_proto_dir, check=True)

    for name, mapping in LEGACY_CPP_METHOD_MAPPINGS.items():
        stem = Path(name).stem
        generated_source_path = arguments.output_dir / f"{stem}.grpc.pb.cc"
        generated_source = generated_source_path.read_text(encoding="utf-8")
        cpp_path = f"/{mapping['service']}/{mapping['cpp_method']}"
        wire_path = f"/{mapping['service']}/{mapping['wire_method']}"
        replacements = generated_source.count(cpp_path)
        generated_source = generated_source.replace(cpp_path, wire_path)
        if replacements == 0:
            raise RuntimeError(f"generated C++ wire path not found: {cpp_path}")
        generated_source_path.write_text(generated_source, encoding="utf-8")

    missing_outputs: list[str] = []
    for name in sorted(names):
        stem = Path(name).stem
        for suffix in (".pb.cc", ".pb.h", ".grpc.pb.cc", ".grpc.pb.h"):
            generated_file = arguments.output_dir / f"{stem}{suffix}"
            if not generated_file.is_file():
                missing_outputs.append(generated_file.name)
    if missing_outputs:
        raise RuntimeError(f"missing generated C++ files: {', '.join(missing_outputs)}")

    (arguments.output_dir / "proto-files.txt").write_text(
        "".join(f"{name}\n" for name in sorted(names)),
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()
