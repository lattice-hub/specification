from __future__ import annotations

import re
import sys
from pathlib import Path

from packaging.version import Version


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit("usage: set_version.py <release-tag>")
    raw_version = sys.argv[1].removeprefix("v")
    version = str(Version(raw_version))
    pyproject = Path(__file__).resolve().parent.parent / "pyproject.toml"
    content = pyproject.read_text(encoding="utf-8")
    updated, count = re.subn(
        r'(?m)^version = "[^"]+"$',
        f'version = "{version}"',
        content,
        count=1,
    )
    if count != 1:
        raise RuntimeError("project version was not updated")
    pyproject.write_text(updated, encoding="utf-8")
    print(version)


if __name__ == "__main__":
    main()
