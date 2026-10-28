# PyCharamel on CPython WASI

This directory is the detector app layout for the shared CPython WASI runtime.
The interpreter and reduced standard library live in `../python_wasi`:

- `../python_wasi/python.wasm`: CPython built for WASI.
- `../python_wasi/lib/`: a reduced CPython standard library, mounted at `/lib`.
- `app/`: detector script plus the vendored `charamel` package, mounted at `/app`.

`DetectEncodingByCharamelWasm` and `DetectEncodingByPyChardet` share the
runtime in `../python_wasi`; `pychardet` keeps its own app package under
`../pychardet_wasi/app`.

The `charamel` resources in `app/charamel/resources` are decompressed from the
upstream `*.gzip` files so the WASI interpreter does not need zlib support just
to load the model data.

The Go implementation embeds `python.wasm`, `lib/`, and `app/` via `go:embed` by
default. Python is not required at run time.

## Build

This detector is opt-in behind the `charamel_wazero` build tag:

```sh
go build -tags charamel_wazero .
```

To compile it together with the PyChardet detector:

```sh
go build -tags 'charamel_wazero pychardet' .
```

Without a tag, the default build does not include this detector.

## Refresh vendored files

Refresh the upstream `charamel` package:

```sh
./build.sh sync-charamel
```

Refresh the shared CPython WASI interpreter and stdlib:

```sh
./build.sh sync-cpython-wasi
```

## External runtime test

Environment variables are only needed when testing an external WASI Python
build:

```sh
CHARDET_CHARAMEL_PYTHON_WASM=/path/to/python.wasm \
CHARDET_CHARAMEL_WASI_ROOT=/path/to/wasi-root \
CHARDET_CHARAMEL_WASI_APP="$(pwd)/chardet/pycharamel_wasi/app" \
go test ./chardet -run TestDetectEncoding -count=1 -v
```
