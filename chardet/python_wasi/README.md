# Shared CPython WASI runtime

This directory contains the CPython WASI runtime shared by the Python-based
detectors:

- `python.wasm`: CPython built for WASI.
- `lib/`: a reduced CPython standard library, mounted at `/lib`.

`DetectEncodingByCharamelWasm` and `DetectEncodingByPyChardet` both embed this
runtime through `chardet/chardet.python.wasi.go`; each detector keeps its own
application code under `../pycharamel_wasi/app` and `../pychardet_wasi/app`.

The committed standard library is intentionally reduced. Refresh
`python.wasm` and `lib/` from a
[cpython-wasi-build](https://github.com/brettcannon/cpython-wasi-build/releases)
release with:

```sh
./build.sh sync-cpython-wasi
```

Overrides:

```sh
CPYTHON_WASI_VERSION=3.15.0b2 \
CPYTHON_WASI_SDK=33 \
./build.sh sync-cpython-wasi
```

You can also point it at an already downloaded archive or an extracted release:

```sh
CPYTHON_WASI_ZIP=/path/to/python-wasi.zip ./build.sh sync-cpython-wasi
CPYTHON_WASI_SOURCE=/path/to/extracted-release ./build.sh sync-cpython-wasi
```

`annotationlib.py` and `ast.py` are part of the reduced stdlib because Python
3.15's `typing.NamedTuple` support requires them. `dataclasses.py`,
`hashlib.py`, `__future__.py`, and `_py_warnings.py` are needed by chardet
7.x. Run the detector tests after replacing the runtime:

```sh
go test -tags charamel_wazero ./chardet -run TestDetectEncoding -count=1 -v
go test -tags pychardet ./chardet -run TestDetectEncoding -count=1 -v
```
