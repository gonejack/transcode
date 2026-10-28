import sys

import chardet


def main() -> None:
    data = sys.stdin.buffer.read()
    result = chardet.detect(data) or {}
    sys.stdout.write(result.get("encoding") or "")


if __name__ == "__main__":
    main()
