#!/usr/bin/env python3
"""Prepare a vendored chardet checkout for the reduced CPython WASI runtime.

The runtime does not ship the zlib extension module and keeps only a small
subset of the standard library.  This script rewrites the model loader to use
an uncompressed model table file and applies the few source changes needed to
avoid modules the runtime cannot import.

Usage:
    ./prepare.py app/chardet 7.7.0
"""

from __future__ import annotations

import struct
import sys
import zlib
from pathlib import Path

MODELS_RAW_MAGIC = b"CMDR"

PARSE_MODELS_RAW = '''def parse_models_raw(
    data: bytes,
) -> tuple[dict[str, bytes], dict[str, float]]:
    """Parse the uncompressed models.raw format used by the WASI runtime.

    The header is identical to models.bin, but the tables are stored
    uncompressed so the runtime does not need the zlib extension module.

    :param data: Raw bytes of models.raw (must be non-empty).
    :returns: A ``(models, norms)`` tuple.
    :raises ValueError: If the data is corrupt or truncated.
    """
    try:
        if data[:4] != MODELS_RAW_MAGIC:
            msg = "corrupt models.raw: missing CMDR magic"
            raise ValueError(msg)

        offset = 4  # skip magic
        (num_models,) = _unpack_uint32(data, offset)
        offset += 4

        if num_models > 10_000:
            msg = f"corrupt models.raw: num_models={num_models} exceeds limit"
            raise ValueError(msg)

        names: list[str] = []
        norms: dict[str, float] = {}
        for _ in range(num_models):
            (name_len,) = _unpack_uint32(data, offset)
            offset += 4
            if name_len > 256:
                msg = f"corrupt models.raw: name_len={name_len} exceeds 256"
                raise ValueError(msg)
            name = data[offset : offset + name_len].decode("utf-8")
            offset += name_len
            (norm,) = _unpack_float64(data, offset)
            offset += 8
            names.append(name)
            norms[name] = norm

        tables = data[offset:]
        expected = num_models * 65536
        if len(tables) != expected:
            msg = (
                f"corrupt models.raw: table size {len(tables)} "
                f"!= expected {expected}"
            )
            raise ValueError(msg)

        models = {
            name: tables[i * 65536 : (i + 1) * 65536]
            for i, name in enumerate(names)
        }
    except (struct.error, UnicodeDecodeError) as e:
        msg = f"corrupt models.raw: {e}"
        raise ValueError(msg) from e

    return models, norms


'''


def fail(message: str) -> None:
    raise SystemExit(f"prepare.py: {message}")


def replace_once(text: str, old: str, new: str, path: Path) -> str:
    if old not in text:
        fail(f"expected text not found in {path}: {old[:80]!r}")
    return text.replace(old, new, 1)


def write_version(chardet_dir: Path, version: str) -> None:
    (chardet_dir / "_version.py").write_text(f'__version__ = "{version}"\n')


def generate_models_raw(chardet_dir: Path) -> None:
    models_dir = chardet_dir / "models"
    data = (models_dir / "models.bin").read_bytes()
    if data[:4] != b"CMD2":
        fail("models.bin is not in CMD2 format")

    count = struct.unpack_from(">I", data, 4)[0]
    offset = 8
    header = bytearray(MODELS_RAW_MAGIC) + data[4:8]
    for _ in range(count):
        name_len = struct.unpack_from(">I", data, offset)[0]
        size = 4 + name_len + 8
        header += data[offset : offset + size]
        offset += size

    tables = zlib.decompress(data[offset:])
    expected = count * 65536
    if len(tables) != expected:
        fail(f"models.bin table size {len(tables)} != expected {expected}")

    (models_dir / "models.raw").write_bytes(bytes(header) + tables)


def patch_models_init(chardet_dir: Path) -> None:
    path = chardet_dir / "models" / "__init__.py"
    text = path.read_text()

    if "import importlib.resources\n" in text:
        text = text.replace("import importlib.resources\n", "from pathlib import Path\n", 1)
    if "_MODELS_DIR = Path(__file__).parent\n" not in text:
        text = replace_once(
            text,
            "from chardet.registry import REGISTRY, lookup_encoding\n",
            "from chardet.registry import REGISTRY, lookup_encoding\n\n"
            "_MODELS_DIR = Path(__file__).parent\n",
            path,
        )

    text = replace_once(text, "    parse_models_bin,\n", "    parse_models_raw,\n", path)
    text = text.replace('importlib.resources.files("chardet.models")', "_MODELS_DIR")
    text = replace_once(
        text,
        '_MODELS_DIR.joinpath("models.bin")',
        '_MODELS_DIR.joinpath("models.raw")',
        path,
    )
    text = replace_once(text, "return parse_models_bin(data)", "return parse_models_raw(data)", path)
    text = text.replace('"chardet models.bin is empty', '"chardet models.raw is empty')
    text = text.replace(
        '"""Load and parse models.bin, returning (models, norms).',
        '"""Load and parse models.raw, returning (models, norms).',
    )

    path.write_text(text)


def patch_format(chardet_dir: Path) -> None:
    path = chardet_dir / "models" / "_format.py"
    text = path.read_text()

    text = replace_once(text, "import struct\nimport zlib\n", "import struct\n", path)
    if 'MODELS_RAW_MAGIC = b"CMDR"\n' not in text:
        text = replace_once(
            text,
            'MODELS_MAGIC = b"CMD2"\n',
            'MODELS_MAGIC = b"CMD2"\n'
            '#: models.raw magic: the uncompressed runtime model format used by WASI.\n'
            'MODELS_RAW_MAGIC = b"CMDR"\n',
            path,
        )

    text = replace_once(
        text,
        '''    :raises ValueError: If the data is corrupt or truncated.
    """
    try:
        if data[:4] != MODELS_MAGIC:
''',
        '''    :raises ValueError: If the data is corrupt or truncated.
    """
    import zlib

    try:
        if data[:4] != MODELS_MAGIC:
''',
        path,
    )
    if "def parse_models_raw(" not in text:
        text = replace_once(text, "def parse_rowmax_bin(\n", PARSE_MODELS_RAW + "def parse_rowmax_bin(\n", path)
    text = replace_once(
        text,
        "    models_blob = bytes(header) + zlib.compress(bytes(tables), 9)\n",
        "    import zlib\n\n"
        "    models_blob = bytes(header) + zlib.compress(bytes(tables), 9)\n",
        path,
    )

    path.write_text(text)


def patch_confusion(chardet_dir: Path) -> None:
    path = chardet_dir / "pipeline" / "confusion.py"
    text = path.read_text()

    if "import importlib.resources\n" in text:
        text = text.replace("import importlib.resources\n", "from pathlib import Path\n", 1)
    if '_MODELS_DIR = Path(__file__).parent.parent / "models"\n' not in text:
        text = replace_once(
            text,
            "from chardet.registry import lookup_encoding\n",
            "from chardet.registry import lookup_encoding\n\n"
            '_MODELS_DIR = Path(__file__).parent.parent / "models"\n',
            path,
        )
    text = text.replace('importlib.resources.files("chardet.models")', "_MODELS_DIR")

    path.write_text(text)


def remove_dataclass_slots(chardet_dir: Path) -> None:
    for path in chardet_dir.rglob("*.py"):
        text = path.read_text()
        patched = (
            text.replace(", slots=True", "")
            .replace("slots=True, ", "")
            .replace("(slots=True)", "()")
        )
        if patched != text:
            path.write_text(patched)


def main() -> None:
    if len(sys.argv) != 3:
        fail("usage: prepare.py <chardet-dir> <version>")
    chardet_dir = Path(sys.argv[1])
    version = sys.argv[2]
    if not (chardet_dir / "models" / "models.bin").is_file():
        fail(f"{chardet_dir} is not a chardet package")

    write_version(chardet_dir, version)
    generate_models_raw(chardet_dir)
    patch_models_init(chardet_dir)
    patch_format(chardet_dir)
    patch_confusion(chardet_dir)
    remove_dataclass_slots(chardet_dir)
    print(f"prepared chardet {version} for CPython WASI")


if __name__ == "__main__":
    main()
