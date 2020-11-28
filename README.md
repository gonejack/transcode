# transcode [![GitHub license](https://img.shields.io/github/license/gonejack/hsize.svg?color=blue)](LICENSE)
Command line tool for translating text encoding

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
