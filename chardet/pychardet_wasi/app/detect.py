import sys

import chardet


def _compat_encoding(data: bytes, encoding: str | None) -> str | None:
    """Keep the historical GB2312 result for GB2312-compatible input.

    Newer chardet releases report the GB18030 superset for data that older
    releases reported as GB2312.  The transcode API historically exposes
    GB2312 for that case, so only relabel when the bytes decode cleanly as
    GB2312; real GB18030 data keeps its wider encoding name.
    """
    if encoding and encoding.upper() in {"GB18030", "GBK"}:
        try:
            data.decode("gb2312")
        except UnicodeDecodeError:
            pass
        else:
            return "gb2312"
    return encoding


def main() -> None:
    data = sys.stdin.buffer.read()
    result = chardet.detect(data) or {}
    sys.stdout.write(_compat_encoding(data, result.get("encoding")) or "")


if __name__ == "__main__":
    main()
