#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CPYTHON_WASI_VERSION="${CPYTHON_WASI_VERSION:-3.15.0b2}"
CPYTHON_WASI_SDK="${CPYTHON_WASI_SDK:-33}"
CPYTHON_WASI_ASSET="python-${CPYTHON_WASI_VERSION}-wasi_sdk-${CPYTHON_WASI_SDK}"
CPYTHON_WASI_URL="${CPYTHON_WASI_URL:-https://github.com/brettcannon/cpython-wasi-build/releases/download/v${CPYTHON_WASI_VERSION}/${CPYTHON_WASI_ASSET}.zip}"

PYTHON_STDLIB_FILES=(
  __future__.py
  _collections_abc.py
  _py_abc.py
  _py_warnings.py
  _weakrefset.py
  abc.py
  annotationlib.py
  ast.py
  codecs.py
  collections/__init__.py
  copyreg.py
  dataclasses.py
  enum.py
  functools.py
  genericpath.py
  hashlib.py
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
  $0 sync-cpython-wasi

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

  - this runtime is shared by chardet/pycharamel_wasi and
    chardet/pychardet_wasi.
  - the committed lib directory is intentionally reduced; run the detector
    tests after replacing it.
  - annotationlib.py and ast.py are required by Python 3.15's typing.NamedTuple
    support used by the Python detectors.

Download source for sync-cpython-wasi:

  https://github.com/brettcannon/cpython-wasi-build/releases

Required Homebrew tools for source builds:

  brew install cmake ninja wasmtime wabt wasi-sdk
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
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

case "${1:-help}" in
  help|-h|--help)
    usage
    ;;
  sync-cpython-wasi)
    sync_cpython_wasi
    ;;
  *)
    usage
    exit 2
    ;;
esac
