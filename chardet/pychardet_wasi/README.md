# PyChardet on CPython WASI

This directory is the runtime layout mounted by `DetectEncodingByPyChardet`.

- `app/`: `detect.py`, a minimal `logging` shim, and the vendored Python
  [`chardet`](https://github.com/chardet/chardet) package.
- `prepare.py`: adapts an upstream chardet checkout to the reduced CPython WASI
  runtime.
- The CPython WASI interpreter and reduced stdlib live in `../python_wasi` and
  are embedded by the shared Python WASI runtime.

The vendored chardet version defaults to **7.7.0** and keeps its upstream
0BSD license in `app/chardet/LICENSE`.

## Why `models.raw`?

Since 7.x, upstream chardet stores its bigram model tables in a
zlib-compressed `models.bin`. The reduced CPython WASI interpreter does not
provide the `zlib` extension module, so `prepare.py` writes an uncompressed
`models.raw` containing the original header followed by the raw 64 KiB model
tables. The vendored loader reads `models.raw`; `models.bin` is kept only for
the SHA-256 validation of the `rowmax.bin` and `langnorm.bin` artifacts.

`prepare.py` also applies a few source changes needed by the reduced stdlib:

- loads package data with `pathlib.Path(__file__)` instead of
  `importlib.resources`;
- removes `dataclass(slots=True)`, which would import `inspect`;
- moves the remaining `zlib` import into the build-time serialization
  functions, so the runtime never imports it.

`detect.py` additionally keeps the historical `gb2312` result: when chardet
7.x reports `GB18030`/`GBK` but the input decodes cleanly as GB2312, the
compatibility shim reports `gb2312`. Real GB18030 data keeps its wider
encoding name.

## Build and test

Build and test with the `pychardet` tag (`pychardet_wazero` is accepted as an
alias for consistency with `charamel_wazero`):

```sh
go test -tags pychardet ./chardet -run TestDetectEncoding -count=1 -v
go build -tags 'charamel_wazero pychardet' .
```

## Refresh the vendored package

```sh
./build.sh sync-chardet
```

This clones the pinned upstream ref, copies the package from `src/chardet`,
and runs `prepare.py`. It requires a host `python3` with `zlib` support. Use
`CHARDET_REF` / `CHARDET_VERSION` to select another release, or
`CHARDET_SOURCE` to use a local checkout:

```sh
CHARDET_REF=7.7.0 ./build.sh sync-chardet
CHARDET_SOURCE=/path/to/chardet ./build.sh sync-chardet
```

The interpreter and stdlib come from the shared runtime. Refresh those with:

```sh
./build.sh sync-cpython-wasi
```

The `logging.py` shim is kept for older chardet releases that emit DEBUG
diagnostics; the full logging package would pull in `traceback` and other
modules that the reduced stdlib does not contain.
