# transcode
Command line tool for translating text encoding

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/transcode)
![Build](https://github.com/gonejack/transcode/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/transcode.svg?color=blue)](LICENSE)

## Installation
```
go get -u github.com/gonejack/transcode
```

## Usage

By arguments:
```
> transcode source.txt
> transcode -s gbk -t utf8 source.txt
```

By stdin:
```
> cat source.txt | transcode
```
