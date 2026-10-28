#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR/app"
CHARAMEL_DIR="$APP_DIR/charamel"
CHARAMEL_REPO="${CHARAMEL_REPO:-https://github.com/chomechome/charamel.git}"
CHARAMEL_REF="${CHARAMEL_REF:-master}"

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

sync-cpython-wasi delegates to ../python_wasi/build.sh because the CPython WASI
interpreter and reduced stdlib are shared by every Python detector:

  chardet/python_wasi/python.wasm
  chardet/python_wasi/lib

Notes:

  - app/detect.py is not overwritten; it is this project's WASI entrypoint.
  - upstream *.gzip model resources are decompressed into plain files so the
    WASI Python runtime does not need gzip/zlib for charamel model loading.
  - the vendored resources loader is patched for Python 3.15's Enum string
    behavior by using encoding.value for model filenames.

CPython WASI layout expected by the Go embed paths:

  chardet/python_wasi/python.wasm
  chardet/python_wasi/lib/...
  chardet/pycharamel_wasi/app/...

Download source for sync-cpython-wasi:

  https://github.com/brettcannon/cpython-wasi-build/releases

External runtime test example:

  CHARDET_CHARAMEL_PYTHON_WASM=/path/to/python.wasm \\
  CHARDET_CHARAMEL_WASI_ROOT=/path/to/wasi-root-with-lib \\
  CHARDET_CHARAMEL_WASI_APP="\$(pwd)/chardet/pycharamel_wasi/app" \\
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
    "$SCRIPT_DIR/../python_wasi/build.sh" sync-cpython-wasi
    ;;
  *)
    usage
    exit 2
    ;;
esac
