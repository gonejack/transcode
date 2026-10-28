# PyChardet on CPython WASI

This directory is the runtime layout mounted by `DetectEncodingByPyChardet`.

- `app/`: `detect.py`, a minimal `logging` shim, and the vendored Python
  [`chardet`](https://github.com/chardet/chardet) package.
- The CPython WASI interpreter and reduced stdlib live in `../python_wasi` and
  are embedded by the shared Python WASI runtime.

The vendored chardet version defaults to 5.2.0 and keeps its upstream
LGPL-2.1 license in `app/chardet/LICENSE`.  Version 5.2.0 is used because the
newer 7.x releases require zlib to load their compressed models, which the
reduced CPython WASI runtime does not provide.

Build and test with the `pychardet` tag (`pychardet_wazero` is accepted as an
alias for consistency with `charamel_wazero`):

```sh
go test -tags pychardet ./chardet -run TestDetectEncoding -count=1 -v
```

Refresh the vendored package:

```sh
./build.sh sync-chardet
```

The interpreter and stdlib come from the shared runtime. Refresh those with:

```sh
./build.sh sync-cpython-wasi
```

The `logging.py` shim is intentional: chardet only uses DEBUG diagnostics, and
the full logging package would pull in `traceback` and other modules that the
reduced stdlib does not contain.
