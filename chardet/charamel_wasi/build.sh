#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR/app"
CHARAMEL_DIR="$APP_DIR/charamel"
CHARAMEL_REPO="${CHARAMEL_REPO:-https://github.com/chomechome/charamel.git}"
CHARAMEL_REF="${CHARAMEL_REF:-master}"
CPYTHON_WASI_VERSION="${CPYTHON_WASI_VERSION:-3.15.0b2}"
CPYTHON_WASI_SDK="${CPYTHON_WASI_SDK:-33}"
CPYTHON_WASI_ASSET="python-${CPYTHON_WASI_VERSION}-wasi_sdk-${CPYTHON_WASI_SDK}"
CPYTHON_WASI_URL="${CPYTHON_WASI_URL:-https://github.com/brettcannon/cpython-wasi-build/releases/download/v${CPYTHON_WASI_VERSION}/${CPYTHON_WASI_ASSET}.zip}"

PYTHON_STDLIB_FILES=(
  _collections_abc.py
  _py_abc.py
  _weakrefset.py
  abc.py
  codecs.py
  collections/__init__.py
  copyreg.py
  enum.py
  functools.py
  genericpath.py
  keyword.py
  operator.py
  os.py
  pathlib/__init__.py
  pathlib/_local.py
  pathlib/_os.py
  pathlib/types.py
  posixpath.py
  re/__init__.py
  re/_casefix.py
  re/_compiler.py
  re/_constants.py
  re/_parser.py
  reprlib.py
  stat.py
  struct.py
  types.py
  typing.py
  warnings.py
  weakref.py
)

usage() {
  cat >&2 <<EOF
Usage:
  $0 help
  $0 sync-charamel
  $0 sync-cpython-wasi

sync-charamel downloads the upstream charamel package and refreshes:

  $CHARAMEL_DIR

By default it clones:

  $CHARAMEL_REPO
  ref: $CHARAMEL_REF

Overrides:

  CHARAMEL_REPO=https://github.com/chomechome/charamel.git
  CHARAMEL_REF=master
  CHARAMEL_SOURCE=/path/to/local/charamel-checkout

sync-cpython-wasi downloads a CPython WASI release and refreshes:

  $SCRIPT_DIR/python.wasm
  $SCRIPT_DIR/lib

By default it downloads:

  $CPYTHON_WASI_URL

Overrides:

  CPYTHON_WASI_VERSION=3.15.0b2
  CPYTHON_WASI_SDK=33
  CPYTHON_WASI_URL=https://github.com/brettcannon/cpython-wasi-build/releases/download/...
  CPYTHON_WASI_ZIP=/path/to/already-downloaded.zip
  CPYTHON_WASI_SOURCE=/path/to/extracted/python-wasi-release

Notes:

  - app/detect.py is not overwritten; it is this project's WASI entrypoint.
  - upstream *.gzip model resources are decompressed into plain files so the
    WASI Python runtime does not need gzip/zlib for charamel model loading.
  - the vendored resources loader is patched for Python 3.15's Enum string
    behavior by using encoding.value for model filenames.

CPython WASI layout expected by the Go embed path:

  chardet/charamel_wasi/python.wasm
  chardet/charamel_wasi/lib/...
  chardet/charamel_wasi/app/...

If you downloaded a cpython-wasi-build release, keep python.wasm and trim/copy
the needed stdlib files under lib/. The committed lib directory is intentionally
reduced; run tests after replacing it.

Download source for sync-cpython-wasi:

  https://github.com/brettcannon/cpython-wasi-build/releases

Required Homebrew tools for source builds:

  brew install cmake ninja wasmtime wabt wasi-sdk

Use the requested uvx shim whenever a host Python command is needed:

  /Users/youi/.local/share/mise/shims/uvx python --version

External runtime test example:

  CHARDET_CHARAMEL_PYTHON_WASM=/path/to/python.wasm \\
  CHARDET_CHARAMEL_WASI_ROOT=/path/to/wasi-root-with-lib \\
  CHARDET_CHARAMEL_WASI_APP="\$(pwd)/chardet/charamel_wasi/app" \\
  go test ./chardet -run TestDetectEncoding -count=1 -v
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

sync_charamel() {
  require_cmd gzip

  local tmpdir source_dir
  if [[ -n "${CHARAMEL_SOURCE:-}" ]]; then
    source_dir="$CHARAMEL_SOURCE"
  else
    require_cmd git
    tmpdir="$(mktemp -d)"
    trap "rm -rf '$tmpdir'" EXIT
    git clone --depth 1 --branch "$CHARAMEL_REF" "$CHARAMEL_REPO" "$tmpdir/charamel"
    source_dir="$tmpdir/charamel"
  fi

  if [[ ! -d "$source_dir/charamel/resources/weights" ]]; then
    echo "invalid charamel source: $source_dir" >&2
    echo "expected: $source_dir/charamel/resources/weights" >&2
    exit 1
  fi

  rm -rf "$CHARAMEL_DIR"
  mkdir -p "$CHARAMEL_DIR/resources/weights"

  cp "$source_dir/charamel/__init__.py" "$CHARAMEL_DIR/"
  cp "$source_dir/charamel/detector.py" "$CHARAMEL_DIR/"
  cp "$source_dir/charamel/encoding.py" "$CHARAMEL_DIR/"
  if [[ -f "$source_dir/LICENSE" ]]; then
    cp "$source_dir/LICENSE" "$CHARAMEL_DIR/LICENSE"
  fi

  for file in "$source_dir"/charamel/resources/*.gzip; do
    gzip -dc "$file" >"$CHARAMEL_DIR/resources/$(basename "$file" .gzip)"
  done
  for file in "$source_dir"/charamel/resources/weights/*.gzip; do
    gzip -dc "$file" >"$CHARAMEL_DIR/resources/weights/$(basename "$file" .gzip)"
  done

  write_wasi_resource_loader

  find "$CHARAMEL_DIR" \( -name '.DS_Store' -o -name '__pycache__' -o -name '*.pyc' \) -prune -exec rm -rf {} +
  echo "synced charamel from $source_dir"
}

sync_cpython_wasi() {
  local tmpdir source_dir zip_file
  tmpdir="$(mktemp -d)"
  trap "rm -rf '$tmpdir'" EXIT

  if [[ -n "${CPYTHON_WASI_SOURCE:-}" ]]; then
    source_dir="$CPYTHON_WASI_SOURCE"
  else
    require_cmd unzip
    if [[ -n "${CPYTHON_WASI_ZIP:-}" ]]; then
      zip_file="$CPYTHON_WASI_ZIP"
    else
      zip_file="$tmpdir/${CPYTHON_WASI_ASSET}.zip"
      download_file "$CPYTHON_WASI_URL" "$zip_file"
    fi

    unzip -q "$zip_file" -d "$tmpdir/unpacked"
    source_dir="$(find "$tmpdir/unpacked" -maxdepth 2 -type f -name python.wasm -exec dirname {} \; | head -n 1)"
  fi

  if [[ -z "${source_dir:-}" || ! -f "$source_dir/python.wasm" || ! -d "$source_dir/lib" ]]; then
    echo "invalid CPython WASI source: ${source_dir:-<empty>}" >&2
    echo "expected python.wasm and lib/ under the same directory" >&2
    exit 1
  fi

  cp "$source_dir/python.wasm" "$SCRIPT_DIR/python.wasm"
  refresh_reduced_python_lib "$source_dir/lib"
  echo "synced CPython WASI from $source_dir"
}

download_file() {
  local url="$1"
  local output="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -L --fail --show-error "$url" -o "$output"
  elif command -v wget >/dev/null 2>&1; then
    wget -O "$output" "$url"
  else
    echo "missing required command: curl or wget" >&2
    exit 1
  fi
}

refresh_reduced_python_lib() {
  local source_lib="$1"
  local pyver source_py target_py item

  pyver="$(find "$source_lib" -maxdepth 1 -type d -name 'python3.*' -exec basename {} \; | sort | tail -n 1)"
  if [[ -z "$pyver" ]]; then
    echo "invalid CPython WASI lib: missing python3.* directory under $source_lib" >&2
    exit 1
  fi

  source_py="$source_lib/$pyver"
  rm -rf "$SCRIPT_DIR/lib"
  target_py="$SCRIPT_DIR/lib/$pyver"
  mkdir -p "$target_py"

  cp -R "$source_py/encodings" "$target_py/"
  for item in "${PYTHON_STDLIB_FILES[@]}"; do
    if [[ ! -e "$source_py/$item" ]]; then
      echo "missing stdlib file: $source_py/$item" >&2
      exit 1
    fi
    mkdir -p "$target_py/$(dirname "$item")"
    cp -R "$source_py/$item" "$target_py/$item"
  done

  mkdir -p "$target_py/wasm32-wasi"
  touch "$target_py/wasm32-wasi/.gitkeep"
  find "$SCRIPT_DIR/lib" \( -name '.DS_Store' -o -name '__pycache__' -o -name '*.pyc' \) -prune -exec rm -rf {} +
}

write_wasi_resource_loader() {
  cat >"$CHARAMEL_DIR/resources/__init__.py" <<'PY'
"""
Charamel WASI resource loader.

This copy reads decompressed model files instead of upstream *.gzip resources.
"""
import pathlib
import struct
from typing import Any, Dict, List, Sequence

from charamel.encoding import Encoding

RESOURCE_DIRECTORY = pathlib.Path(__file__).parent.absolute()
WEIGHT_DIRECTORY = RESOURCE_DIRECTORY / 'weights'


def _unpack(file: pathlib.Path, pattern: str) -> List[Any]:
    with open(file, 'rb') as data:
        return [values[0] for values in struct.iter_unpack(pattern, data.read())]


def load_features() -> Dict[int, int]:
    features = _unpack(RESOURCE_DIRECTORY / 'features', pattern='>H')
    return {feature: index for index, feature in enumerate(features)}


def load_biases(encodings: Sequence[Encoding]) -> Dict[Encoding, float]:
    biases = {}
    with open(RESOURCE_DIRECTORY / 'biases', 'rb') as data:
        for line in data:
            encoding, bias = line.decode().split()
            biases[encoding] = float(bias)

    return {encoding: biases[encoding] for encoding in encodings}


def load_weights(encodings: Sequence[Encoding]) -> Dict[Encoding, List[float]]:
    weights = {}
    for encoding in encodings:
        weights[encoding] = _unpack(WEIGHT_DIRECTORY / encoding.value, pattern='>e')
    return weights
PY
}

case "${1:-help}" in
  help|-h|--help)
    usage
    ;;
  sync-charamel)
    sync_charamel
    ;;
  sync-cpython-wasi)
    sync_cpython_wasi
    ;;
  *)
    usage
    exit 2
    ;;
esac
