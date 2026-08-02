from __future__ import annotations

import argparse
import hashlib
import re
import shutil
import tarfile
import tempfile
import zipfile
from pathlib import Path


VERSION_PATTERN = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+(?:[.-][0-9A-Za-z.-]+)?$")


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True)
    parser.add_argument("--generated-dir", type=Path, required=True)
    parser.add_argument("--proto-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    return parser.parse_args()


def copy_directory(source: Path, destination: Path) -> None:
    shutil.copytree(source, destination)


def sha256(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> None:
    arguments = parse_arguments()
    version = arguments.version.removeprefix("v")
    if not VERSION_PATTERN.fullmatch(version):
        raise RuntimeError(f"invalid release version: {arguments.version}")
    if not (arguments.generated_dir / "proto-files.txt").is_file():
        raise RuntimeError("C++ generated source manifest is missing")

    package_directory = Path(__file__).resolve().parent.parent
    repository_directory = package_directory.parent.parent.parent
    arguments.output_dir.mkdir(parents=True, exist_ok=True)
    archive_root_name = f"pole-specification-cpp-{version}"

    with tempfile.TemporaryDirectory() as temporary_directory:
        archive_root = Path(temporary_directory) / archive_root_name
        archive_root.mkdir()
        copy_directory(arguments.generated_dir, archive_root / "generated")
        copy_directory(arguments.proto_dir, archive_root / "proto")
        (archive_root / "cmake").mkdir()
        shutil.copy2(
            package_directory / "cmake" / "PoleSpecificationConfig.cmake.in",
            archive_root / "cmake" / "PoleSpecificationConfig.cmake.in",
        )
        shutil.copy2(
            package_directory / "cmake" / "package-CMakeLists.txt",
            archive_root / "CMakeLists.txt",
        )
        shutil.copy2(package_directory / "README.md", archive_root / "README.md")
        for license_name in ("LICENSE-APACHE", "LICENSE-MIT"):
            shutil.copy2(
                repository_directory / "source" / "rust" / "pole-specification" / license_name,
                archive_root / license_name,
            )
        (archive_root / "VERSION").write_text(f"{version}\n", encoding="utf-8")

        tar_path = arguments.output_dir / f"{archive_root_name}.tar.gz"
        with tarfile.open(tar_path, "w:gz") as archive:
            archive.add(archive_root, arcname=archive_root.name)

        zip_path = arguments.output_dir / f"{archive_root_name}.zip"
        with zipfile.ZipFile(zip_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
            for path in sorted(archive_root.rglob("*")):
                if path.is_file():
                    archive.write(path, Path(archive_root.name) / path.relative_to(archive_root))

    archives = (tar_path, zip_path)
    (arguments.output_dir / "SHA256SUMS").write_text(
        "".join(f"{sha256(path)}  {path.name}\n" for path in archives),
        encoding="utf-8",
    )


if __name__ == "__main__":
    main()
