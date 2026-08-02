from __future__ import annotations

import re
import shutil
from pathlib import Path

import grpc_tools
from grpc_tools import protoc


PACKAGE_DIRECTORY = Path(__file__).resolve().parent.parent
REPOSITORY_DIRECTORY = PACKAGE_DIRECTORY.parent.parent
SOURCE_DIRECTORY = PACKAGE_DIRECTORY / "src" / "pole_specification"
PROTO_DIRECTORY = SOURCE_DIRECTORY / "proto"
GENERATED_DIRECTORY = SOURCE_DIRECTORY / "generated"


def main() -> None:
    shutil.rmtree(PROTO_DIRECTORY, ignore_errors=True)
    shutil.rmtree(GENERATED_DIRECTORY, ignore_errors=True)
    PROTO_DIRECTORY.mkdir(parents=True)
    GENERATED_DIRECTORY.mkdir(parents=True)

    proto_files = sorted((REPOSITORY_DIRECTORY / "api").rglob("*.proto"))
    names: set[str] = set()
    for proto_file in proto_files:
        if proto_file.name in names:
            raise RuntimeError(f"duplicate proto basename: {proto_file.name}")
        names.add(proto_file.name)
        shutil.copy2(proto_file, PROTO_DIRECTORY / proto_file.name)

    arguments = [
        "grpc_tools.protoc",
        f"-I{PROTO_DIRECTORY}",
        f"-I{Path(grpc_tools.__file__).resolve().parent / '_proto'}",
        f"--python_out={GENERATED_DIRECTORY}",
        f"--pyi_out={GENERATED_DIRECTORY}",
        f"--grpc_python_out={GENERATED_DIRECTORY}",
        *[str(PROTO_DIRECTORY / name) for name in sorted(names)],
    ]
    if protoc.main(arguments) != 0:
        raise RuntimeError("grpc_tools.protoc generation failed")

    import_pattern = re.compile(r"^import ([A-Za-z0-9_]+_pb2) as ", re.MULTILINE)
    for generated_file in GENERATED_DIRECTORY.glob("*_pb2*.py"):
        content = generated_file.read_text(encoding="utf-8")
        generated_file.write_text(
            import_pattern.sub(r"from . import \1 as ", content),
            encoding="utf-8",
        )

    (GENERATED_DIRECTORY / "__init__.py").write_text("", encoding="utf-8")
    for license_name in ("LICENSE-APACHE", "LICENSE-MIT"):
        shutil.copy2(
            REPOSITORY_DIRECTORY / "source" / "rust" / "pole-specification" / license_name,
            PACKAGE_DIRECTORY / license_name,
        )


if __name__ == "__main__":
    main()
