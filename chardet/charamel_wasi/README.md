# Charamel on CPython WASI

This directory is the runtime layout mounted by `DetectEncodingByCharamelWasm`.

- `python.wasm`: CPython built for WASI.
- `lib/`: a reduced CPython standard library, mounted at `/lib`.
- `app/`: detector script plus the vendored `charamel` package, mounted at `/app`.

The `charamel` resources in `app/charamel/resources` are decompressed from the
upstream `*.gzip` files so the WASI interpreter does not need zlib support just
to load the model data.

The Go implementation embeds `python.wasm`, `lib/`, and `app/` via `go:embed` by
default. Environment variables are only needed when testing an external WASI
Python build:

```sh
CHARDET_CHARAMEL_PYTHON_WASM=/path/to/python.wasm \
CHARDET_CHARAMEL_WASI_ROOT=/path/to/wasi-root \
CHARDET_CHARAMEL_WASI_APP="$(pwd)/chardet/charamel_wasi/app" \
go test ./chardet -run TestDetectEncoding -count=1 -v
```
