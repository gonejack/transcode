#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR/app"
CHARDET_DIR="$APP_DIR/chardet"
CHARDET_REPO="${CHARDET_REPO:-https://github.com/chardet/chardet.git}"
CHARDET_REF="${CHARDET_REF:-7.7.0}"
CHARDET_VERSION="${CHARDET_VERSION:-$CHARDET_REF}"

usage() {
  cat >&2 <<EOF
Usage:
  $0 help
  $0 sync-chardet
  $0 sync-cpython-wasi

sync-chardet downloads the upstream Python chardet package and refreshes:

  $CHARDET_DIR

It then runs prepare.py to generate an uncompressed models.raw (the WASI
runtime has no zlib module) and apply the small source patches needed by the
reduced CPython stdlib.  A host python3 with zlib support is required.

By default it clones:

  $CHARDET_REPO
  ref: $CHARDET_REF

Overrides:

  CHARDET_REPO=https://github.com/chardet/chardet.git
  CHARDET_REF=7.7.0
  CHARDET_VERSION=7.7.0
  CHARDET_SOURCE=/path/to/local/chardet-checkout

sync-cpython-wasi delegates to ../python_wasi/build.sh because the CPython
WASI interpreter and reduced stdlib are shared by every Python detector:

  chardet/python_wasi/python.wasm
  chardet/python_wasi/lib

The Go implementation embeds this app plus the shared runtime from
chardet/python_wasi.  It is compiled with the pychardet build tag
(pychardet_wazero is an alias):

  go test -tags pychardet ./chardet -run TestDetectEncoding -count=1 -v
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

sync_chardet() {
  local tmpdir source_dir package_dir
  if [[ -n "${CHARDET_SOURCE:-}" ]]; then
    source_dir="$CHARDET_SOURCE"
  else
    require_cmd git
    tmpdir="$(mktemp -d)"
    trap "rm -rf '$tmpdir'" EXIT
    git clone --depth 1 --branch "$CHARDET_REF" "$CHARDET_REPO" "$tmpdir/chardet"
    source_dir="$tmpdir/chardet"
  fi

  if [[ -d "$source_dir/src/chardet" ]]; then
    package_dir="$source_dir/src/chardet"
  elif [[ -d "$source_dir/chardet" ]]; then
    package_dir="$source_dir/chardet"
  else
    echo "invalid chardet source: $source_dir" >&2
    echo "expected: $source_dir/chardet or $source_dir/src/chardet" >&2
    exit 1
  fi

  require_cmd python3
  rm -rf "$CHARDET_DIR"
  cp -R "$package_dir" "$CHARDET_DIR"
  if [[ -f "$source_dir/LICENSE" ]]; then
    cp "$source_dir/LICENSE" "$CHARDET_DIR/LICENSE"
  fi

  find "$CHARDET_DIR" \( -name '.DS_Store' -o -name '__pycache__' -o -name '*.pyc' \) -prune -exec rm -rf {} +
  python3 "$SCRIPT_DIR/prepare.py" "$CHARDET_DIR" "$CHARDET_VERSION"
  find "$CHARDET_DIR" \( -name '.DS_Store' -o -name '__pycache__' -o -name '*.pyc' \) -prune -exec rm -rf {} +
  echo "synced chardet from $package_dir"
}

case "${1:-help}" in
  help|-h|--help)
    usage
    ;;
  sync-chardet)
    sync_chardet
    ;;
  sync-cpython-wasi)
    "$SCRIPT_DIR/../python_wasi/build.sh" sync-cpython-wasi
    ;;
  *)
    usage
    exit 2
    ;;
esac
