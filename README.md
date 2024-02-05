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
  -w, --overwrite                 Overwrite source file.
      --about                     Show about.
```
