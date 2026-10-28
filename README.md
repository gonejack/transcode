# transcode
This command line do text file encoding conversions.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/transcode)
![Build](https://github.com/gonejack/transcode/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/transcode.svg?color=blue)](LICENSE)

## Installation
```bash
> go install github.com/gonejack/transcode@latest
```

## Usage

By arguments:
```bash
> transcode source.txt
> transcode -s gbk -t utf8 source.txt
```

By stdin:
```bash
> cat source.txt | transcode
```

## Flags
```
Flags:
  -h, --help                      Show context-sensitive help.
  -s, --source-encoding="auto"    Set source encoding, default as auto-detection.
  -t, --target-encoding="utf8"    Set target encoding, default as utf8.
  -d, --detect-encoding           Detect encoding only.
  -w, --overwrite                 Overwrite source file.
  -l, --list-encodings            list supported encodings
      --about                     Show about.
```

## Optional detectors

Some detectors are opt-in build tags:

- `charamel` / `charamel_wazero`: the Python
  [charamel](https://github.com/chomechome/charamel) detector.
- `pychardet` / `pychardet_wazero`: the Python
  [chardet](https://github.com/chardet/chardet) detector.

The Python detectors share a CPython WASI runtime vendored in
`chardet/python_wasi`.

For example:

```bash
> go build -tags pychardet .
```

## Acknowledgements

Thanks to [charamel](https://github.com/chomechome/charamel) for the encoding
detection model, to [chardet](https://github.com/chardet/chardet) for the Python
encoding detector, and to
[cpython-wasi-build](https://github.com/brettcannon/cpython-wasi-build) for the
CPython WASI release used by the embedded detector runtime.
